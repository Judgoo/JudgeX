package v1

import (
	"JudgeX/api/v1/entities"
	"JudgeX/languages"
	pkg "JudgeX/pkg"
	xUtils "JudgeX/utils"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
)

type LanguageInfo struct {
	Language languages.LanguageType
	Version  string
}

type File struct {
	Path    string
	Content string
}

var CodeEmptyError = errors.New("code is empty")
var TestDataLengthNotEqual = errors.New("length of inputs and outputs are not equal")
var TestDataEmpty = errors.New("no testdata found")

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

func WriteFile(file File) error {
	err := os.MkdirAll(filepath.Dir(file.Path), os.ModeDir|(xUtils.OS_USER_RWX|xUtils.OS_ALL_R))
	if err != nil {
		return errors.Wrapf(err, "create directory %s fail", file.Path)
	}
	return os.WriteFile(file.Path, []byte(file.Content), (xUtils.OS_USER_RW | xUtils.OS_ALL_R))
}

func WriteCode(file File) error {
	if strings.TrimSpace(file.Content) == "" {
		return CodeEmptyError
	}

	return WriteFile(file)
}

type TestData = map[int][2]File

func WriteTestData(workPath string, data entities.JudgePostData) (TestData, error) {
	inputs := data.Inputs
	outputs := data.Outputs
	if len(inputs) != len(outputs) {
		return nil, TestDataLengthNotEqual
	}
	if len(inputs) == 0 {
		return nil, TestDataEmpty
	}
	testdata := make(TestData)
	for i := range inputs {
		in := File{
			path.Join(workPath, fmt.Sprintf("%v.in", i)),
			inputs[i],
		}
		out := File{
			path.Join(workPath, fmt.Sprintf("%v.out", i)),
			outputs[i],
		}
		WriteFile(in)
		WriteFile(out)
		testdata[i] = [2]File{in, out}
	}
	return testdata, nil
}

type TestDataResult struct {
	Result TestData
	Error  error
}

func getTestData(workPath string, data entities.JudgePostData) TestDataResult {
	tdCh := make(chan TestDataResult)

	go func() {
		testdata, err := WriteTestData(workPath, data)
		if err != nil {
			tdCh <- TestDataResult{nil, err}
		} else {
			tdCh <- TestDataResult{testdata, nil}
		}
	}()

	return <-tdCh
}

func doJudge(c *fiber.Ctx, data entities.JudgePostData, languageInfo LanguageInfo) error {
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
		codeCh <- WriteCode(File{
			filepath.Join(workPath, languageInfo.Language.Profile().Filename),
			data.Code,
		})
	}()
	err := <-codeCh
	if err != nil {
		switch errors.Cause(err) {
		case CodeEmptyError:
		default:
		}
		return pkg.ApiAbortWithoutData(c, 400, err.Error())

	}
	tdResult := getTestData(workPath, data)
	if tdResult.Error != nil {
		switch errors.Cause(tdResult.Error) {
		case TestDataEmpty:
		case TestDataLengthNotEqual:
		default:
		}
		return pkg.ApiAbortWithoutData(c, 400, tdResult.Error.Error())

	}
	return c.JSON(struct {
		Data     string
		Language string
		Version  string
		Build    []string
		Run      []string
	}{
		Data:     "Hello, World!",
		Language: languageInfo.Language.String(),
		Version:  languageInfo.Version,
		Build:    languageInfo.Language.Profile().Build,
		Run:      languageInfo.Language.Profile().Run,
	})
}

func judgeLanguageByVersion(c *fiber.Ctx) error {
	language := utils.CopyString(c.Params("language"))
	version := utils.CopyString(c.Params("version", "latest"))
	languageEnum, err := languages.ParseLanguageType(language)
	languageInfo := LanguageInfo{Language: languageEnum, Version: version}
	if err != nil {
		return pkg.ApiAbortWithoutData(c, fiber.StatusBadRequest, err.Error())
	}
	var requestBody entities.JudgePostData
	err = xUtils.ParseJSONBody(c, &requestBody)
	if err != nil {
		return pkg.ApiAbort(c, fiber.StatusBadRequest, "Parse JSON Body Error", err.Error())
	}
	validationErrors := entities.Validate(requestBody)
	if validationErrors != nil {
		return pkg.ApiAbort(c, fiber.StatusUnprocessableEntity, "Validation Error", validationErrors)
	}
	return doJudge(c, requestBody, languageInfo)
}
