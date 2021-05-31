package judge

import (
	_ "embed"

	"github.com/Judgoo/languages"

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
	VersionId   string `json:"id"`
	DisplayName string `json:"name"`
}

type languageInfoMap map[string][]languageInfoDisplay

type Service interface {
	GetLanguages() languageInfoMap
	Judge(requestid string, data *judger.JudgePostData, li *judger.JudgeInfo) (*JudgeResponse, error)
	JudgeX(requestid string, data *judger.JudgePostData, li *judger.JudgeInfo) (*JudgeResponse, error)
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
		for _, versionId := range vss {
			realVersionId, versionInfo, ok := lt.GetVersionInfo(versionId)
			if !ok {
				continue
			}
			result[lang] = append(result[lang], languageInfoDisplay{
				realVersionId,
				versionInfo.Name,
			})
		}
	}
	return result
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
