package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kevwan/chatbot/bot"
	"github.com/kevwan/chatbot/bot/adapters/logic"
	"github.com/kevwan/chatbot/bot/adapters/storage"
)

var (
	verbose   = flag.Bool("v", false, "verbose mode")
	storeFile = flag.String("c", "corpus.gob", "the file to store corpora")
	tops      = flag.Int("t", 5, "the number of answers to return")
)

func main() {
	flag.Parse()

	store, err := storage.NewSeparatedMemoryStorage(*storeFile)
	if err != nil {
		log.Fatal(err)
	}

	chatbot := &bot.ChatBot{
		LogicAdapter: logic.NewClosestMatch(store, *tops),
	}
	if *verbose {
		chatbot.LogicAdapter.SetVerbose()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Q: ")
		scanner.Scan()
		question := scanner.Text()
		if question == "exit" {
			break
		}

		startTime := time.Now()
		answers := chatbot.GetResponse(question)
		if len(answers) == 0 {
			fmt.Println("No answer!")
			continue
		}

		if *tops == 1 {
			fmt.Printf("A: %s\n", answers[0].Content)
			continue
		}

		for i, answer := range answers {
			fmt.Printf("%d: %s\n", i+1, answer.Content)
			if *verbose {
				fmt.Printf("%d: %s\tConfidence: %.3f\t%s\n", i+1, answer.Content,
					answer.Confidence, time.Since(startTime))
			}
		}
		fmt.Println(time.Since(startTime))
	}
}
