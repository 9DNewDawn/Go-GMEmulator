package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gm-emulator/crypto"
	"gm-emulator/gms"
	"gm-emulator/system"
	"os"
	"time"
)

var jcrypto *crypto.JCrypto

func Send(msg interface{}, size int) int {
	// check if jcrypto is initialized
	if jcrypto == nil {
		fmt.Println("JCrypto not initialized, initializing now...")
		jcrypto = crypto.NewJCrypto(4096)
		err := jcrypto.Init("D:\\Dev\\NewDawn9D\\Go-GMEmulator\\lump.dat")
		if err != nil {
			fmt.Printf("Warning: Could not load key file: %v\n", err)
			os.Exit(1)
		}
	}

	if system.GlobalSystem == nil {
		fmt.Println("Time gap not set or GlobalSystem is nil, cannot send message")
		return -1
	}

	structBuf := new(bytes.Buffer)
	err := binary.Write(structBuf, binary.LittleEndian, msg)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return -1
	}
	buffer := structBuf.Bytes()
	if len(buffer) < binary.Size(gms.GmsHeader{}) {
		fmt.Println("Buffer too small for header")
		return -1
	}

	// Read header from buffer
	header := gms.GmsHeader{}
	headerBuf := bytes.NewReader(buffer[:binary.Size(gms.GmsHeader{})])
	err = binary.Read(headerBuf, binary.LittleEndian, &header)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
		return -1
	}

	now := time.Now().Unix()
	header.UITime = uint32(now) + uint32(system.GlobalSystem.TimeGapBetweenDS)

	// Write updated header back to buffer
	headerBufOut := new(bytes.Buffer)
	err = binary.Write(headerBufOut, binary.LittleEndian, &header)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return -1
	}
	copy(buffer[:binary.Size(gms.GmsHeader{})], headerBufOut.Bytes())

	// Prepend 2-byte length prefix (size)
	finalBuf := new(bytes.Buffer)
	err = binary.Write(finalBuf, binary.LittleEndian, uint16(size))
	if err != nil {
		fmt.Println("Failed to write length prefix:", err)
		return -1
	}
	finalBuf.Write(buffer[:size])

	if header.CMessage != 0 {
		jcrypto.Encryption(finalBuf.Bytes()[2+9:], uint8(header.UITime%100))
	}

	_, err = system.GlobalSystem.DSConnection.Write(finalBuf.Bytes())
	if err != nil {
		fmt.Println("Error sending data:", err)
		return -1
	}
	return 0
}
