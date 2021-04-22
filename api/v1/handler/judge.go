package handler

import (
	"encoding/hex"
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
	pkg "github.com/Judgoo/JudgeX/pkg"
	xUtils "github.com/Judgoo/JudgeX/utils"

	judger "github.com/Judgoo/Judger/entities"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
	"gopkg.in/yaml.v2"
)

type LanguageInfo struct {
	Language    languages.LanguageType
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

func generateJudgerYml(workPath string, data *entities.JudgePostData, languageInfo *LanguageInfo, testdataEntrys *TestDataEntrys) error {
	lang := languageInfo.Language
	langProfile := lang.Profile()
	var judger = judger.IJudger{
		Language: lang.String(),
		Build:    langProfile.Build,
		Run:      langProfile.Run,
		RunnerArgs: &judger.IRunnerArgs{
			CpuTime: int(data.TimeLimit),
			Memory:  int(data.MemoryLimit),
			Mco:     langProfile.Mco,
		},
		TestData: *testdataEntrys,
	}
	fileContent, err := yaml.Marshal(judger)
	if err != nil {
		return err
	}
	file := &File{
		Path:    filepath.Join(workPath, "judger.yml"),
		Content: fileContent,
	}

	return WriteFile(file)
}

func doJudge(c *fiber.Ctx, data *entities.JudgePostData, languageInfo *LanguageInfo) error {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	hashCh := make(chan string)
	go func(content string) {
		hasher := blake3.New()
		hasher.Write([]byte(content))
		hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
		hashCh <- hex.EncodeToString(hasher.Sum(nil))
	}(data.Code)
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
	generateJudgerYml(workPath, data, languageInfo, &testdataResult.Result)
	return c.JSON(struct {
		Data     string
		Language string
		Version  string
		Build    []string
		Run      string
	}{
		Data:     "Hello, World!",
		Language: languageInfo.Language.String(),
		Version:  languageInfo.VersionName,
		Build:    languageInfo.Language.Profile().Build,
		Run:      languageInfo.Language.Profile().Run,
	})
}

func JudgeLanguageByVersion(c *fiber.Ctx) error {
	languageString := utils.CopyString(c.Params("language"))
	language, err := languages.ParseLanguageType(languageString)
	if err != nil {
		return pkg.ApiAbortWithoutData(c, fiber.StatusBadRequest, err.Error())
	}

	version := c.Params("version", "")

	versionName, versionInfo, exists := language.GetVersion(version)
	if !exists {
		return pkg.ApiAbort(c, fiber.StatusBadRequest, ErrorLanguageVersionNotFound.Error(), fmt.Sprintf("version %s not found in language %s", version, languageString))
	}

	languageInfo := LanguageInfo{Language: language, VersionName: versionName, Version: versionInfo}

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
