package GameControllers

import (
	"encoding/binary"
	"encoding/json"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
)

type MSG_GM_ADD_INVGOLD_DTO struct {
	CGMName		string
	CCharacName	string
	IGold       int32
}

func EditGoldHandler(w http.ResponseWriter, r *http.Request) {
	var editGoldMsgDto MSG_GM_ADD_INVGOLD_DTO
	var editGoldMsg gms.MSG_GM_ADD_INVGOLD

	err := json.NewDecoder(r.Body).Decode(&editGoldMsgDto)
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

	copy(editGoldMsg.Header.CGMName[:], []byte(editGoldMsgDto.CGMName))
	copy(editGoldMsg.CCharacName[:], []byte(editGoldMsgDto.CCharacName))
	editGoldMsg.Header.IKey = gms.MSG_KEY
	editGoldMsg.Header.CMessage = gms.MSG_GM_ADD_INVGOLD_NUM
	editGoldMsg.IGold = editGoldMsgDto.IGold
	network.Send(&editGoldMsg, int(binary.Size(editGoldMsg)), conn)
	w.WriteHeader(http.StatusOK)
}