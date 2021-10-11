package nlp

import (
	"strings"
	"unicode"

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
	// we don't check whether question or not in English
	if isAscii(sentence) {
		return false
	}

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

func isAscii(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}
