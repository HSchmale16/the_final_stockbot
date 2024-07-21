package app

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/handlebars/v2"
)

//go:embed html_templates/*
var templates embed.FS

func getEngine() *handlebars.Engine {
	// Check for debug environment variable
	if os.Getenv("DEBUG") == "true" {
		engine := handlebars.New("./internal/app/html_templates", ".hbs")
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

	// register an isEquals helper or else
	engine.AddFunc("isEqualApplyClass", func(a, b, class string) string {
		if a == b {
			return class
		}
		return ""
	})

	engine.AddFunc("eq", func(a, b string) bool {
		return a == b
	})

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

	engine.AddFunc("partyColor", func(s string) string {
		s = strings.ToLower(s)
		switch s {
		case "republican":
			return "red"
		case "independent":
			return "purple"
		case "democrat":
			return "blue"
		}
		return "slate"
	})

	engine.AddFunc("firstChar", func(s string) string {
		return string(s[0])
	})

	for k := range engine.Templates {
		fmt.Println(k)
	}

	return engine
}
