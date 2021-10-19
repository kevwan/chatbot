package bot

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kevwan/chatbot/bot/adapters/storage"
	"github.com/kevwan/chatbot/bot/corpus"
)

type (
	Trainer interface {
		Train(interface{}) error
		TrainWithCorpus(corpuses map[string][][]string) error
	}

	ConversationTrainer struct {
		storage storage.StorageAdapter
	}

	CorpusTrainer struct {
		storage storage.StorageAdapter
	}
)

func NewConversationTrainer(storage storage.StorageAdapter) *ConversationTrainer {
	return &ConversationTrainer{
		storage: storage,
	}
}

func (trainer *ConversationTrainer) getOrCreate(text string) map[string]int {
	if value, ok := trainer.storage.Find(text); ok {
		return value
	} else {
		return make(map[string]int)
	}
}

func (trainer *ConversationTrainer) Train(data interface{}) error {
	sentences, ok := data.([]string)
	if !ok {
		return errors.New("ConversationTrainer.Train needs arguments to be []string")
	}

	if len(sentences) == 2 {
		sentences[1] = fmt.Sprintf("%s$$$$%s", sentences[0], sentences[1])
	}

	var history string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) == 0 {
			continue
		}

		if len(history) > 0 {
			responses := trainer.getOrCreate(history)
			responses[sentence] += 1
			trainer.storage.Update(history, responses)
		}

		history = sentence
	}

	return nil
}

func NewCorpusTrainer(storage storage.StorageAdapter) *CorpusTrainer {
	return &CorpusTrainer{
		storage: storage,
	}
}

func (trainer *CorpusTrainer) TrainWithCorpus(corpuses map[string][][]string) error {

	convTrainer := NewConversationTrainer(trainer.storage)

	for _, convs := range corpuses {
		for _, conv := range convs {
			convTrainer.Train(conv)
		}
	}
	trainer.storage.BuildIndex()
	return nil
}

func (trainer *CorpusTrainer) Train(data interface{}) error {
	files, ok := data.([]string)
	if !ok {
		return errors.New("CorpusTrainer.Train needs argument to be []string")
	}

	fmt.Println("Loading corpora...")

	convTrainer := NewConversationTrainer(trainer.storage)
	corpora, err := corpus.LoadCorpora(files)
	if err != nil {
		return err
	}

	fmt.Println("Creating Q/A mappings...")

	for _, convs := range corpora {
		for _, conv := range convs {
			convTrainer.Train(conv)
		}
	}

	fmt.Println("Building indexes...")

	trainer.storage.BuildIndex()
	return nil
}
