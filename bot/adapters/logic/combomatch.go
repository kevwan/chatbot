package logic

type comboMatch struct {
	matches []LogicAdapter
}

func NewComboMatch(matches ...LogicAdapter) LogicAdapter {
	return &comboMatch{
		matches: matches,
	}
}

func (match *comboMatch) CanProcess(question string) bool {
	for _, each := range match.matches {
		if each.CanProcess(question) {
			return true
		}
	}
	return false
}

func (match *comboMatch) Process(question string) []Answer {
	for _, each := range match.matches {
		if each.CanProcess(question) {
			return each.Process(question)
		}
	}
	return nil
}

func (match *comboMatch) SetVerbose() {
	for _, each := range match.matches {
		each.SetVerbose()
	}
}
