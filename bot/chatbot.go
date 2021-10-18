package bot

import (
	"errors"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/kevwan/chatbot/bot/corpus"
	"path/filepath"

	//"github.com/jinzhu/gorm"
	"github.com/kevwan/chatbot/bot/adapters/logic"
	"github.com/kevwan/chatbot/bot/adapters/storage"
	_ "github.com/mattn/go-sqlite3"
	"runtime"
	"time"
)

const mega = 1024 * 1024

type ChatBot struct {
	PrintMemStats bool
	//InputAdapter   input.InputAdapter
	LogicAdapter logic.LogicAdapter
	//OutputAdapter  output.OutputAdapter
	StorageAdapter storage.StorageAdapter
	Trainer        Trainer
	Config         Config
}

var engine *xorm.Engine

func (chatbot *ChatBot) Init() {
	var err error
	if engine == nil && chatbot.Config.Sqlite3 != "" {
		engine, err = xorm.NewEngine("sqlite3", chatbot.Config.Sqlite3)
	}

	engine.Sync2(&Corpus{})

	if err != nil {
		panic(err)
	}
	if chatbot.Config.DirCorpus != "" {
		files := chatbot.FindCorporaFiles(chatbot.Config.DirCorpus)
		if len(files) > 0 {
			corpuses, _ := chatbot.LoadCorpusFromFiles(files)
			if len(corpuses) > 0 {
				chatbot.SaveCorpusToDB(corpuses)

			}
		}
	}
	err = chatbot.TrainWithDB()
	if err != nil {
		panic(err)
	}
}

type Corpus struct {
	Id       int    `json:"id" form:"id" xorm:"int pk autoincr notnull 'id' comment('编号')"`
	Class    string `json:"class" form:"class"  xorm:"varchar(255) notnull 'class' comment('分类')"`
	Project  string `json:"project" form:"project" xorm:"varchar(255) notnull 'project' comment('项目')"`
	Question string `json:"question" form:"question"  xorm:"varchar(512) notnull index  'question' comment('问题')"`
	Answer   string `json:"answer" form:"answer" xorm:"varchar(102400) notnull  'answer' comment('回答')"`
	Qtype    int    `json:"qtype" form:"qtype" xorm:"int notnull 'qtype' comment('类型，需求，问答')"`
}

type Config struct {
	Sqlite3   string `json:"sqlite3"`
	Project   string `json:"project"`
	DirCorpus string `json:"dir_corpus"`
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
func (chatbot *ChatBot) LoadCorpusFromDB() (map[string][][]string, error) {
	results := make(map[string][][]string)
	var rows []Corpus
	query := Corpus{
		Project: chatbot.Config.Project,
	}
	err := engine.Find(&rows, &query)
	if err != nil {
		return nil, err
	}
	var corpuses [][]string
	for _, row := range rows {
		var corpus []string
		corpus = append(corpus, row.Question, row.Answer)
		corpuses = append(corpuses, corpus)
	}
	results[chatbot.Config.Project] = corpuses
	return results, nil

}

func (chatbot *ChatBot) LoadCorpusFromFiles(filePaths []string) (map[string][][]string, error) {
	return corpus.LoadCorpora(filePaths)
}

func (chatbot *ChatBot) SaveCorpusToDB(corpuses map[string][][]string) {
	for k, v := range corpuses {
		for _, cp := range v {
			if len(cp) == 2 {
				corpus := Corpus{
					Class:    k,
					Question: cp[0],
					Answer:   cp[1],
					Qtype:    1,
					Project:  chatbot.Config.Project,
				}
				chatbot.AddCorpusToDB(&corpus)
			}
		}
	}

}

func (chatbot *ChatBot) AddCorpusToDB(corpus *Corpus) error {
	q := Corpus{
		Question: corpus.Question,
		Class:    corpus.Class,
	}
	if ok, err := engine.Get(&q); !ok {
		_, err = engine.Insert(corpus)
		return err
	} else {
		if q.Id > 0 {
			corpus.Id = q.Id
			_, err = engine.Update(corpus, &Corpus{Id: q.Id})
			return err
		}
	}
	return nil
}

func (chatbot *ChatBot) RemoveCorpusFromDB(corpus *Corpus) error {
	q := Corpus{}
	if corpus.Id > 0 {
		q.Id = corpus.Id
	} else {
		if corpus.Question == "" {
			return errors.New("id or question must be set value")
		}
		q.Question = corpus.Question
	}
	if ok, err := engine.Get(&q); ok {
		chatbot.StorageAdapter.Remove(q.Question)
		engine.Delete(&q)
		return err
	}
	return nil
}

func (chatbot *ChatBot) TrainWithDB() error {
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

	corpuses, err := chatbot.LoadCorpusFromDB()
	if err != nil {
		return err
	}

	if err := chatbot.Trainer.TrainWithCorpus(corpuses); err != nil {
		return err
	} else {
		return nil
		//return chatbot.StorageAdapter.Sync()
	}

}

func (chatbot *ChatBot) FindCorporaFiles(dir string) []string {
	var files []string

	jsonFiles, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	files = append(files, jsonFiles...)

	ymlFiles, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	files = append(files, ymlFiles...)

	yamlFiles, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return append(files, yamlFiles...)
}

func (chatbot *ChatBot) GetResponse(text string) []logic.Answer {
	if chatbot.LogicAdapter.CanProcess(text) {
		return chatbot.LogicAdapter.Process(text)
	}

	return nil
}
