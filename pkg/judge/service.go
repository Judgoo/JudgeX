package judge

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Judgoo/JudgeX/logger"
	"github.com/Judgoo/JudgeX/utils"
	"github.com/Judgoo/languages"
	"github.com/pkg/errors"
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
	Status         judger.RunnerStatus `json:"status"`
	CpuTimeUsed    int                 `json:"cpu_time_used"`
	CpuTimeUsedUs  int                 `json:"cpu_time_used_us"`
	RealTimeUsed   int                 `json:"real_time_used"`
	RealTimeUsedUs int                 `json:"real_time_used_us"`
	MemoryUsed     int                 `json:"memory_used"`
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
}

type languageInfoMap map[string][]languageInfoDisplay

type Service interface {
	GetLanguages() languageInfoMap
	Judge(requestid string, data *judger.JudgePostData, li *judger.JudgeInfo) (*JudgeResponse, error)
}

type service struct {
}

func NewService() Service {
	return &service{}
}

func (s *service) GetLanguages() languageInfoMap {
	var result = languageInfoMap{}
	for lang, vss := range languages.LanguagesProfile {
		result[lang] = make([]languageInfoDisplay, 0)
		lt, err := languages.ParseLanguageType(lang)
		if err != nil {
			continue
		}
		for _, versionName := range vss {
			_versionName, versionInfo, ok := lt.GetVersionInfo(versionName)
			if !ok {
				continue
			}
			result[lang] = append(result[lang], languageInfoDisplay{
				_versionName,
				fmt.Sprintf("%s(%s)", lang, versionInfo.Name),
			})
		}
	}
	return result
}

func generatePodmanCmd(judgeInfo *judger.JudgeInfo) string {
	capsToDrop := [...]string{"MKNOD"}
	var capsToDropString string
	for _, ct := range capsToDrop {
		capsToDropString += fmt.Sprintf("--cap-drop %s ", ct)
	}
	args := fmt.Sprintf("--rm --privileged -i %s", strings.TrimSpace(capsToDropString))
	judgeCommand := fmt.Sprintf("podman --runtime /usr/bin/crun run %s %s", args, judgeInfo.Version.Image)
	return judgeCommand
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
	case judger.CodeRunError:
		response.Status = judger.RUNTIME_ERROR
	case judger.CodeNoInputDataError:
		response.Status = judger.SYSTEM_ERROR
	case judger.CodeInitLoggerError:
		fallthrough
	default:
		response.Status = judger.SYSTEM_ERROR
	}
}

// var podman = container.New()

// func (s *service) JudgebyRest(requestid string, postData *judger.JudgePostData, judgeInfo *judger.JudgeInfo) (*JudgeResponse, error) {
// 	cd := &judger.JudgerConsumeData{
// 		RequestId: requestid,
// 		Post:      *postData,
// 		JudgeInfo: *judgeInfo,
// 	}
// 	dataBytes, yErr := yaml.Marshal(cd)
// 	if yErr != nil {
// 		return &JudgeResponse{}, errors.WithMessage(yErr, "JudgeX 内部出现错误")
// 	}
// 	// cmd1 := generatePodmanCmd(judgeInfo)
// 	judgerResult, err := podman.Run(judgeInfo.Version.Image, dataBytes)
// 	if err != nil {
// 		return &JudgeResponse{}, errors.WithMessage(err, "JudgeX 内部出现错误")
// 	}

// 	response := &JudgeResponse{
// 		Id: requestid,
// 		JudgerInfo: &JudgerInfo{
// 			Language: judgeInfo.Language.String(),
// 			Version:  judgeInfo.VersionName,
// 			Build:    judgeInfo.Version.Build,
// 			Run:      judgeInfo.Version.Run,
// 		},
// 	}
// 	// 解析 Judger 的输出
// 	result := new(judger.NormalResult)
// 	err2 := json.Unmarshal([]byte(judgerResult.Stdout), &result)
// 	if err2 != nil {
// 		// Judger 没有输出一个有效 JSON
// 		// 说明 Judger 可能崩了
// 		logger.Sugar.Infow("judger output is not json format", "stdout", judgerResult.Stdout)
// 		return &JudgeResponse{}, ErrorJudgerError
// 	}
// 	if judgerResult.Exit != 0 {
// 		logger.Sugar.Infow("cmd exit non-zero")
// 		judgeJudgerErrorResult(result, response)
// 	}
// 	logger.Sugar.Infow("parse result success", "result", result)
// 	logger.Sugar.Infof("result  %#v", result)
// 	if result.Code == judger.CodeSuccess {
// 		var r = make([]*RunnerSuccessResult, 0)
// 		var finalStatus judger.RunnerStatus = judger.ACCEPTED
// 		for _, item := range result.RunnerResult {
// 			_res := &RunnerSuccessResult{}
// 			if item.Code == judger.Success {
// 				// else {
// 				// 如果不等的话说明本次判题不成功
// 				// Judger 的顶层已经会报错了
// 				// }
// 				if item.Runner.Status > judger.ACCEPTED {
// 					finalStatus = item.Runner.Status
// 				}
// 				_res.Status = item.Runner.Status
// 				_res.CpuTimeUsed = item.Runner.CpuTimeUsed
// 				_res.CpuTimeUsedUs = item.Runner.CpuTimeUsedUs
// 				_res.RealTimeUsed = item.Runner.RealTimeUsed
// 				_res.RealTimeUsedUs = item.Runner.RealTimeUsedUs
// 				_res.MemoryUsed = item.Runner.MemoryUsed
// 			}
// 			r = append(r, _res)
// 		}
// 		response.Status = finalStatus
// 		response.Result = r
// 	} else {
// 		logger.Sugar.Infow("judger result code is not zero")
// 		judgeJudgerErrorResult(result, response)
// 		response.Message = result.Error
// 	}

// 	response.StatusInfo = GetStatusInfo(response.Status)
// 	return response, nil
// }

func (s *service) Judge(requestid string, postData *judger.JudgePostData, judgeInfo *judger.JudgeInfo) (*JudgeResponse, error) {
	DockerRunCmd := generatePodmanCmd(judgeInfo)
	cd := &judger.JudgerConsumeData{
		RequestId: requestid,
		Post:      *postData,
		JudgeInfo: *judgeInfo,
	}
	dataBytes, yErr := yaml.Marshal(cd)
	if yErr != nil {
		return &JudgeResponse{}, errors.WithMessage(yErr, "JudgeX 内部出现错误")
	}
	cmdStatus := utils.Exec(DockerRunCmd, "", bytes.NewReader(dataBytes))
	logger.Sugar.Infof("requestid: %s cmdStatus %#v", requestid, cmdStatus)
	if cmdStatus.Error != nil {
		return &JudgeResponse{}, cmdStatus.Error
	}
	if cmdStatus.Complete {
		response := &JudgeResponse{
			Id: requestid,
			JudgerInfo: &JudgerInfo{
				Language: judgeInfo.Language.String(),
				Version:  judgeInfo.VersionName,
				Build:    judgeInfo.Version.Build,
				Run:      judgeInfo.Version.Run,
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
