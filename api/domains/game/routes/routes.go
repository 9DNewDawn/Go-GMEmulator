package ApiRoutes

import (
	"context"
	GameControllersPlayer "gm-emulator/api/domains/game/controllers/player"
	GameControllersServer "gm-emulator/api/domains/game/controllers/server"
	// GameControllersInventory "gm-emulator/api/domains/game/controllers/player/inventory"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetGameRoutes(r chi.Router) {
	r.Route("/server", func(r chi.Router) {
		r.Post("/edit-level", GameControllersPlayer.EditLevelHandler)              // POST /api/server/edit-level
		r.Post("/edit-gold", GameControllersPlayer.EditGoldHandler)                // POST /api/server/edit-gold
		r.Post("/notice", GameControllersServer.AddGameNotice)					   // POST /api/server/notice
		r.Post("/teleport", GameControllersServer.TeleportCharacter)			   // POST /api/server/teleport
		r.Post("/shutdown", GameControllersServer.ShutdownServer)				   // POST /api/server/shutdown
		r.Post("/elixir", GameControllersPlayer.EditElixirHandler)				   // POST /api/server/elixir
		r.Post("/rename", GameControllersPlayer.EditNameHandler)				   // POST /api/server/rename
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
