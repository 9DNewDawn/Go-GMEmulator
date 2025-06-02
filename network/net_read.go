package network

import (
	"encoding/binary"
	"fmt"
	"gm-emulator/gms"
	"net"
)

func Read(conn net.Conn) ([]byte, error) {
	const MAX_OF_RECV_BUFFER = 32768
	recvBuffer := make([]byte, MAX_OF_RECV_BUFFER)
	curStartPos := 0
	curEndPos := 0

	n, err := conn.Read(recvBuffer[curEndPos:])
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	curEndPos += n

	if curStartPos < curEndPos-2 {
		packetLen := binary.LittleEndian.Uint16(recvBuffer[curStartPos : curStartPos+2])
		if packetLen > MAX_OF_RECV_BUFFER {
			return nil, fmt.Errorf("[RecvBuffer] Exception: Packet Length Overflow")
		}
		if curEndPos-curStartPos >= int(packetLen)+2 {
			packet := make([]byte, packetLen)
			copy(packet, recvBuffer[curStartPos+2:curStartPos+2+int(packetLen)])

			if len(packet) >= 5 {
				msgKey := int(binary.LittleEndian.Uint32(packet[0:4]))
				if msgKey == gms.MSG_KEY {
					cMessage := packet[4]
					if handler, ok := handlersMap[cMessage]; ok {
						fmt.Printf("Received packet with CMessage: %d\n", cMessage)
						handler(packet)
						// Put this packet into a global channel, which will be split into
					} else {
						fmt.Printf("Unknown CMessage: %d\n", cMessage)
					}
				} else {
					fmt.Printf("MSG_KEY mismatch: got %d, expected %d\n", msgKey, gms.MSG_KEY)
				}
			} else {
				fmt.Println("Packet too short to parse MSG_KEY and CMessage")
			}

			return packet, nil
		}
	}

	return nil, fmt.Errorf("not enough data for a full packet")
}
