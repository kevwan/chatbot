package logic

import (
	"fmt"
	"math"
	"sort"

	"github.com/kevwan/chatbot/bot/adapters/storage"
	"github.com/kevwan/chatbot/bot/nlp"
	"github.com/zeromicro/go-zero/core/mr"
)

const (
	chunkSize     = 10000
	topAnswerSize = 10
)

type (
	sourceAndTargets struct {
		source  string
		targets []string
	}

	questionAndScore struct {
		question string
		score    float32
	}

	answerAndOccurrence struct {
		answer     string
		occurrence int
	}

	topOccurAnswers struct {
		answers []*answerAndOccurrence
	}

	topScoreQuestions struct {
		questions []questionAndScore
	}

	closestMatch struct {
		verbose bool
		storage storage.StorageAdapter
		tops    int
	}
)

func NewClosestMatch(storage storage.StorageAdapter, tops int) LogicAdapter {
	return &closestMatch{
		storage: storage,
		tops:    tops,
	}
}

func (match *closestMatch) CanProcess(string) bool {
	return true
}

func (match *closestMatch) Process(text string) []Answer {
	if responses, ok := match.storage.Find(text); ok {
		return match.processExactMatch(responses)
	} else {
		return match.processSimilarMatch(text)
	}
}

func (match *closestMatch) SetVerbose() {
	match.verbose = true
}

func (match *closestMatch) processExactMatch(responses map[string]int) []Answer {
	var top topOccurAnswers

	for key, occurrence := range responses {
		top.put(key, occurrence)
	}

	sort.Slice(top.answers, func(i, j int) bool {
		return top.answers[i].occurrence > top.answers[j].occurrence
	})

	tops := match.tops
	if len(top.answers) < tops {
		tops = len(top.answers)
	}

	answers := make([]Answer, tops)
	for i := 0; i < tops; i++ {
		answers[i].Content = top.answers[i].answer
		answers[i].Confidence = 1
	}

	return answers
}

func (match *closestMatch) processSimilarMatch(text string) []Answer {
	result, err := mr.MapReduce(generator(match, text), mapper(match), reducer(match))
	if err != nil {
		return nil
	}

	var answers []Answer
	slice := result.([]questionAndScore)
	for _, each := range slice {
		if each.score > 0 {
			if responses, ok := match.storage.Find(each.question); ok {
				matches := match.processExactMatch(responses)
				if len(matches) > 0 {
					answers = append(answers, Answer{
						Content:    matches[0].Content,
						Confidence: each.score,
					})
				}
			}
		}
	}

	return answers
}

func (top *topOccurAnswers) put(answer string, occurrence int) {
	if len(top.answers) < topAnswerSize {
		top.answers = append(top.answers, &answerAndOccurrence{
			answer:     answer,
			occurrence: occurrence,
		})
	} else {
		var leastIndex int
		leastOccurrence := math.MaxInt32

		for i, each := range top.answers {
			if each.occurrence < leastOccurrence {
				leastOccurrence = each.occurrence
				leastIndex = i
			}
		}

		if leastOccurrence < occurrence {
			top.answers[leastIndex] = &answerAndOccurrence{
				answer:     answer,
				occurrence: occurrence,
			}
		}
	}
}

func generator(match *closestMatch, text string) mr.GenerateFunc {
	return func(source chan<- interface{}) {
		keys := match.storage.Search(text)
		if match.verbose {
			printMatches(keys)
		}

		chunks := splitStrings(keys, chunkSize)
		for _, chunk := range chunks {
			source <- sourceAndTargets{
				source:  text,
				targets: chunk,
			}
		}
	}
}

func mapper(match *closestMatch) mr.MapperFunc {
	return func(data interface{}, writer mr.Writer, cancel func(error)) {
		tops := newTopScoreQuestions(match.tops)
		pair := data.(sourceAndTargets)
		for i := range pair.targets {
			score := nlp.SimilarityForStrings(pair.source, pair.targets[i])
			tops.add(questionAndScore{
				question: pair.targets[i],
				score:    score,
			})
		}

		writer.Write(tops)
	}
}

func reducer(match *closestMatch) mr.ReducerFunc {
	return func(input <-chan interface{}, writer mr.Writer, cancel func(error)) {
		tops := newTopScoreQuestions(match.tops)
		for each := range input {
			qs := each.(*topScoreQuestions)
			for _, question := range qs.questions {
				tops.add(question)
			}
		}

		sort.Slice(tops.questions, func(i, j int) bool {
			return tops.questions[i].score > tops.questions[j].score
		})

		writer.Write(tops.questions)
	}
}

func splitStrings(slice []string, size int) [][]string {
	var result [][]string
	count := len(slice)

	for i := 0; i < count; i += size {
		var end int
		if i+size < count {
			end = i + size
		} else {
			end = count
		}
		result = append(result, slice[i:end])
	}

	return result
}

func printMatches(matches []string) {
	fmt.Println("matched size:", len(matches))
	for i, sentence := range matches {
		if i > 10 {
			break
		}
		fmt.Println("\t", sentence)
	}
}

func newTopScoreQuestions(n int) *topScoreQuestions {
	return &topScoreQuestions{
		questions: make([]questionAndScore, n),
	}
}

func (tq *topScoreQuestions) add(q questionAndScore) {
	var score float32 = 1
	var index int
	for i, each := range tq.questions {
		if each.score < score {
			score = each.score
			index = i
		}
	}
	if q.score > score {
		tq.questions[index] = q
	}
}
