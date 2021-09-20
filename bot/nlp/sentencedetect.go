package nlp

import (
	"strings"

	"github.com/tal-tech/go-zero/core/lang"
)

var (
	endQuestionChars = map[rune]lang.PlaceholderType{
		'?': lang.Placeholder,
		'？': lang.Placeholder,
		'吗': lang.Placeholder,
		'么': lang.Placeholder,
	}

	embededQuestionMarks = []string{
		"什么", "为何", "干嘛", "干吗", "怎么", "咋",
	}
)

func IsQuestion(sentence string) bool {
	chars := []rune(strings.TrimSpace(sentence))
	if len(chars) == 0 {
		return false
	}

	if _, ok := endQuestionChars[chars[len(chars)-1]]; ok {
		return true
	}

	for i := range embededQuestionMarks {
		if strings.Contains(sentence, embededQuestionMarks[i]) {
			return true
		}
	}

	return false
}
