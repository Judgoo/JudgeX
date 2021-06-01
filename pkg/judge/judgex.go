package judge

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Judgoo/JudgeX/logger"
	"github.com/Judgoo/JudgeX/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	judger "github.com/Judgoo/Judger/entities"
)

func generateCRunXCmd(judgeInfo *judger.JudgeInfo) string {
	capsToDrop := [...]string{"MKNOD"}
	var capsToDropString string
	for _, ct := range capsToDrop {
		capsToDropString += fmt.Sprintf("--cap-drop %s ", ct)
	}
	args := fmt.Sprintf("--rm --privileged -i %s", strings.TrimSpace(capsToDropString))
	judgeCommand := fmt.Sprintf("docker run %s %s -x", args, judgeInfo.Version.Image)
	return judgeCommand
}

func (s *service) JudgeX(requestid string, postData *judger.JudgePostData, judgeInfo *judger.JudgeInfo) (*JudgeResponse, error) {
	cRunCmd := generateCRunXCmd(judgeInfo)
	fmt.Printf("cRunCmd %s", cRunCmd)
	cd := &judger.JudgerConsumeData{
		RequestId: requestid,
		Post:      *postData,
		JudgeInfo: *judgeInfo,
	}
	dataBytes, yErr := yaml.Marshal(cd)
	if yErr != nil {
		return &JudgeResponse{}, errors.WithMessage(yErr, "JudgeX 内部出现错误")
	}
	cmdStatus := utils.Exec(cRunCmd, "", bytes.NewReader(dataBytes))
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
