package ApiRoutes

import (
	"context"
	GameControllersPlayer "gm-emulator/api/domains/game/controllers/player"
	GameControllersInventory "gm-emulator/api/domains/game/controllers/player/inventory"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetGameRoutes(r chi.Router) {
	r.Route("/game", func(r chi.Router) {
		r.Route("/character", func(r chi.Router) {
			r.Route("/{characterName}", func(r chi.Router) {
				r.Use(CharacterCtx)                                                        // Middleware to set player context
				r.Post("/edit-level", GameControllersPlayer.EditLevelHandler)              // POST /characters/{characterName}/edit-level
				r.Get("/inventory", GameControllersInventory.GetCharacterInventoryHandler) // GET /characters/{characterName}/inventory
				// r.Put("/inventory", GameC)

			})
		})
	})
}

func getPlayerRoutes(r chi.Router) {

}

func CharacterCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		characName := chi.URLParam(r, "characterName")

		ctx := context.WithValue(r.Context(), "character", characName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// func getInventoryRoutes(r chi.Router) {
// 	r.Route("/inventory", func(r chi.Router) {
// 		r.Get("/", getInventory)          // GET /inventory
// 		r.Post("/", addItem)              // POST /inventory
// 		r.Delete("/{itemId}", deleteItem) // DELETE /inventory/123
// 	})
// }
