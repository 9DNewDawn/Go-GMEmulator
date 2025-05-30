package ApiRoutes

import (
	GameControllers "gm-emulator/api/domains/game/controllers"

	"github.com/go-chi/chi/v5"
)

func GetGameRoutes(r chi.Router) {
	r.Route("/game", func(r chi.Router) {
		r.Post("/edit-level", GameControllers.EditLevelHandler)
	})
}
