package GameControllers

import (
	"encoding/binary"
	"encoding/json"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
)

type MSG_GM_EDIT_ELIXIR_DTO struct {
	CGAuthKey	string
	CCharacName string
	CElixirType int8
}

func EditElixirHandler(w http.ResponseWriter, r *http.Request) {
	var packetDto MSG_GM_EDIT_ELIXIR_DTO
	var packetMsg gms.MSG_GM_EDIT_ELIXIR
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
	packetMsg.Header.IKey = gms.MSG_KEY
	packetMsg.Header.CMessage = gms.MSG_GM_EDIT_ELIXIR_NUM
	packetMsg.CElixirType = byte(packetDto.CElixirType)
	packetMsg.ASData[0] = 5
	packetMsg.ASData[1] = 6
	packetMsg.ASData[2] = 7
	packetMsg.ASData[3] = 8
	packetMsg.ASData[4] = 9
	packetMsg.ASData[5] = 10
	packetMsg.CGrade = 6
	network.Send(&packetMsg, int(binary.Size(packetMsg)), conn)

	w.WriteHeader(http.StatusOK)
}