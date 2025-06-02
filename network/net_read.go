package network

import (
	"encoding/binary"
	"fmt"
	"gm-emulator/gms"
	"time"
)

// ReadFromChannel reads a packet from the connection's dedicated channel
func ReadFromChannel(channel chan []byte, timeout time.Duration) ([]byte, error) {
	select {
	case packet := <-channel:
		return packet, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for packet")
	}
}

// ReadSpecificFromChannel reads and waits for a specific CMessage type from the channel
func ReadSpecificFromChannel(channel chan []byte, expectedCMessage byte, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		remainingTime := time.Until(deadline)
		if remainingTime <= 0 {
			break
		}

		select {
		case packet := <-channel:
			if len(packet) >= 5 {
				msgKey := int(binary.LittleEndian.Uint32(packet[0:4]))
				if msgKey != gms.MSG_KEY {
					continue
				}

				cMessage := packet[4]
				if cMessage == expectedCMessage {
					return packet, nil
				}
			}
		case <-time.After(remainingTime):
			break
		}
	}

	return nil, fmt.Errorf("timeout: did not receive expected CMessage %d within %v", expectedCMessage, timeout)
}
