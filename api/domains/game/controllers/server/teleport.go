package GameControllers

import (
	"encoding/binary"
	"encoding/json"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
	"strconv"
    "strings"
)

type MSG_GM_EDIT_ZONE_DTO struct {
	CGAuthKey	string
	CCharacName string
	IZone       string
	FX          float32
    FZ          float32
}

func TeleportCharacter(w http.ResponseWriter, r *http.Request) {
	var zoneDto MSG_GM_EDIT_ZONE_DTO
	var zoneMsg gms.MSG_GM_EDIT_ZONE

	err := json.NewDecoder(r.Body).Decode(&zoneDto)
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

	if zoneDto.CGAuthKey != "some-auth-key" {
    	w.WriteHeader(http.StatusForbidden)
    	return
	}

	copy(zoneMsg.Header.CGMName[:], "GMSE")
	copy(zoneMsg.CCharacName[:], []byte(zoneDto.CCharacName))
	zoneMsg.Header.IKey = gms.MSG_KEY
	zoneMsg.Header.CMessage = gms.MSG_GM_EDIT_ZONE_NUM
	zoneStr := strings.TrimSpace(zoneDto.IZone)
    z, err := strconv.Atoi(zoneStr)
    if err != nil {
        http.Error(w, "Invalid MapID: "+err.Error(), http.StatusBadRequest)
        return
    }
    zoneMsg.IZone = int32(z)
	zoneMsg.FX = zoneDto.FX
	zoneMsg.FZ = zoneDto.FZ
	network.Send(&zoneMsg, int(binary.Size(zoneMsg)), conn)

	w.WriteHeader(http.StatusOK)
}