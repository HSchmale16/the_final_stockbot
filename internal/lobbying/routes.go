package lobbying

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App) {
	app.Get("/lobbying", RenderLobbyingPage)
}

func RenderLobbyingPage(c *fiber.Ctx) error {
	return c.Render("lobbying", fiber.Map{
		"Title": "Lobbying",
	})
}
