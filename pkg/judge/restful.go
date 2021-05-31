package judge

// import (
// 	_ "embed"
// 	"fmt"
// 	"strings"

// 	"github.com/Judgoo/languages"

// 	judger "github.com/Judgoo/Judger/entities"
// )

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
