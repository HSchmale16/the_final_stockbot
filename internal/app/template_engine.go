package app

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/handlebars/v2"
)

//go:embed html_templates/*
var templates embed.FS

func getEngine() *handlebars.Engine {
	// Check for debug environment variable
	if os.Getenv("DEBUG") == "true" {
		engine := handlebars.New("./html_templates", ".hbs")
		engine.Reload(true)
		return engine
	}
	subFS, err := fs.Sub(templates, "html_templates")
	if err != nil {
		panic(err)
	}
	engine := handlebars.NewFileSystem(http.FS(subFS), ".hbs")
	return engine
}

func GetTemplateEngine() fiber.Views {
	engine := getEngine()
	engine.AddFunc("formatDate", func(date string) string {
		// parse the date from isoformat 2023-11-08 00:22:00 +0000 UTC to Jan 2, 2006
		date = date[:10]
		layout := "2006-01-02"
		t, err := time.Parse(layout, date)
		if err != nil {
			fmt.Println(err)
			return "FUCK!"
		}
		return t.Format("Jan 2, 2006")
	})

	engine.AddFunc("firstChar", func(s string) string {
		return string(s[0])
	})

	for k := range engine.Templates {
		fmt.Println(k)
	}

	return engine
}
