package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/kevwan/chatbot/bot"
	"github.com/kevwan/chatbot/bot/adapters/logic"
	"github.com/kevwan/chatbot/bot/adapters/storage"
	"github.com/sjqzhang/gdi"
)

var chatbot *bot.ChatBot

var (
	verbose  = flag.Bool("v", false, "verbose mode")
	tops     = flag.Int("t", 5, "the number of answers to return")
	dir      = flag.String("d", "/Users/dev/repo/chatterbot-corpus/chatterbot_corpus/data/chinese", "the directory to look for corpora files")
	sqliteDB = flag.String("sqlite3", "/Users/dev/repo/chatbot/chatbot.db", "the file path of the corpus sqlite3")
	//sqliteDB      = flag.String("sqlite3", "", "the file path of the corpus sqlite3")
	project       = flag.String("project", "DMS", "the name of the project in sqlite3 db")
	corpora       = flag.String("i", "", "the corpora files, comma to separate multiple files")
	storeFile     = flag.String("o", "/Users/dev/repo/chatbot/corpus.gob", "the file to store corpora")
	printMemStats = flag.Bool("m", true, "enable printing memory stats")
)

type JsonResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func init() {

	flag.Parse()
	gdi.Register(func() (*bot.ChatBot, error) {
		store, err := storage.NewSeparatedMemoryStorage(*storeFile)
		if err != nil {
			return nil, err
		}

		chatbot = &bot.ChatBot{
			LogicAdapter:   logic.NewClosestMatch(store, *tops),
			PrintMemStats:  *printMemStats,
			Trainer:        bot.NewCorpusTrainer(store),
			StorageAdapter: store,
			Config: bot.Config{
				Sqlite3:   *sqliteDB,
				Project:   *project,
				DirCorpus: *dir,
			},
		}
		chatbot.Init()

		if *verbose {
			chatbot.LogicAdapter.SetVerbose()
		}
		return chatbot, nil
	})

}

func bindRounter(router *gin.Engine) {
	v1 := router.Group("v1")
	v1.POST("add", func(context *gin.Context) {
		var corpus bot.Corpus
		context.Bind(&corpus)
		err := chatbot.AddCorpus(corpus)
		if err != nil {
			context.JSON(500, JsonResult{
				Code: 500,
				Msg:  err.Error(),
			})
			return
		}
		answer := make(map[string]int)
		answer[corpus.Answer] = 1
		chatbot.StorageAdapter.Update(corpus.Question, answer)
		chatbot.StorageAdapter.BuildIndex()
		context.JSON(200, JsonResult{
			Code: 0,
			Msg:  "success",
		})

	})

	v1.GET("search", func(context *gin.Context) {
		kw := context.Query("kw")
		results := chatbot.GetResponse(kw)
		msg := "ok"
		if len(results) == 0 {
			msg = "not found"
		}
		context.JSON(200, JsonResult{
			Code: 0,
			Msg:  msg,
			Data: results,
		})
	})

}
func main() {
	gdi.Init()
	chatbot.Init()
	router := gin.Default()
	bindRounter(router)
	router.Run(":8080")

}
