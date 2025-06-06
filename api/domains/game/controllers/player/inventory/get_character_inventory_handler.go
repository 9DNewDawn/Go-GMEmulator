package GameControllers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"gm-emulator/gms"
	"gm-emulator/network"
	"net"
	"net/http"
	"time"
)

// Item represents a single inventory item
type InventoryItem struct {
	Category uint32 // 64190602 = item signature/category
	Type     uint8  // Item type (02, 03, etc.)
	Subtype  uint8  // Item subtype (34, 00, etc.)
	Slot     uint8  // Slot number (01, 02, 03, etc.)
	Unknown1 uint8  // Usually 00
	Count    uint8  // Item count (01, 09, etc.)
	Unknown2 uint8  // Usually 00
	ItemID   uint32 // Item identifier
	Unknown3 uint32 // Additional data
}

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

	// STEP 3: Read and reconstruct inventory data from multiple chunks
	var totalInventoryData []byte
	chunkCount := 0
	const udpPacketCutSize = 300 // Based on C++ code
	const maxChunks = 100

	for chunkCount < maxChunks {
		packet, err := network.ReadSpecificFromChannel(channel, gms.MSG_INVEN_RES_NUM, 10*time.Second)
		if err != nil {
			if chunkCount == 0 {
				http.Error(w, fmt.Sprintf("Failed to read first inventory response: %v", err), http.StatusRequestTimeout)
				return
			}
			break
		}

		dataOffset := 0
		headerSize := 22
		charNameSize := 13

		if len(packet) < headerSize+charNameSize+4+1 {
			break
		}
		dataOffset += headerSize
		dataOffset += charNameSize

		iSize := int(binary.LittleEndian.Uint32(packet[dataOffset : dataOffset+4]))
		dataOffset += 4
		// cNum := int(packet[dataOffset])
		dataOffset += 1

		if iSize > 0 && len(packet) >= dataOffset+iSize {
			inventoryChunk := packet[dataOffset : dataOffset+iSize]
			totalInventoryData = append(totalInventoryData, inventoryChunk...)
		}
		chunkCount++

		// Last chunk: iSize < udpPacketCutSize
		if iSize < udpPacketCutSize {
			break
		}
	}

	// At this point, totalInventoryData should be the full _CHARAC_INVENTORY struct
	// The first 3888 bytes are the main inventory (cInventory)
	const invenSize = 3888
	// if len(totalInventoryData) < invenSize {
	// 	http.Error(w, "Inventory data incomplete", http.StatusInternalServerError)
	// 	return
	// }
	inventoryBytes := totalInventoryData[:invenSize]

	// Parse inventory items from cInventory
	items := parseInventoryItems(inventoryBytes)

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	type response struct {
		Items []InventoryItem `json:"items"`
	}
	_ = json.NewEncoder(w).Encode(response{Items: items})
}

func parseInventoryItems(data []byte) []InventoryItem {
	var items []InventoryItem

	// Based on packet analysis, each item starts with signature 0x64190602
	// From your hex dump, I can see items at:
	// Offset 0x60: 64 19 06 02 02 34 01 00 01 00 c2 2f 85 3b 00 01...
	// Offset 0x420: 64 19 06 02 03 00 0d 00 13 00 af 2e b0 ef 20 01...
	// Offset 0x4e0: 64 19 06 02 02 34 02 00 09 00 ae 01 00 00 00 01...

	i := 0
	for i < len(data)-16 { // Need at least 16 bytes for an item
		// Look for the item signature 0x64190602
		if i+4 <= len(data) {
			signature := binary.LittleEndian.Uint32(data[i : i+4])
			if signature == 0x02061964 { // 64190602 in little endian
				// Found an item, parse it
				if i+16 <= len(data) {
					item := InventoryItem{
						Category: signature,
						Type:     data[i+4],
						Subtype:  data[i+5],
						Slot:     data[i+6],
						Unknown1: data[i+7],
						Count:    data[i+8],
						Unknown2: data[i+9],
					}

					if i+14 <= len(data) {
						item.ItemID = binary.LittleEndian.Uint32(data[i+10 : i+14])
					}
					if i+18 <= len(data) {
						item.Unknown3 = binary.LittleEndian.Uint32(data[i+14 : i+18])
					}

					items = append(items, item)

					// Each item appears to be about 96 bytes based on your packet spacing
					// Move to next potential item location
					i += 96
				} else {
					break
				}
			} else {
				i++
			}
		} else {
			break
		}
	}

	return items
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
