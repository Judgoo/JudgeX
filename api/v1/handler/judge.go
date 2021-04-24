package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Judgoo/JudgeX/api/v1/entities"
	"github.com/Judgoo/JudgeX/languages"
	"github.com/Judgoo/JudgeX/logger"
	pkg "github.com/Judgoo/JudgeX/pkg"
	xUtils "github.com/Judgoo/JudgeX/utils"
	"github.com/go-cmd/cmd"

	judger "github.com/Judgoo/Judger/entities"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
	"gopkg.in/yaml.v2"
)

type LanguageInfo struct {
	Language    *languages.LanguageType
	VersionName string
	Version     *languages.VersionInfo
}

type File struct {
	Path    string
	Content []byte
}

var ErrorEmptyCode = errors.New("code is empty")
var ErrorTestDataLengthNotEqual = errors.New("length of inputs and outputs are not equal")
var ErrorTestDataEmpty = errors.New("no testdata found")
var ErrorLanguageVersionNotFound = errors.New("version not found")

func getWorkspacePath(id string, hash string) string {
	// 也许可以换成专业的文件系统来做这件事
	// 文件夹分层 b6eec00f2b9335ece97f7a8f8b2cfeb1 -> b6/ee/b6eec00f2b9335ece97f7a8f8b2cfeb1
	folder1 := hash[:2]
	folder2 := hash[2:4]
	prefix := hash[:32]

	// TODO `JudgeWorkspace` 这个换成放在设置项中的可配置的
	workDir := filepath.Join(os.TempDir(), "JudgeWorkspace", folder1, folder2)
	// 这样构造是因为这个 id 是需要返回到用户的，之后我们可以通过这个 ID 找到本次判题究竟存在哪儿
	folderName := fmt.Sprintf("%s-%s", prefix, id)
	return path.Join(workDir, folderName)
}

func WriteFile(file *File) error {
	err := os.MkdirAll(filepath.Dir(file.Path), os.ModeDir|(xUtils.OS_USER_RWX|xUtils.OS_ALL_R))
	if err != nil {
		return errors.Wrapf(err, "create directory %s fail", file.Path)
	}
	return os.WriteFile(file.Path, file.Content, (xUtils.OS_USER_RW | xUtils.OS_ALL_R))
}

type TestData = map[int][2]File
type TestDataEntrys = []string

func writeTestData(workPath string, data *entities.JudgePostData) (TestData, TestDataEntrys, error) {
	inputs := data.Inputs
	outputs := data.Outputs
	if len(inputs) != len(outputs) {
		return nil, nil, ErrorTestDataLengthNotEqual
	}
	if len(inputs) == 0 {
		return nil, nil, ErrorTestDataEmpty
	}
	testdata := make(TestData)
	testdataEntrys := make(TestDataEntrys, 0, len(inputs)+1)
	for i := range inputs {
		inS := fmt.Sprintf("%d.in", i)
		outS := fmt.Sprintf("%d.out", i)
		entry := fmt.Sprintf("%s::%s", inS, outS)
		testdataEntrys = append(testdataEntrys, entry)
		in := File{
			path.Join(workPath, inS),
			[]byte(inputs[i]),
		}
		out := File{
			path.Join(workPath, outS),
			[]byte(outputs[i]),
		}
		WriteFile(&in)
		WriteFile(&out)
		testdata[i] = [2]File{in, out}
	}
	return testdata, testdataEntrys, nil
}

type TestDataResult struct {
	Result TestDataEntrys
	Error  error
}

func processTestData(workPath string, data *entities.JudgePostData) TestDataResult {
	tdCh := make(chan TestDataResult)

	go func() {
		_, testdataEntrys, err := writeTestData(workPath, data)
		if err != nil {
			tdCh <- TestDataResult{nil, err}
		} else {
			tdCh <- TestDataResult{testdataEntrys, nil}
		}
	}()

	return <-tdCh
}

func generateJudgerYml(workPath string, data *entities.JudgePostData, languageInfo *LanguageInfo, testdataEntrys *TestDataEntrys) (*judger.IJudger, error) {
	lang := languageInfo.Language
	langProfile := lang.Profile()
	judgeCommand := fmt.Sprintf("docker run --rm -v %s:/workspace %s", workPath, languageInfo.Version.ImageName)
	var judgerStruct = judger.IJudger{
		Language: lang.String(),
		Build:    langProfile.Build,
		Run:      langProfile.Run,
		RunnerArgs: &judger.IRunnerArgs{
			CpuTime: int(data.TimeLimit),
			Memory:  int(data.MemoryLimit),
			Mco:     langProfile.Mco,
		},
		TestData:     *testdataEntrys,
		DockerRunCmd: judgeCommand,
	}
	fileContent, err := yaml.Marshal(judgerStruct)
	if err != nil {
		return new(judger.IJudger), err
	}
	file := &File{
		Path:    filepath.Join(workPath, "judger.yml"),
		Content: fileContent,
	}

	return &judgerStruct, WriteFile(file)
}

func execJudger(str string, dir string) *cmd.Status {
	target := strings.Split(str, " ")
	dockerCmd := cmd.NewCmd(target[0], target[1:]...)
	if dir != "" {
		dockerCmd.Dir = dir
	}
	statusChan := dockerCmd.Start() // non-blocking

	// 3 分钟后杀死进程
	go func() {
		<-time.After(3 * time.Minute)
		fmt.Printf("stop docker cmd")
		dockerCmd.Stop()
	}()

	// Block waiting for command to exit, be stopped, or be killed
	finalStatus := <-statusChan
	return &finalStatus
}

