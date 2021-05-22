package judge

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Judgoo/JudgeX/logger"
	"github.com/Judgoo/JudgeX/pkg/entities"
	"github.com/Judgoo/JudgeX/utils"
	"github.com/Judgoo/languages"
	"github.com/go-cmd/cmd"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

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
	Judge(requestid string, data *entities.JudgePostData, li *languages.LanguageInfo) (*JudgeResponse, error)
}

type service struct {
}

func NewService() Service {
	return &service{}
}

func (s *service) GetLanguages() languageInfoMap {
	var result = languageInfoMap{}
	for lang, vs := range languages.VersionNameMap {
		result[lang.String()] = make([]languageInfoDisplay, 0)
		for _, versionName := range vs {
			versionInfo := languages.VersionInfos[versionName]
			result[lang.String()] = append(result[lang.String()], languageInfoDisplay{
				versionName,
				fmt.Sprintf("%s(%s)", lang.String(), versionInfo.DisplayName),
				versionInfo.Description,
			})
		}
	}
	return result
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

type TestData = map[int][2]utils.File
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
		in := utils.File{
			Path:    path.Join(workPath, inS),
			Content: []byte(inputs[i]),
		}
		out := utils.File{
			Path:    path.Join(workPath, outS),
			Content: []byte(outputs[i]),
		}
		utils.WriteFile(&in)
		utils.WriteFile(&out)
		testdata[i] = [2]utils.File{in, out}
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

func generateJudgerYml(workPath string, data *entities.JudgePostData, languageInfo *languages.LanguageInfo, testdataEntrys *TestDataEntrys, langProfile *languages.LanguageProfile) (*judger.IJudger, error) {
	lang := languageInfo.Language
	capsToDrop := [...]string{"MKNOD", "NET_RAW", "NET_BIND_SERVICE"}
	var capsToDropString string
	for _, ct := range capsToDrop {
		capsToDropString += fmt.Sprintf("--cap-drop %s ", ct)
	}
	args := fmt.Sprintf("--network none --cpus 1 -m 100m %s --rm -v %s:/workspace", strings.TrimSpace(capsToDropString), workPath)
	judgeCommand := fmt.Sprintf("docker run %s %s", args, languageInfo.Version.ImageName)
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
	file := &utils.File{
		Path:    filepath.Join(workPath, "judger.yml"),
		Content: fileContent,
	}

	return &judgerStruct, utils.WriteFile(file)
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

type ExecJudgerArg struct {
	WorkPath string
	Cmd      string
	Ch       chan *cmd.Status
}

var p *ants.PoolWithFunc

func init() {
	p, _ = ants.NewPoolWithFunc(6, func(i interface{}) {
		c := i.(*ExecJudgerArg)
		fmt.Print("in Goroutine Pool\n")
		fmt.Printf("WorkPath %s\n", c.WorkPath)
		fmt.Printf("Cmd %s\n", c.Cmd)
		c.Ch <- utils.Exec(c.Cmd, c.WorkPath)
	}, ants.WithExpiryDuration(2*time.Minute))
}

func Release() {
	p.Release()
}

func (s *service) Judge(requestid string, data *entities.JudgePostData, languageInfo *languages.LanguageInfo) (*JudgeResponse, error) {
	langProfile := languageInfo.Language.Profile()
	workPath := getWorkspacePath(data.ID, requestid)
	fmt.Println(workPath)
	file := &utils.File{
		Path:    filepath.Join(workPath, langProfile.Filename),
		Content: []byte(data.Code),
	}
	err := utils.WriteFile(file)
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
	cmdCh := make(chan *cmd.Status)
	p.Invoke(&ExecJudgerArg{
		Ch:       cmdCh,
		WorkPath: workPath,
		Cmd:      judgerResult.DockerRunCmd,
	})
	cmdStatus := <-cmdCh
	fmt.Printf("cmdStatus %#v", cmdStatus)
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
