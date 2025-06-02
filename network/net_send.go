package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gm-emulator/crypto"
	"gm-emulator/gms"
	"gm-emulator/system"
	"net"
	"os"
	"time"
)

var jcrypto *crypto.JCrypto

func Send(msg interface{}, size int, conn net.Conn) int {
	// check if jcrypto is initialized
	if jcrypto == nil {
		// fmt.Println("JCrypto not initialized, initializing now...")
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
	header.UITime = uint32(now) + uint32(system.GlobalSystem.GetTimeGapBetweenDSAsUint32())
	adjustedTime := uint32(now) + system.GlobalSystem.GetTimeGapBetweenDSAsUint32()

	// Add these debug lines:
	fmt.Printf("DEBUG TIME: now=%d, gap=%d, adjusted=%d\n", now, system.GlobalSystem.GetTimeGapBetweenDS(), adjustedTime)
	fmt.Printf("DEBUG TIME: successful UITime was 1748890605\n")
	fmt.Printf("DEBUG TIME: difference = %d seconds\n", int64(adjustedTime)-1748890605)

	adjustedTime = 1748890605 // Use the exact UITime from successful packet
	fmt.Printf("DEBUG: Using manual UITime override: %d\n", adjustedTime)

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

	_, err = conn.Write(finalBuf.Bytes())
	if err != nil {
		fmt.Println("Error sending data:", err)
		return -1
	}
	return 0
}

// DebugSend - debug version of Send function with detailed logging
func DebugSend(msg interface{}, size int, conn net.Conn) int {
	// Check if jcrypto is initialized (same as original)
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

	// Serialize the message
	structBuf := new(bytes.Buffer)
	err := binary.Write(structBuf, binary.LittleEndian, msg)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return -1
	}

	buffer := structBuf.Bytes()
	fmt.Printf("DEBUG: Original buffer length: %d\n", len(buffer))
	fmt.Printf("DEBUG: Original buffer (hex): %x\n", buffer)

	// Try to extract and display the message content
	if invReq, ok := msg.(gms.MSG_INVEN_REQ); ok {
		fmt.Printf("DEBUG: Original MSG_INVEN_REQ:\n")
		fmt.Printf("  Header.IKey: %d\n", invReq.Header.IKey)
		fmt.Printf("  Header.CMessage: %d\n", invReq.Header.CMessage)
		fmt.Printf("  Header.UITime: %d\n", invReq.Header.UITime)
		fmt.Printf("  Header.CGMName: %q\n", string(invReq.Header.CGMName[:]))
		fmt.Printf("  CCharacName: %q\n", string(invReq.CCharacName[:]))
		fmt.Printf("  CCharacName (hex): %x\n", invReq.CCharacName[:])
	}

	// Check buffer size
	headerSize := binary.Size(gms.GmsHeader{})
	if len(buffer) < headerSize {
		fmt.Printf("DEBUG: Buffer too small for header: %d < %d\n", len(buffer), headerSize)
		return -1
	}

	// Read header from buffer
	header := gms.GmsHeader{}
	headerBuf := bytes.NewReader(buffer[:headerSize])
	err = binary.Read(headerBuf, binary.LittleEndian, &header)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
		return -1
	}

	fmt.Printf("DEBUG: Extracted header:\n")
	fmt.Printf("  IKey: %d\n", header.IKey)
	fmt.Printf("  CMessage: %d\n", header.CMessage)
	fmt.Printf("  UITime: %d\n", header.UITime)
	fmt.Printf("  CGMName: %q\n", string(header.CGMName[:]))

	// Update UITime
	now := time.Now().Unix()
	header.UITime = uint32(now) + system.GlobalSystem.GetTimeGapBetweenDSAsUint32()

	fmt.Printf("DEBUG: Updated UITime: %d\n", header.UITime)

	// Write updated header back to buffer
	headerBufOut := new(bytes.Buffer)
	err = binary.Write(headerBufOut, binary.LittleEndian, &header)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return -1
	}

	// CRITICAL: Copy the updated header back to the original buffer
	copy(buffer[:headerSize], headerBufOut.Bytes())

	fmt.Printf("DEBUG: Buffer after header update (hex): %x\n", buffer)

	// Show what the character name looks like in the final buffer
	if len(buffer) > headerSize+13 {
		characNameBytes := buffer[headerSize : headerSize+13]
		fmt.Printf("DEBUG: Character name in final buffer: %q\n", string(characNameBytes))
		fmt.Printf("DEBUG: Character name in final buffer (hex): %x\n", characNameBytes)
	}

	// Prepend 2-byte length prefix
	finalBuf := new(bytes.Buffer)
	err = binary.Write(finalBuf, binary.LittleEndian, uint16(size))
	if err != nil {
		fmt.Println("Failed to write length prefix:", err)
		return -1
	}
	finalBuf.Write(buffer[:size])

	fmt.Printf("DEBUG: Final packet length: %d\n", finalBuf.Len())
	fmt.Printf("DEBUG: Final packet (hex): %x\n", finalBuf.Bytes())

	// Encrypt if needed (same as original)
	if header.CMessage != 0 {
		jcrypto.Encryption(finalBuf.Bytes()[2+9:], uint8(header.UITime%100))
		fmt.Printf("DEBUG: Packet after encryption (hex): %x\n", finalBuf.Bytes())
	}

	// Send the packet
	_, err = conn.Write(finalBuf.Bytes())
	if err != nil {
		fmt.Println("Error sending data:", err)
		return -1
	}

	fmt.Printf("DEBUG: Packet sent successfully\n")
	return 0
}
