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
			CGMName:  [13]byte{'f', 'i', 'r', 'e', 'f', 'o', 'x', 0, 0, 0, 0, 0, 0},
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
	// Buffer management variables (simulate the C++ logic)
	const MAX_OF_RECV_BUFFER = 4096
	recvBuffer := make([]byte, MAX_OF_RECV_BUFFER)
	curStartPos := 0
	curEndPos := 0

	// Read data into buffer
	n, err := conn.Read(recvBuffer[curEndPos:])
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	curEndPos += n

	// Try to extract a packet
	var packetLen uint16
	if curStartPos < curEndPos-2 {
		packetLen = binary.LittleEndian.Uint16(recvBuffer[curStartPos : curStartPos+2])
		if packetLen > MAX_OF_RECV_BUFFER {
			fmt.Println("[RecvBuffer] Exception: Packet Length Overflow")
			return
		}
		if curEndPos-curStartPos >= int(packetLen)+2 {
			packet := recvBuffer[curStartPos+2 : curStartPos+2+int(packetLen)]
			curStartPos += int(packetLen) + 2

			// Parse MSG_KEY and CMessage
			if len(packet) >= 5 {
				msgKey := int(binary.LittleEndian.Uint32(packet[0:4]))
				if msgKey == MSG_KEY {
					cMessage := packet[4]
					switch cMessage {
					// Add your case handlers here
					case MSG_SYSTEM_TIME_RES_NUM:
						// Cast packet to MSG_SYSTEM_TIME_RES
						expectedSize := binary.Size(MSG_SYSTEM_TIME_RES{})
						if len(packet) >= binary.Size(MSG_SYSTEM_TIME_RES{}) {
							var res MSG_SYSTEM_TIME_RES
							buf := bytes.NewReader(packet)
							err := binary.Read(buf, binary.LittleEndian, &res)
							if err != nil {
								fmt.Println("Failed to parse MSG_SYSTEM_TIME_RES:", err)
								break
							}
							fmt.Printf("Received MSG_SYSTEM_TIME_RES: ServerIndex=%d, Time=%d\n", res.UServerIndex, res.UITime)
							// You can add logic here to compare with local time and set time gap if needed
						} else {
							fmt.Printf("Packet too short: got %d bytes, expected at least %d bytes\n", len(packet), expectedSize)
						}
						break
					default:
						fmt.Printf("Unknown CMessage: %d\n", cMessage)
					}
				} else {
					fmt.Printf("MSG_KEY mismatch: got %d, expected %d\n", msgKey, MSG_KEY)
				}
			} else {
				fmt.Println("Packet too short to parse MSG_KEY and CMessage")
			}
			fmt.Println("Packet received:", packet)
			return
		}
		// Not enough data for a full packet, move remaining data to start
		if curStartPos != 0 {
			fmt.Println("Partial packet received, shifting buffer")
			copy(recvBuffer[0:], recvBuffer[curStartPos:curEndPos])
			curEndPos -= curStartPos
			curStartPos = 0
			for i := curEndPos; i < MAX_OF_RECV_BUFFER; i++ {
				recvBuffer[i] = 0
			}
		} else {
			fmt.Println("Partial packet received, waiting for more data")
		}
	} else {
		fmt.Println("Not enough data to determine packet length")
	}
}
