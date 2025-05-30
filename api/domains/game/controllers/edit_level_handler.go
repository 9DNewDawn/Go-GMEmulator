package GameControllers

import (
	"encoding/json"
	"fmt"
	"gm-emulator/gms"
	"net/http"
)

type MSG_GM_EDIT_LEVEL_DTO struct {
	CGMName     string
	CCharacName string
	ILevel      int32
}

func EditLevelHandler(w http.ResponseWriter, r *http.Request) {
	var editLevelMsgDto MSG_GM_EDIT_LEVEL_DTO
	var editLevelMsg gms.MSG_GM_EDIT_LEVEL
	// Extract the level ID from the URL parameters
	err := json.NewDecoder(r.Body).Decode(&editLevelMsgDto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Person: %+v", editLevelMsg)
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(response)
}
