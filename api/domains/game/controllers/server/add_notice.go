package GameControllers

import (
	"encoding/binary"
	"encoding/json"
	// "fmt"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
	"time"
)

type MSG_MESSAGE_DIRECT_DTO struct {
	CGAuthKey	string
	CString     string
}

func AddGameNotice(w http.ResponseWriter, r *http.Request) {
	maps := []int8{1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 17, 19, 20, 23}
	var addGameNoticeDto MSG_MESSAGE_DIRECT_DTO

	err := json.NewDecoder(r.Body).Decode(&addGameNoticeDto)
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

	if addGameNoticeDto.CGAuthKey != "some-auth-key" {
    	w.WriteHeader(http.StatusForbidden)
    	return
	}

	for _, m := range maps {
        var addGameNoticeMsg gms.MSG_MESSAGE_DIRECT
        copy(addGameNoticeMsg.Header.CGMName[:], "GMSE")
		copy(addGameNoticeMsg.CString[:], []byte(addGameNoticeDto.CString))
		addGameNoticeMsg.Header.IKey = gms.MSG_KEY
		addGameNoticeMsg.Header.CMessage = gms.MSG_MESSAGE_DIRECT_NUM
		addGameNoticeMsg.CServerNum = m
		network.Send(&addGameNoticeMsg, int(binary.Size(addGameNoticeMsg)), conn)
		time.Sleep(100 * time.Millisecond)
    }

	w.WriteHeader(http.StatusOK)
}