func doJudge(c *fiber.Ctx, data *entities.JudgePostData, languageInfo *LanguageInfo) error {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	hashCh := make(chan string)
	go func(content *string) {
		hasher := blake3.New()
		hasher.Write([]byte(*content))
		hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
		hashCh <- hex.EncodeToString(hasher.Sum(nil))
	}(&data.Code)
	codeHash := <-hashCh
	workPath := getWorkspacePath(data.ID, codeHash)
	fmt.Println(workPath)
	codeCh := make(chan error)
	go func() {
		if strings.TrimSpace(data.Code) == "" {
			codeCh <- ErrorEmptyCode
		}
		file := &File{
			filepath.Join(workPath, languageInfo.Language.Profile().Filename),
			[]byte(data.Code),
		}
		codeCh <- WriteFile(file)
	}()
	err := <-codeCh
	if err != nil {
		switch errors.Cause(err) {
		case ErrorEmptyCode:
		default:
		}
		return pkg.ApiAbortWithoutData(c, 400, err.Error())

	}
	testdataResult := processTestData(workPath, data)
	if testdataResult.Error != nil {
		switch errors.Cause(testdataResult.Error) {
		case ErrorTestDataEmpty:
		case ErrorTestDataLengthNotEqual:
		default:
		}
		return pkg.ApiAbortWithoutData(c, 400, testdataResult.Error.Error())
	}
	judgerResult, errG := generateJudgerYml(workPath, data, languageInfo, &testdataResult.Result)
	if errG != nil {
		return pkg.ApiAbort(c, 400, "生成 judger.yml 时出错", testdataResult.Error.Error())
	}
	cmdStatus := execJudger(judgerResult.DockerRunCmd, workPath)
	if cmdStatus.Error != nil {
		return pkg.ApiAbort(c, 400, "docker 运行出错", testdataResult.Error.Error())
	}
	if cmdStatus.Complete {
		// 解析 Judger 的输出
		result := new(judger.NormalResult)
		stdout := strings.Join(cmdStatus.Stdout, "\n")
		fmt.Printf(stdout)
		err2 := json.Unmarshal([]byte(stdout), &result)
		if err2 != nil {
			// Judger 没有输出一个有效 JSON
			// 说明 Judger 可能崩了
			logger.Sugar.Infow("judger output is not json format", "stdout", stdout)
			return pkg.ApiAbortWithoutData(c, 400, err.Error())
		}
		if cmdStatus.Exit != 0 {
			// Judger 执行出错
			switch result.Code {
			case judger.CodeCompileError:
				return pkg.ApiAbort(c, 400, "编译错误", result)
			case judger.CodeRunnerRunError:
				return pkg.ApiAbort(c, 400, "runner 执行用户代码报错", result)
			case judger.CodeInitLoggerError:
				return pkg.ApiAbort(c, 400, "Judger logger 报错", result)
			case judger.CodeReadConfigFileError:
				return pkg.ApiAbort(c, 400, "读取 judger.yml 出错", result)
			case judger.CodeUserCodeRunnerRunError:
				return pkg.ApiAbort(c, 400, "用户代码执行出错(用户的问题)", result)
			default:
				return pkg.ApiAbort(c, 400, "系统错误", result)
			}
		} else {
			// 执行成功
			return c.JSON(struct {
				Language string
				Version  string
				Build    []string
				Run      string
				Result   *judger.NormalResult
			}{
				Language: languageInfo.Language.String(),
				Version:  languageInfo.VersionName,
				Build:    languageInfo.Language.Profile().Build,
				Run:      languageInfo.Language.Profile().Run,
				Result:   result,
			})
		}
	} else {
		// golang 在执行命令的时候出了问题, maybe I/O problem
		return pkg.ApiAbort(c, 400, "JudgeX 内部出现错误", cmdStatus.Error.Error())
	}
}

func JudgeLanguageByVersion(c *fiber.Ctx) error {
	languageString := utils.CopyString(c.Params("language"))
	language, err := languages.ParseLanguageType(languageString)
	if err != nil {
		return pkg.ApiAbortWithoutData(c, fiber.StatusBadRequest, err.Error())
	}

	version := c.Params("version", "")

	versionName, versionInfo, exists := language.GetVersionInfo(version)
	if !exists {
		return pkg.ApiAbort(c, fiber.StatusBadRequest, ErrorLanguageVersionNotFound.Error(), fmt.Sprintf("version %s not found in language %s", version, languageString))
	}

	languageInfo := LanguageInfo{Language: &language, VersionName: versionName, Version: versionInfo}

	var requestBody entities.JudgePostData
	err = xUtils.ParseJSONBody(c, &requestBody)
	if err != nil {
		return pkg.ApiAbort(c, fiber.StatusBadRequest, "Parse JSON Body Error", err.Error())
	}
	validationErrors := entities.Validate(requestBody)
	if validationErrors != nil {
		return pkg.ApiAbort(c, fiber.StatusUnprocessableEntity, "Validation Error", validationErrors)
	}
	return doJudge(c, &requestBody, &languageInfo)
}
