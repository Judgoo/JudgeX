package judge

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Judgoo/JudgeX/logger"
	"github.com/Judgoo/JudgeX/utils"
	"github.com/Judgoo/Judger/entities"
	judger "github.com/Judgoo/Judger/entities"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func getWorkspacePath(key string, requestid string) string {
	// 也许可以换成专业的文件系统来做这件事
	// 文件夹分层 b6eec00f2b9335ece97f7a8f8b2cfeb1 -> b6/ee/b6eec00f2b9335ece97f7a8f8b2cfeb1
	folder1 := requestid[:2]
	folder2 := requestid[2:4]
	prefix := requestid[:]

	// TODO `JudgeWorkspace` 这个换成放在设置项中的可配置的
	workDir := filepath.Join(os.TempDir(), "JudgeWorkspace", folder1, folder2)
	// 这样构造是因为这个 key 是需要返回到用户的，之后我们可以通过这个 key 找到本次判题究竟存在哪儿
	return path.Join(workDir, fmt.Sprintf("%s-%s", prefix, key))
}

type TestDataItem = *[2]*utils.File

func generateJudgerYml(workPath string, data *entities.JudgePostData, languageInfo *judger.JudgeInfo, testdataEntrys []string) (*judger.IJudger, error) {
	lang := languageInfo.Language
	capsToDrop := [...]string{"MKNOD"}
	var capsToDropString string
	for _, ct := range capsToDrop {
		capsToDropString += fmt.Sprintf("--cap-drop %s ", ct)
	}
	args := fmt.Sprintf("--privileged --cpus 2 -m 100m %s --rm -v %s:/workspace", strings.TrimSpace(capsToDropString), workPath)
	judgeCommand := fmt.Sprintf("podman --runtime /usr/bin/crun run %s %s", args, languageInfo.Version.Image)

	fmt.Println(judgeCommand)

	var judgerStruct = judger.IJudger{
		Language: lang.String(),
		Build:    languageInfo.Version.Build,
		Run:      languageInfo.Version.Run,
		RunnerArgs: &judger.IRunnerArgs{
			CpuTime: int(data.TimeLimit),
			Memory:  int(data.MemoryLimit),
			Mco:     languageInfo.Version.Mco,
			Stderr:  true,
		},
		TestData:     testdataEntrys,
		DockerRunCmd: judgeCommand,
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

func writeTestData(item TestDataItem, w *sync.WaitGroup) {
	utils.WriteFile(item[0])
	utils.WriteFile(item[1])
	w.Done()
}

func (s *service) Judge(requestid string, data *entities.JudgePostData, languageInfo *judger.JudgeInfo) (*JudgeResponse, error) {
	workPath := getWorkspacePath(data.Key, requestid)

	codeErrChan := make(chan error)

	go func() {
		file := &utils.File{
			Path:    filepath.Join(workPath, languageInfo.Version.Filename),
			Content: []byte(data.Code),
		}
		err := utils.WriteFile(file)
		codeErrChan <- err
	}()

	inputs := data.Inputs
	outputs := data.Outputs
	testdataList := make([]TestDataItem, 0)
	testdataEntrys := make([]string, 0, len(inputs)+1)

	for i := range inputs {
		inS := fmt.Sprintf("%d.in", i)
		outS := fmt.Sprintf("%d.out", i)
		entry := fmt.Sprintf("%s::%s", inS, outS)

		in := utils.File{
			Path:    path.Join(workPath, inS),
			Content: []byte(inputs[i]),
		}
		out := utils.File{
			Path:    path.Join(workPath, outS),
			Content: []byte(outputs[i]),
		}
		testdataList = append(testdataList, &[2]*utils.File{&in, &out})
		testdataEntrys = append(testdataEntrys, entry)
	}

	var w *sync.WaitGroup = new(sync.WaitGroup)
	w.Add(len(testdataList))
	for _, item := range testdataList {
		go writeTestData(item, w)
	}
	w.Wait()

	judgerResult, errG := generateJudgerYml(workPath, data, languageInfo, testdataEntrys)
	if errG != nil {
		return &JudgeResponse{}, errG
	}
	if codeErr := <-codeErrChan; codeErr != nil {
		return &JudgeResponse{}, codeErr
	}

	logger.Sugar.Infof("workPath: %s", workPath)
	cmdStatus := utils.Exec(judgerResult.DockerRunCmd, workPath, nil)
	logger.Sugar.Infof("workPath: %s cmdStatus %#v", workPath, cmdStatus)
	if cmdStatus.Error != nil {
		return &JudgeResponse{}, cmdStatus.Error
	}
	if cmdStatus.Complete {
		response := &JudgeResponse{
			Id: requestid,
			JudgerInfo: &JudgerInfo{
				Language: languageInfo.Language.String(),
				Version:  languageInfo.VersionName,
				Build:    languageInfo.Version.Build,
				Run:      languageInfo.Version.Run,
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
		logger.Sugar.Infow("parse result success", "result", result)
		logger.Sugar.Infof("result  %#v", result)
		if result.Code == judger.CodeSuccess {
			var r = make([]*RunnerSuccessResult, 0)
			var finalStatus judger.RunnerStatus = judger.ACCEPTED
			for _, item := range result.RunnerResult {
				_res := &RunnerSuccessResult{}
				if item.Code == judger.Success {
					// else {
					// 如果不等的话说明本次判题不成功
					// Judger 的顶层已经会报错了
					// }
					if item.Runner.Status > judger.ACCEPTED {
						finalStatus = item.Runner.Status
					}
					_res.Status = item.Runner.Status
					_res.CpuTimeUsed = item.Runner.CpuTimeUsed
					_res.CpuTimeUsedUs = item.Runner.CpuTimeUsedUs
					_res.RealTimeUsed = item.Runner.RealTimeUsed
					_res.RealTimeUsedUs = item.Runner.RealTimeUsedUs
					_res.MemoryUsed = item.Runner.MemoryUsed
				}
				r = append(r, _res)
			}
			response.Status = finalStatus
			response.Result = r
		} else {
			logger.Sugar.Infow("judger result code is not zero")
			judgeJudgerErrorResult(result, response)
			response.Message = result.Error
		}

		response.StatusInfo = GetStatusInfo(response.Status)
		return response, nil
	} else {
		// golang 在执行命令的时候出了问题, maybe I/O problem
		return &JudgeResponse{}, errors.WithMessage(cmdStatus.Error, "JudgeX 内部出现错误")
	}
}
