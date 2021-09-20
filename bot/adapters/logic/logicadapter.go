package logic

type (
	Answer struct {
		Content    string
		Confidence float32
	}

	LogicAdapter interface {
		CanProcess(string) bool
		Process(string) []Answer
		SetVerbose()
	}
)
