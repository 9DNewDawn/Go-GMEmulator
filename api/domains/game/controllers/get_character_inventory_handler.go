package GameControllers

import (
	"encoding/binary"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
)

func GetCharacterInventoryHandler(w http.ResponseWriter, r *http.Request) {
	// Handler logic for getting character inventory
	var invReqMessage gms.MSG_INVEN_REQ
	characterName := r.Context().Value("character").(string)

	copy(invReqMessage.CCharacName[:], characterName)
	copy(invReqMessage.Header.CGMName[:], "Go-GMEmulator") // Assuming "GM" is the GM name
	invReqMessage.Header = gms.GmsHeader{
		IKey:     gms.MSG_KEY,           // Set appropriate key value
		CMessage: gms.MSG_INVEN_REQ_NUM, // Use the correct message ID
	}

	ctx := r.Context()
	conn, ok := ctx.Value("conn").(net.Conn)
	if !ok || conn == nil {
		http.Error(w, "Connection not found", http.StatusInternalServerError)
		return
	}

	network.Send(invReqMessage, int(binary.Size(invReqMessage)), conn)

	// read the response
	response, err := network.Read(conn)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	// process the response
	// ...

}
