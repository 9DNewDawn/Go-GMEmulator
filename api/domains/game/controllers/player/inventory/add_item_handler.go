package GameControllers

import (
	"encoding/binary"
	"fmt"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
	"time"
)

type MSG_GM_ADD_ITEM_DTO struct {
	CGMName      string
	CFirstType   uint8
	CSecondType  uint8
	SItemID      int16
	UCItemCount  uint8
	USDurability uint16
	UCSlotCount  uint8
	UCInchant    uint8
}

func AddItem(w http.ResponseWriter, r *http.Request) {
	conn, ok := r.Context().Value("conn").(net.Conn)
	if !ok {
		http.Error(w, "No connection available", http.StatusInternalServerError)
		return
	}

	channel, ok := r.Context().Value("channel").(chan []byte)
	if !ok {
		http.Error(w, "No channel available", http.StatusInternalServerError)
		return
	}

	_ = r.Context().Value("connID").(int)

	// Get character name from context or URL parameter
	characterName, ok := r.Context().Value("character").(string)
	if !ok {
		http.Error(w, "No character name available", http.StatusBadRequest)
		return
	}

	// STEP 1: Send character request first
	var characReqMessage gms.MSG_CHARAC_REQ
	characReqMessage.Header = gms.GmsHeader{
		IKey:     gms.MSG_KEY,
		CMessage: gms.MSG_CHARAC_REQ_NUM,
		UITime:   0,
	}

	copy(characReqMessage.Header.CGMName[:], "Go-GMEmulator")
	copy(characReqMessage.CCharacName[:], characterName)

	result := network.Send(characReqMessage, int(binary.Size(characReqMessage)), conn)
	if result != 0 {
		http.Error(w, "Failed to send character request", http.StatusInternalServerError)
		return
	}

	// Wait for character response
	_, err := network.ReadSpecificFromChannel(channel, gms.MSG_CHARAC_RES_NUM, 10*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read character response: %v", err), http.StatusRequestTimeout)
		return
	}

	// STEP 2: Send inventory request
	var invReqMessage gms.MSG_INVEN_REQ
	invReqMessage.Header = gms.GmsHeader{
		IKey:     gms.MSG_KEY,
		CMessage: gms.MSG_INVEN_REQ_NUM,
		UITime:   0,
	}

	copy(invReqMessage.Header.CGMName[:], "Go-GMEmulator")
	copy(invReqMessage.CCharacName[:], characterName)

	result = network.Send(invReqMessage, int(binary.Size(invReqMessage)), conn)
	if result != 0 {
		http.Error(w, "Failed to send inventory request", http.StatusInternalServerError)
		return
	}

	// var addItemDTO MSG_GM_ADD_ITEM_DTO
	// var addItemReq gms.MSG_GM_ADD_INVITEM
	// addItemReq.Header.IKey = gms.MSG_KEY
	// addItemReq.Header.CMessage = gms.MSG_GM_ADD_INVITEM_NUM
	// // editLevelMsg.ILevel = editLevelMsgDto.ILevel
}
