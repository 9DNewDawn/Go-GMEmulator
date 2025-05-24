package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "51.222.8.230:9998")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	// Create a Data struct
	data := MSG_SYSTEM_TIME_REQ{
		Header: GmsHeader{
			IKey:     MSG_KEY,
			CMessage: MSG_SYSTEM_TIME_REQ_NUM,
			UITime:   0,
			CGMName:  [15]byte{'f', 'i', 'r', 'e', 'f', 'o', 'x', 0, 0, 0, 0, 0, 0},
		},
	}

	// Serialize the struct first to get its size
	structBuf := new(bytes.Buffer)
	err = binary.Write(structBuf, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return
	}

	// Create final buffer with length prefix
	buf := new(bytes.Buffer)
	// Write length prefix (16 00 = 22 in little endian)
	err = binary.Write(buf, binary.LittleEndian, uint16(22))
	if err != nil {
		fmt.Println("Failed to write length prefix:", err)
		return
	}
	// Write the struct data
	buf.Write(structBuf.Bytes())
	if err != nil {
		fmt.Println("binary.Write failed:", err)
		return
	}

	fmt.Printf("Serialized data length: %d\n", len(buf.Bytes()))
	fmt.Println("Serialized data:", buf.Bytes())

	// Send the serialized data
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}

	fmt.Println("Data sent successfully!")

	// Read the response
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Println("Response received:", response[:n])
}
