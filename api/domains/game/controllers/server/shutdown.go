package GameControllers

import (
	"encoding/binary"
	"encoding/json"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
)

type MSG_SYSTEM_OFF_REQ_DTO struct {
	CGAuthKey	string
	CServer_num int8
}

func ShutdownServer(w http.ResponseWriter, r *http.Request) {
	var callDto MSG_SYSTEM_OFF_REQ_DTO
	var callMsg gms.MSG_SYSTEM_OFF_REQ

	err := json.NewDecoder(r.Body).Decode(&callDto)
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

	if callDto.CGAuthKey != "some-auth-key" {
    	w.WriteHeader(http.StatusForbidden)
    	return
	}

	copy(callMsg.Header.CGMName[:], "GMSE")
	callMsg.Header.IKey = gms.MSG_KEY
	callMsg.Header.CMessage = gms.MSG_SYSTEM_OFF_REQ_NUM
	callMsg.CServer_num = callDto.CServer_num

	network.Send(&callMsg, int(binary.Size(callMsg)), conn)

	w.WriteHeader(http.StatusOK)
}