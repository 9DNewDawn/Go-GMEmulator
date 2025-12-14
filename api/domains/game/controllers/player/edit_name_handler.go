package GameControllers

import (
	"encoding/binary"
	"encoding/json"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
)

type MSG_GM_EDIT_NAME_DTO struct {
	CGAuthKey	string
	CCharacName string
	Name 		string
}

func EditNameHandler(w http.ResponseWriter, r *http.Request) {
	var packetDto MSG_GM_EDIT_NAME_DTO
	var packetMsg gms.MSG_GM_EDIT_NAME
	
	err := json.NewDecoder(r.Body).Decode(&packetDto)
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

	if packetDto.CGAuthKey != "some-auth-key" {
    	w.WriteHeader(http.StatusForbidden)
    	return
	}

	copy(packetMsg.Header.CGMName[:], "GMSE")
	copy(packetMsg.CCharacName[:], []byte(packetDto.CCharacName))
	copy(packetMsg.Name[:], []byte(packetDto.Name))
	packetMsg.Header.IKey = gms.MSG_KEY
	packetMsg.Header.CMessage = gms.MSG_GM_EDIT_NAME_NUM
	network.Send(&packetMsg, int(binary.Size(packetMsg)), conn)

	w.WriteHeader(http.StatusOK)
}