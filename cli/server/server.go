package main

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
	"github.com/kevwan/chatbot/bot"
	"github.com/kevwan/chatbot/bot/adapters/logic"
	"github.com/kevwan/chatbot/bot/adapters/storage"
	"github.com/sjqzhang/gdi"
	"net/http"
	"strings"
	"time"
)

var chatbot *bot.ChatBot

var (
	verbose = flag.Bool("v", false, "verbose mode")
	tops    = flag.Int("t", 5, "the number of answers to return")
	dir     = flag.String("d", "/Users/dev/repo/chatterbot-corpus/chatterbot_corpus/data/chinese", "the directory to look for corpora files")
	//sqliteDB = flag.String("sqlite3", "/Users/junqiang.zhang/repo/go/chatbot/chatbot.db", "the file path of the corpus sqlite3")
	sqliteDB      = flag.String("sqlite3", "", "the file path of the corpus sqlite3")
	bind          = flag.String("b", ":8080", "bind addr")
	project       = flag.String("project", "DMS", "the name of the project in sqlite3 db")
	corpora       = flag.String("i", "", "the corpora files, comma to separate multiple files")
	storeFile     = flag.String("o", "/Users/dev/repo/chatbot/corpus.gob", "the file to store corpora")
	printMemStats = flag.Bool("m", false, "enable printing memory stats")
)

type JsonResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type QA struct {
	Question string  `json:"question"`
	Answer   string  `json:"answer"`
	Score    float32 `json:"score"`
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
	buildAnswer := func(answers []logic.Answer) []QA {
		var qas []QA
		for _, answer := range answers {
			contents := strings.Split(answer.Content, "$$$$")
			if len(contents) > 1 {
				qa := QA{
					Question: contents[0],
					Answer:   contents[1],
					Score:    answer.Confidence,
				}
				qas = append(qas, qa)
			}
		}
		return qas
	}
	v1 := router.Group("v1")
	v1.POST("add", func(context *gin.Context) {
		var corpus bot.Corpus

		context.Bind(&corpus)
		err := chatbot.AddCorpusToDB(&corpus)
		if err != nil {
			context.JSON(500, JsonResult{
				Msg: err.Error(),
			})
			return
		}
		answer := make(map[string]int)
		answer[fmt.Sprintf("%s$$$$%s", corpus.Question, corpus.Answer)] = 1
		chatbot.StorageAdapter.Update(corpus.Question, answer)
		chatbot.StorageAdapter.BuildIndex()
		context.JSON(200, JsonResult{
			Code: 0,
			Msg:  "success",
		})

	})

	v1.GET("search", func(context *gin.Context) {
		q := context.Query("q")
		results := chatbot.GetResponse(q)
		qas := buildAnswer(results)
		msg := "ok"
		if len(results) == 0 {
			msg = "not found"
		}
		context.JSON(200, JsonResult{
			Code: 0,
			Msg:  msg,
			Data: qas,
		})
	})

	v1.POST("remove", func(context *gin.Context) {
		var corpus bot.Corpus

		context.Bind(&corpus)
		err := chatbot.RemoveCorpusFromDB(&corpus)
		if err != nil {
			context.JSON(500, JsonResult{
				Msg: err.Error(),
			})
			return
		}
		chatbot.StorageAdapter.BuildIndex()
		context.JSON(200, JsonResult{
			Code: 0,
			Msg:  "success",
		})

	})

}

func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	},
	)
}
func main() {
	gdi.Init()
	chatbot.Init()
	router := gin.Default()
	router.Use(Cors())
	box := packr.NewBox("./static")
	_=box
	router.StaticFS("/static", http.Dir("static"))
	bindRounter(router)
	router.Run(*bind)
}
