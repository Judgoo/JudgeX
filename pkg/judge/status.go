package judge

import judger "github.com/Judgoo/Judger/entities"

var StatusInfos = &map[judger.RunnerStatus]string{
	// 还未执行答案检查
	judger.PENDING: "PENDING",
	// 答案正确
	judger.ACCEPTED: "ACCEPTED",
	// 换行问题，输出的换行符有误
	judger.PRESENTATION_ERROR: "PRESENTATION_ERROR",
	// 超时
	judger.TIME_LIMIT_EXCEEDED: "TIME_LIMIT_EXCEEDED",
	// 超内存限制
	judger.MEMORY_LIMIT_EXCEEDED: "MEMORY_LIMIT_EXCEEDED",
	// 答案错误
	judger.WRONG_ANSWER: "WRONG_ANSWER",
	// 用户的程序运行时发生错误
	judger.RUNTIME_ERROR: "RUNTIME_ERROR",
	// 编译错误
	judger.COMPILE_ERROR: "COMPILE_ERROR",
	// 判题系统发生错误
	judger.SYSTEM_ERROR: "SYSTEM_ERROR",
}

func GetStatusInfo(rs judger.RunnerStatus) string {
	return (*StatusInfos)[rs]
}
