package GameControllers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
	"time"
)

func GetCharacterInventoryHandler(w http.ResponseWriter, r *http.Request) {
	// Get connection and channel from context
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

	connID, _ := r.Context().Value("connID").(int)

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

	// STEP 3: Read multiple inventory response packets
	var inventoryChunks []gms.MSG_INVEN_RES
	var totalInventoryData []byte
	expectedChunk := 0
	maxChunks := 100

	for chunkCount := 0; chunkCount < maxChunks; chunkCount++ {
		packet, err := network.ReadSpecificFromChannel(channel, gms.MSG_INVEN_RES_NUM, 10*time.Second)
		if err != nil {
			if chunkCount == 0 {
				http.Error(w, fmt.Sprintf("Failed to read first inventory response: %v", err), http.StatusRequestTimeout)
				return
			}
			break
		}

		expectedSize := binary.Size(gms.MSG_INVEN_RES{})
		if len(packet) < expectedSize {
			http.Error(w, fmt.Sprintf("Inventory response %d too short: got %d, expected at least %d", chunkCount, len(packet), expectedSize), http.StatusInternalServerError)
			return
		}

		var invResMessage gms.MSG_INVEN_RES
		buf := bytes.NewReader(packet)
		err = binary.Read(buf, binary.LittleEndian, &invResMessage)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing inventory response %d: %v", chunkCount, err), http.StatusInternalServerError)
			return
		}

		if int(invResMessage.CNum) != expectedChunk {
			fmt.Printf("Connection %d: Warning - expected chunk %d but got %d\n", connID, expectedChunk, invResMessage.CNum)
		}

		if invResMessage.ISize > 0 && int(invResMessage.ISize) <= len(invResMessage.PInvData) {
			totalInventoryData = append(totalInventoryData, invResMessage.PInvData[:invResMessage.ISize]...)
		}

		inventoryChunks = append(inventoryChunks, invResMessage)
		expectedChunk++

		if invResMessage.ISize < 1024 {
			break
		}
	}

	if len(inventoryChunks) == 0 {
		http.Error(w, "No inventory data received", http.StatusInternalServerError)
		return
	}

	// Build response
	firstChunk := inventoryChunks[0]
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	charName := string(bytes.Trim(firstChunk.CCharacName[:], "\x00"))

	fmt.Fprintf(w, "Character: %s\n", charName)
	fmt.Fprintf(w, "Total Chunks Received: %d\n", len(inventoryChunks))
	fmt.Fprintf(w, "Total Inventory Data Size: %d bytes\n", len(totalInventoryData))

	for i, chunk := range inventoryChunks {
		fmt.Fprintf(w, "Chunk %d: cNum=%d, iSize=%d\n", i, chunk.CNum, chunk.ISize)
	}

	maxDisplay := 100
	if len(totalInventoryData) < maxDisplay {
		maxDisplay = len(totalInventoryData)
	}
	fmt.Fprintf(w, "Raw Inventory Data: %x\n", totalInventoryData[:maxDisplay])
}
