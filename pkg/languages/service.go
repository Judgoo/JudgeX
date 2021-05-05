package languages

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Judgoo/JudgeX/logger"
	"github.com/Judgoo/JudgeX/pkg/api"
	"github.com/Judgoo/JudgeX/pkg/entities"
	"github.com/go-cmd/cmd"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
	"gopkg.in/yaml.v2"

	xUtils "github.com/Judgoo/JudgeX/utils"

	judger "github.com/Judgoo/Judger/entities"
)

type JudgeResponse struct {
	Language string
	Version  string
	Build    []string
	Run      string
	Result   *judger.NormalResult
}

type languageInfoDisplay struct {
	VersionName string `json:"version"`
	DisplayName string `json:"name"`
	Description string `json:"description"`
}

type languageInfoMap map[string][]languageInfoDisplay

type Service interface {
	GetLanguages() languageInfoMap
	GetLangProfile(*LanguageType) *LanguageProfile
	Judge(c *fiber.Ctx, data *entities.JudgePostData, lt *LanguageType, versionStr string) (*JudgeResponse, error)
}

type service struct {
	ProfileMap *LanguageProfileMap
}

//go:embed languages_impl.yml
var LanguageData []byte

func NewService() Service {
	var profileMap = new(LanguageProfileMap)
	var err = yaml.Unmarshal(LanguageData, profileMap)
	if err != nil {
		log.Fatalf("err when load languages: %v", err)
	}
	return &service{
		ProfileMap: profileMap,
	}
}

func (s *service) GetLanguages() languageInfoMap {
	var result = languageInfoMap{}
	for lang, vs := range VersionNameMap {
		result[lang.String()] = make([]languageInfoDisplay, 0)
		for _, versionName := range vs {
			versionInfo := VersionInfos[versionName]
			result[lang.String()] = append(result[lang.String()], languageInfoDisplay{
				versionName,
				fmt.Sprintf("%s(%s)", lang.String(), versionInfo.DisplayName),
				versionInfo.Description,
			})
		}
	}
	return result
}

func (s *service) GetLangProfile(lang *LanguageType) *LanguageProfile {
	return (*s.ProfileMap)[lang.String()]
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

func generateJudgerYml(workPath string, data *entities.JudgePostData, languageInfo *LanguageInfo, testdataEntrys *TestDataEntrys, langProfile *LanguageProfile) (*judger.IJudger, error) {
	lang := languageInfo.Language
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

func (s *service) Judge(c *fiber.Ctx, data *entities.JudgePostData, lt *LanguageType, versionStr string) (*JudgeResponse, error) {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())
	versionName, versionInfo, exists := lt.GetVersionInfo(versionStr)
	if !exists {
		return &JudgeResponse{}, ErrorLanguageVersionNotFound
	}
	languageInfo := LanguageInfo{Language: lt, VersionName: versionName, Version: versionInfo}

	langProfile := s.GetLangProfile(languageInfo.Language)

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
			filepath.Join(workPath, langProfile.Filename),
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
		return &JudgeResponse{}, api.ApiAbortWithoutData(c, 400, err.Error())

	}
	testdataResult := processTestData(workPath, data)
	if testdataResult.Error != nil {
		switch errors.Cause(testdataResult.Error) {
		case ErrorTestDataEmpty:
		case ErrorTestDataLengthNotEqual:
		default:
		}
		return &JudgeResponse{}, api.ApiAbortWithoutData(c, 400, testdataResult.Error.Error())
	}
	judgerResult, errG := generateJudgerYml(workPath, data, &languageInfo, &testdataResult.Result, langProfile)
	if errG != nil {
		return &JudgeResponse{}, api.ApiAbort(c, 400, "生成 judger.yml 时出错", testdataResult.Error.Error())
	}
	cmdStatus := execJudger(judgerResult.DockerRunCmd, workPath)
	if cmdStatus.Error != nil {
		return &JudgeResponse{}, api.ApiAbort(c, 400, "docker 运行出错", testdataResult.Error.Error())
	}
	if cmdStatus.Complete {
		// 解析 Judger 的输出
		result := new(judger.NormalResult)
		stdout := strings.Join(cmdStatus.Stdout, "\n")
		fmt.Println(stdout)
		err2 := json.Unmarshal([]byte(stdout), &result)
		if err2 != nil {
			// Judger 没有输出一个有效 JSON
			// 说明 Judger 可能崩了
			logger.Sugar.Infow("judger output is not json format", "stdout", stdout)
			return &JudgeResponse{}, api.ApiAbortWithoutData(c, 400, err.Error())
		}
		if cmdStatus.Exit != 0 {
			// Judger 执行出错
			switch result.Code {
			case judger.CodeCompileError:
				return &JudgeResponse{}, api.ApiAbort(c, 400, "编译错误", result)
			case judger.CodeRunnerRunError:
				return &JudgeResponse{}, api.ApiAbort(c, 400, "runner 执行用户代码报错", result)
			case judger.CodeInitLoggerError:
				return &JudgeResponse{}, api.ApiAbort(c, 400, "Judger logger 报错", result)
			case judger.CodeReadConfigFileError:
				return &JudgeResponse{}, api.ApiAbort(c, 400, "读取 judger.yml 出错", result)
			case judger.CodeUserCodeRunnerRunError:
				return &JudgeResponse{}, api.ApiAbort(c, 400, "用户代码执行出错(用户的问题)", result)
			default:
				return &JudgeResponse{}, api.ApiAbort(c, 400, "系统错误", result)
			}
		}
		// 执行成功
		return &JudgeResponse{
			Language: languageInfo.Language.String(),
			Version:  languageInfo.VersionName,
			Build:    langProfile.Build,
			Run:      langProfile.Run,
			Result:   result,
		}, nil
	} else {
		// golang 在执行命令的时候出了问题, maybe I/O problem
		return &JudgeResponse{}, api.ApiAbort(c, 400, "JudgeX 内部出现错误", cmdStatus.Error.Error())
	}
}
