package entities

type JudgePostData struct {
	// 需要把这个 id 回调回去
	ID          string   `json:"id" validate:"required"`
	DownloadUrl string   `json:"download_url"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
	TimeLimit   uint     `json:"time_limit"`
	MemoryLimit uint     `json:"memory_limit"`
	Code        string   `json:"code" validate:"required"`
	CallbackUrl string   `json:"callback"`
}
