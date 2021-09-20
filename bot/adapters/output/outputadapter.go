package output

type OutputAdapter interface {
	Process(string, float32) (string, bool)
}
