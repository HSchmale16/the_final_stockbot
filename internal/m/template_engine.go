package m

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/handlebars/v2"
	"github.com/yalue/merged_fs"
)

//go:embed html_templates/*
var builtInTemplates embed.FS

func init() {
	RegisterEmbededFS(builtInTemplates)
	RegisterDebugFilePath("internal/m/html_templates")
}

func RegisterDebugFilePath(path string) {
	if os.Getenv("DEBUG") == "true" {
		if strings.HasPrefix(path, "/") {
			path = path[1:]
			log.Println("Are you a dumbass we never debug from root? Use a relative path!!!")
		}
		log.Println("Registering debug path: ", path)
		// Convert the path to a fs.FS
		target1 := os.DirFS(path)

		templatesFS = append(templatesFS, target1)
	}
}

func RegisterEmbededFS(embededFS embed.FS) {
	if os.Getenv("DEBUG") == "true" {
		return
	}
	subFS, err := fs.Sub(embededFS, "html_templates")
	if err != nil {
		log.Fatal(err)
	}
	templatesFS = append(templatesFS, subFS)
}

var templatesFS = make([]fs.FS, 0, 10)

func getEngine() *handlebars.Engine {
	fmt.Println(templatesFS)
	myFS := merged_fs.MergeMultiple(templatesFS...)

	engine := handlebars.NewFileSystem(http.FS(myFS), ".hbs")

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

	engine.AddFunc("eqTernary", func(a, b, c, d string) string {
		if a == b {
			return c
		}
		return d
	})

	engine.AddFunc("formatDate", func(date string) string {
		// parse the date from isoformat 2023-11-08 00:22:00 +0000 UTC to Jan 2, 2006
		date = date[:10]
		layout := "2006-01-02"
		t, err := time.Parse(layout, date)
		if err != nil {
			fmt.Println(err)
			return "The date is stupid!"
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

	engine.AddFunc("eqTernaryShort", eqTernaryShort)

	// This is the end
	err := engine.Load()
	if err != nil {
		log.Fatal(err)
	}
	for k := range engine.Templates {
		fmt.Println(k)
	}

	return engine
}

/**
 * Compares 2 a and b strings up to the length of the shortest string. If they match return c else d
 */
func eqTernaryShort(a, b, c, d string) string {
	length := min(len(a), len(b))

	if a == b {
		return c
	} else if length < 2 {
		return d
	}

	for i := 0; i < length; i++ {
		if a[i] != b[i] {
			return d
		}
	}
	return c
}
