package main

import (
	"flag"
	"log"
	"strings"

	"github.com/kevwan/chatbot/bot"
	"github.com/kevwan/chatbot/bot/adapters/storage"
)

var (
	corpora       = flag.String("i", "", "the corpora files, comma to separate multiple files")
	storeFile     = flag.String("o", "corpus.gob", "the file to store corpora")
	printMemStats = flag.Bool("m", false, "enable printing memory stats")
)

func main() {
	flag.Parse()

	if len(*corpora) == 0 {
		flag.Usage()
		return
	}

	store, err := storage.NewSeparatedMemoryStorage(*storeFile)
	if err != nil {
		log.Fatal(err)
	}

	chatbot := &bot.ChatBot{
		PrintMemStats:  *printMemStats,
		Trainer:        bot.NewCorpusTrainer(store),
		StorageAdapter: store,
	}
	if err := chatbot.Train(strings.Split(*corpora, ",")); err != nil {
		log.Fatal(err)
	}
}
