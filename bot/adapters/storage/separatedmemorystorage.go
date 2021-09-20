package storage

import (
	"encoding/gob"
	"os"

	"github.com/kevwan/chatbot/bot/nlp"
)

type separatedMemoryStorage struct {
	filepath           string
	declarativeStorage GobStorage
	questionStorage    GobStorage
}

func NewSeparatedMemoryStorage(filepath string) (*separatedMemoryStorage, error) {
	var declarativeStorage, questionStorage GobStorage

	if _, err := os.Stat(filepath); err == nil {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		decoder := gob.NewDecoder(f)
		if declarativeStorage, err = RestoreMemoryStorage(decoder); err != nil {
			return nil, err
		}

		if questionStorage, err = RestoreMemoryStorage(decoder); err != nil {
			return nil, err
		}
	} else {
		declarativeStorage = NewMemoryStorage()
		questionStorage = NewMemoryStorage()
	}

	return &separatedMemoryStorage{
		filepath:           filepath,
		declarativeStorage: declarativeStorage,
		questionStorage:    questionStorage,
	}, nil
}

func (storage *separatedMemoryStorage) BuildIndex() {
	storage.declarativeStorage.BuildIndex()
	storage.questionStorage.BuildIndex()
}

func (storage *separatedMemoryStorage) Count() int {
	return storage.declarativeStorage.Count() + storage.questionStorage.Count()
}

func (storage *separatedMemoryStorage) Find(sentence string) (map[string]int, bool) {
	if nlp.IsQuestion(sentence) {
		return storage.questionStorage.Find(sentence)
	} else {
		return storage.declarativeStorage.Find(sentence)
	}
}

func (storage *separatedMemoryStorage) Search(sentence string) []string {
	if nlp.IsQuestion(sentence) {
		return storage.questionStorage.Search(sentence)
	} else {
		return storage.declarativeStorage.Search(sentence)
	}
}

func (storage *separatedMemoryStorage) Remove(sentence string) {
	if nlp.IsQuestion(sentence) {
		storage.questionStorage.Remove(sentence)
	} else {
		storage.declarativeStorage.Remove(sentence)
	}
}

func (storage *separatedMemoryStorage) Sync() error {
	f, err := os.Create(storage.filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := gob.NewEncoder(f)

	storage.declarativeStorage.SetOutput(encoder)
	if err := storage.declarativeStorage.Sync(); err != nil {
		return err
	}

	storage.questionStorage.SetOutput(encoder)
	return storage.questionStorage.Sync()
}

func (storage *separatedMemoryStorage) Update(sentence string, responses map[string]int) {
	if nlp.IsQuestion(sentence) {
		storage.questionStorage.Update(sentence, responses)
	} else {
		storage.declarativeStorage.Update(sentence, responses)
	}
}
