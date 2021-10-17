package logic

type (
	Answer struct {
		//Title      string  `json:"title"`
		Content    string  `json:"content"`
		Confidence float32 `json:"confidence"`
	}

	LogicAdapter interface {
		CanProcess(string) bool
		Process(string) []Answer
		SetVerbose()
	}
)
