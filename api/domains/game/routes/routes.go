package ApiRoutes

import (
	"fmt"
	GameControllers "gm-emulator/api/domains/game/controllers"

	"github.com/go-chi/chi/v5"
)

func GetGameRoutes(r chi.Router) {
	r.Route("/game", func(r chi.Router) {
		fmt.Println("Hello? Setting up game routes...")
		r.Get("/edit-level", GameControllers.EditLevelHandler)
	})
}
