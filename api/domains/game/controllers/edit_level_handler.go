package GameControllers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func EditLevelHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the level ID from the URL parameters
	levelID := chi.URLParam(r, "levelID")

	// For demonstration purposes, let's assume we are editing a level with the given ID
	// In a real application, you would retrieve the level data from a database or other storage
	// and apply the changes based on the request body.

	// Here we just simulate a successful edit operation
	response := map[string]string{
		"message": "Level " + levelID + " edited successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
