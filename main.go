package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knights-analytics/hugot"
	"github.com/robfig/cron"
)

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

const MODEL_PATH = "./DOWNLOAD_MODELS"

func DoAnalysis() {
	// start a new session. This looks for the onnxruntime.so library in its default path, e.g. /usr/lib/onnxruntime.so
	//session, err := hugot.NewSession()
	session, err := hugot.NewSession(hugot.WithOnnxLibraryPath("/usr/lib/libonnxruntime.so"))

	// if your onnxruntime.so is somewhere else, you can explicitly set it by using WithOnnxLibraryPath
	// session, err := hugot.NewSession(WithOnnxLibraryPath("/path/to/onnxruntime.so"))
	check(err)
	// A successfully created hugot session needs to be destroyed when you're done
	defer func(session *hugot.Session) {
		err := session.Destroy()
		check(err)
	}(session)

	modelPath, err := session.DownloadModel("KnightsAnalytics/distilbert-base-uncased-finetuned-sst-2-english", MODEL_PATH, hugot.NewDownloadOptions())
	//modelPath, err := session.DownloadModel("nickmuchi/sec-bert-finetuned-finance-classification", MODEL_PATH, hugot.NewDownloadOptions())
	check(err)

	// we now create the configuration for the text classification pipeline we want to create.
	// Options to the pipeline can be set here using the Options field
	config := hugot.TextClassificationConfig{
		ModelPath: modelPath,
		Name:      "testPipeline",
	}
	// then we create out pipeline.
	// Note: the pipeline will also be added to the session object so all pipelines can be destroyed at once
	sentimentPipeline, err := hugot.NewPipeline(session, config)
	check(err)

	// we can now use the pipeline for prediction on a batch of strings
	batch := []string{"This movie is disgustingly good!", "The director tried too much"}
	batchResult, err := sentimentPipeline.RunPipeline(batch)
	check(err)

	// and do whatever we want with it :)
	s, err := json.Marshal(batchResult)
	check(err)
	fmt.Println(string(s))
}

func main() {
	db, err := setupDB()
	if err != nil {
		// Handle error
		log.Println("Error occurred:", err)
		return
	}
	defer db.Close()

	c := cron.New()
	c.AddFunc("@every 10m", func() {
		fetchFeeds(db)
	})
	c.AddFunc("@daily", func() {
		downloadSecurities(db)
	})

	log.Print("Started feed reader cron.")
	c.Start()

	router := gin.Default()

	router.GET("/checkLoadedStories", func(c *gin.Context) {
		var count int64
		db.Model(&RSSItem{}).Where("DATE(published_date) >= DATE('now', '-2 days')").Count(&count)
		c.JSON(http.StatusOK, gin.H{
			"count": count,
		})
	})

	router.Run(":8080")
}
