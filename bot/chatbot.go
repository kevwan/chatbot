package bot

import (
	"database/sql"
	"fmt"
	"github.com/kevwan/chatbot/bot/adapters/input"
	"github.com/kevwan/chatbot/bot/adapters/logic"
	"github.com/kevwan/chatbot/bot/adapters/output"
	"github.com/kevwan/chatbot/bot/adapters/storage"
	_ "github.com/mattn/go-sqlite3"
	"runtime"
	"time"
)

const mega = 1024 * 1024

type ChatBot struct {
	PrintMemStats  bool
	InputAdapter   input.InputAdapter
	LogicAdapter   logic.LogicAdapter
	OutputAdapter  output.OutputAdapter
	StorageAdapter storage.StorageAdapter
	Trainer        Trainer
}

func (chatbot *ChatBot) Train(data interface{}) error {
	start := time.Now()
	defer func() {
		fmt.Printf("Elapsed: %s\n", time.Since(start))
	}()

	if chatbot.PrintMemStats {
		go func() {
			for {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Printf("Alloc = %vm\nTotalAlloc = %vm\nSys = %vm\nNumGC = %v\n\n",
					m.Alloc/mega, m.TotalAlloc/mega, m.Sys/mega, m.NumGC)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	if err := chatbot.Trainer.Train(data); err != nil {
		return err
	} else {
		return chatbot.StorageAdapter.Sync()
	}
}

/*
CREATE TABLE "chatbot_tab" (
"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
"class" TEXT,
"project" TEXT,
"question" TEXT,
"answer" TEXT,
"qtype" INTEGER
);
*/
func LoadCorpusFromSqite(dbpath string, project string) (map[string][][]string, error) {
	results := make(map[string][][]string)
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("SELECT id,question,answer,class FROM corpus_tab where project=?", project)
	if err != nil {
		return nil, err
	}
	var corpuses [][]string
	for rows.Next() {
		var question string
		var answer string
		var class string
		var id int
		var corpus []string
		err = rows.Scan(&id, &question, &answer, &class)
		if err != nil {
			return nil, err
		}
		corpus = append(corpus, question, answer)
		corpuses = append(corpuses, corpus)
	}
	results[project] = corpuses
	return results, nil

}

func (chatbot *ChatBot) TrainWithSqite(dbpath string, project string) error {
	start := time.Now()
	defer func() {
		fmt.Printf("Elapsed: %s\n", time.Since(start))
	}()

	if chatbot.PrintMemStats {
		go func() {
			for {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Printf("Alloc = %vm\nTotalAlloc = %vm\nSys = %vm\nNumGC = %v\n\n",
					m.Alloc/mega, m.TotalAlloc/mega, m.Sys/mega, m.NumGC)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	corpuses,err:=LoadCorpusFromSqite(dbpath, project)
	if err!=nil {
		return err
	}

	fmt.Println(corpuses)

	//chatbot.Trainer.Train()

	return nil

	//if err := chatbot.Trainer.Train(data); err != nil {
	//	return err
	//} else {
	//	return chatbot.StorageAdapter.Sync()
	//}
}

func (chatbot *ChatBot) GetResponse(text string) []logic.Answer {
	if chatbot.LogicAdapter.CanProcess(text) {
		return chatbot.LogicAdapter.Process(text)
	}

	return nil
}
