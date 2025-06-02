package GameControllers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
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
	ctx := r.Context()
	conn, ok := ctx.Value("conn").(net.Conn)
	if !ok || conn == nil {
		http.Error(w, "Connection not found", http.StatusInternalServerError)
		return
	}
	// copy everything to the message
	copy(editLevelMsg.CCharacName[:], []byte(editLevelMsgDto.CCharacName))
	copy(editLevelMsg.Header.CGMName[:], []byte(editLevelMsgDto.CGMName))
	editLevelMsg.Header.IKey = gms.MSG_KEY
	editLevelMsg.Header.CMessage = gms.MSG_GM_EDIT_LEVEL_NUM
	editLevelMsg.ILevel = editLevelMsgDto.ILevel

	fmt.Fprintf(w, "Person: %+v", editLevelMsg)

	network.Send(&editLevelMsg, int(binary.Size(editLevelMsg)), conn)
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(response)
}
