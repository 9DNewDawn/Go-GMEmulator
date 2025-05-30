package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gm-emulator/gms"
	"gm-emulator/system"
	"time"
)

// Example handler function
func HandleSystemTimeRes(packet []byte) {
	expectedSize := binary.Size(gms.MSG_SYSTEM_TIME_RES{})
	if len(packet) >= expectedSize {
		var res gms.MSG_SYSTEM_TIME_RES
		buf := bytes.NewReader(packet)
		err := binary.Read(buf, binary.LittleEndian, &res)
		if err != nil {
			fmt.Println("Failed to parse MSG_SYSTEM_TIME_RES:", err)
			return
		}
		fmt.Printf("Received MSG_SYSTEM_TIME_RES: ServerIndex=%d, Time=%d\n", res.UServerIndex, res.UITime)
		// time now
		now := time.Now().Unix()

		system.GlobalSystem.SetTimeGapBetweenDS(res.UITime - uint32(now))

	} else {
		fmt.Printf("Packet too short: got %d bytes, expected at least %d bytes\n", len(packet), expectedSize)
	}
}
