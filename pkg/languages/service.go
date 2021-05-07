package languages

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Judgoo/JudgeX/logger"
	"github.com/Judgoo/JudgeX/pkg/entities"
	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	xUtils "github.com/Judgoo/JudgeX/utils"

	judger "github.com/Judgoo/Judger/entities"
)

type JudgerInfo struct {
	Language string
	Version  string
	Build    []string
	Run      string
}
type RunnerSuccessResult struct {
	Status         *judger.RunnerStatus `json:"status"`
	CpuTimeUsed    int                  `json:"cpu_time_used"`
	CpuTimeUsedUs  int                  `json:"cpu_time_used_us"`
	RealTimeUsed   int                  `json:"real_time_used"`
	RealTimeUsedUs int                  `json:"real_time_used_us"`
	MemoryUsed     int                  `json:"memory_used"`
}
type JudgeResponse struct {
	Status     judger.RunnerStatus    `json:"status"`
	StatusInfo string                 `json:"status_info"`
	Result     []*RunnerSuccessResult `json:"result,omitempty"`
	JudgerInfo *JudgerInfo            `json:"info,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Stdout     string                 `json:"stdout,omitempty"`
	Stderr     string                 `json:"stderr,omitempty"`
	ExitCode   int                    `json:"exit_code,omitempty"`
	Id         string                 `json:"id"`
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
	Judge(requestid string, data *entities.JudgePostData, li *LanguageInfo) (*JudgeResponse, error)
}

type service struct {
	ProfileMap *LanguageProfileMap
}

//go:embed languages_impl.yml
var LanguageData []byte

var profileMap = new(LanguageProfileMap)

func init() {
	var err = yaml.Unmarshal(LanguageData, profileMap)
	if err != nil {
		log.Fatalf("err when load languages: %v", err)
	}
}

func NewService() Service {
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

func getWorkspacePath(id string, requestid string) string {
	// 也许可以换成专业的文件系统来做这件事
	// 文件夹分层 b6eec00f2b9335ece97f7a8f8b2cfeb1 -> b6/ee/b6eec00f2b9335ece97f7a8f8b2cfeb1
	folder1 := requestid[:2]
	folder2 := requestid[2:4]
	prefix := requestid[:32]

	// TODO `JudgeWorkspace` 这个换成放在设置项中的可配置的
	workDir := filepath.Join(os.TempDir(), "JudgeWorkspace", folder1, folder2)
	// 这样构造是因为这个 id 是需要返回到用户的，之后我们可以通过这个 ID 找到本次判题究竟存在哪儿
	return path.Join(workDir, fmt.Sprintf("%s-%s", prefix, id))
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
		DirectRun:    false,
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

	stopCh := make(chan struct{})
	// 2 分钟后杀死进程
	go func() {
		t := time.After(2 * time.Minute)
		for {
			// Check if command is done
			select {
			case <-stopCh:
				fmt.Println("done cmd")
				t = nil
				return
			case <-t:
				fmt.Println("stop docker cmd")
				dockerCmd.Stop()
				return
			}
		}
	}()

	// Block waiting for command to exit, be stopped, or be killed
	finalStatus := <-statusChan
	close(stopCh)
	return &finalStatus
}

func judgeJudgerErrorResult(result *judger.NormalResult, response *JudgeResponse) {
	if result.Code == judger.CodeSuccess {
		return
	}
	response.Stdout = result.Stdout
	response.Stderr = result.Stderr
	// Judger 执行出错
	switch result.Code {
	case judger.CodeCompileError:
		response.Status = judger.COMPILE_ERROR
	case judger.CodeUserCodeRunnerRunError:
		response.Status = judger.RUNTIME_ERROR
	case judger.CodeInitLoggerError:
		fallthrough
	case judger.CodeNoInputDataError:
		// TODO 这里需要优化
		// 应该报错
		fallthrough
	default:
		response.Status = judger.SYSTEM_ERROR
	}
}

func (s *service) Judge(requestid string, data *entities.JudgePostData, languageInfo *LanguageInfo) (*JudgeResponse, error) {
	langProfile := s.GetLangProfile(languageInfo.Language)
	workPath := getWorkspacePath(data.ID, requestid)
	fmt.Println(workPath)
	file := &File{
		filepath.Join(workPath, langProfile.Filename),
		[]byte(data.Code),
	}
	err := WriteFile(file)
	if err != nil {
		return &JudgeResponse{}, err
	}
	testdataResult := processTestData(workPath, data)
	if testdataResult.Error != nil {
		return &JudgeResponse{}, testdataResult.Error
	}
	judgerResult, errG := generateJudgerYml(workPath, data, languageInfo, &testdataResult.Result, langProfile)
	if errG != nil {
		return &JudgeResponse{}, errG
	}
	cmdStatus := execJudger(judgerResult.DockerRunCmd, workPath)
	fmt.Printf("%#v", cmdStatus)
	if cmdStatus.Error != nil {
		return &JudgeResponse{}, cmdStatus.Error
	}
	if cmdStatus.Complete {
		response := &JudgeResponse{
			Id: requestid,
			JudgerInfo: &JudgerInfo{
				Language: languageInfo.Language.String(),
				Version:  languageInfo.VersionName,
				Build:    langProfile.Build,
				Run:      langProfile.Run,
			},
		}
		// 解析 Judger 的输出
		result := new(judger.NormalResult)
		stdout := strings.Join(cmdStatus.Stdout, "\n")
		fmt.Println(stdout)
		err2 := json.Unmarshal([]byte(stdout), &result)
		if err2 != nil {
			// Judger 没有输出一个有效 JSON
			// 说明 Judger 可能崩了
			logger.Sugar.Infow("judger output is not json format", "stdout", stdout)
			return &JudgeResponse{}, ErrorJudgerError
		}
		if cmdStatus.Exit != 0 {
			logger.Sugar.Infow("cmd exit non-zero")
			judgeJudgerErrorResult(result, response)
		}
		logger.Sugar.Info("parse result success", zap.Any("result", result))
		if result.Code == judger.CodeSuccess {
			var r = make([]*RunnerSuccessResult, 0)
			var status judger.RunnerStatus = judger.ACCEPTED
			for _, item := range result.RunnerResult {
				_status := item.Runner.Status
				if _status > judger.ACCEPTED {
					status = _status
				}
				r = append(r, &RunnerSuccessResult{
					Status:         &_status,
					CpuTimeUsed:    item.Runner.CpuTimeUsed,
					CpuTimeUsedUs:  item.Runner.CpuTimeUsedUs,
					RealTimeUsed:   item.Runner.RealTimeUsed,
					RealTimeUsedUs: item.Runner.RealTimeUsedUs,
					MemoryUsed:     item.Runner.MemoryUsed,
				})
			}
			response.Status = status
			response.Result = r
		} else {
			// 默认就是系统错误
			logger.Sugar.Infow("judger result code is not zero")
			judgeJudgerErrorResult(result, response)
			response.Message = fmt.Sprintf("JudgeX error: %s", result.Error)
		}

		response.StatusInfo = GetStatusInfo(response.Status)
		return response, nil
	} else {
		// golang 在执行命令的时候出了问题, maybe I/O problem
		return &JudgeResponse{}, errors.WithMessage(cmdStatus.Error, "JudgeX 内部出现错误")
	}
}
