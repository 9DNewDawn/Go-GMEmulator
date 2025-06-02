package handlers

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

// DebugPacketMonitor - shows all packets received on a connection for debugging
func DebugPacketMonitor(w http.ResponseWriter, r *http.Request) {
	channel, ok := r.Context().Value("channel").(chan []byte)
	if !ok {
		http.Error(w, "No channel available", http.StatusInternalServerError)
		return
	}

	connID, _ := r.Context().Value("connID").(int)
	fmt.Printf("=== PACKET MONITOR for Connection %d ===\n", connID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "Monitoring packets on connection %d...\n", connID)
	fmt.Fprintf(w, "Current channel buffer: %d packets\n\n", len(channel))

	// Monitor packets for 30 seconds
	timeout := time.After(30 * time.Second)
	packetCount := 0

	for {
		select {
		case packet := <-channel:
			packetCount++
			if len(packet) >= 5 {
				cMessage := packet[4]
				fmt.Printf("Connection %d MONITOR: Packet %d - CMessage=%d, Length=%d\n", connID, packetCount, cMessage, len(packet))
				fmt.Fprintf(w, "Packet %d: CMessage=%d, Length=%d bytes\n", packetCount, cMessage, len(packet))

				if len(packet) >= 10 {
					fmt.Fprintf(w, "  Hex dump (first 10 bytes): %x\n", packet[:10])
				}
				fmt.Fprintf(w, "\n")
			} else {
				fmt.Printf("Connection %d MONITOR: Packet %d - Too short (%d bytes)\n", connID, packetCount, len(packet))
				fmt.Fprintf(w, "Packet %d: Too short (%d bytes)\n", packetCount, len(packet))
			}

		case <-timeout:
			fmt.Printf("Connection %d MONITOR: Timeout after 30 seconds\n", connID)
			fmt.Fprintf(w, "\nMonitoring complete. Total packets received: %d\n", packetCount)
			return

		case <-time.After(5 * time.Second):
			if packetCount == 0 {
				fmt.Printf("Connection %d MONITOR: No packets received in 5 seconds\n", connID)
				fmt.Fprintf(w, "No packets received in 5 seconds...\n")
			}
		}
	}
}

// TestSystemTimeHandler - sends a system time request and waits for response to test if requests work
func TestSystemTimeHandler(w http.ResponseWriter, r *http.Request) {
	conn, ok := r.Context().Value("conn").(net.Conn)
	if !ok {
		http.Error(w, "No connection available", http.StatusInternalServerError)
		return
	}

	_, ok = r.Context().Value("channel").(chan []byte)
	if !ok {
		http.Error(w, "No channel available", http.StatusInternalServerError)
		return
	}

	connID, _ := r.Context().Value("connID").(int)
	fmt.Printf("=== TESTING SYSTEM TIME REQUEST on Connection %d ===\n", connID)

	// Build system time request
	data := gms.MSG_SYSTEM_TIME_REQ{
		Header: gms.GmsHeader{
			IKey:     gms.MSG_KEY,
			CMessage: gms.MSG_SYSTEM_TIME_REQ_NUM,
			UITime:   0,
		},
	}
	copy(data.Header.CGMName[:], "Go-GMEmulator")

	fmt.Printf("Connection %d: Sending manual system time request...\n", connID)
	result := network.Send(data, int(binary.Size(data)), conn)
	if result != 0 {
		http.Error(w, "Failed to send system time request", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Connection %d: Request sent, waiting for response...\n", connID)

	// This SHOULD receive a CMessage 12 response
	// But since we auto-handle those, we need to temporarily modify the receive loop
	// OR we can just wait and see if the heartbeat response comes back

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "System time request sent on connection %d\n", connID)
	fmt.Fprintf(w, "Check the console logs for the response.\n")
	fmt.Fprintf(w, "You should see 'Connection %d: Auto-handling heartbeat response (CMessage 12)' soon.\n", connID)
}

func EnhancedPacketMonitor(w http.ResponseWriter, r *http.Request) {
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
	fmt.Printf("=== ENHANCED MONITOR: Sending inventory request and monitoring ALL responses ===\n")

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "Enhanced packet monitor for connection %d\n", connID)
	fmt.Fprintf(w, "Sending inventory request for 'FirefoxUpper'...\n\n")

	// Send inventory request
	var invReqMessage gms.MSG_INVEN_REQ
	invReqMessage.Header = gms.GmsHeader{
		IKey:     gms.MSG_KEY,
		CMessage: gms.MSG_INVEN_REQ_NUM,
		UITime:   0,
	}
	copy(invReqMessage.Header.CGMName[:], "Go-GMEmulator")
	copy(invReqMessage.CCharacName[:], "FirefoxUpper")

	result := network.Send(invReqMessage, int(binary.Size(invReqMessage)), conn)
	if result != 0 {
		fmt.Fprintf(w, "ERROR: Failed to send inventory request\n")
		return
	}

	fmt.Printf("Connection %d: Inventory request sent, monitoring for 30 seconds...\n", connID)
	fmt.Fprintf(w, "Request sent! Monitoring for responses...\n\n")

	// Monitor ALL packets for 30 seconds
	timeout := time.After(30 * time.Second)
	packetCount := 0

	for {
		select {
		case packet := <-channel:
			packetCount++
			if len(packet) >= 5 {
				cMessage := packet[4]

				fmt.Printf("Connection %d MONITOR: Packet %d - CMessage=%d, Length=%d\n", connID, packetCount, cMessage, len(packet))
				fmt.Fprintf(w, "=== PACKET %d ===\n", packetCount)
				fmt.Fprintf(w, "CMessage: %d\n", cMessage)
				fmt.Fprintf(w, "Length: %d bytes\n", len(packet))

				// Show packet details
				if len(packet) >= 10 {
					fmt.Fprintf(w, "Hex dump (first 20 bytes): %x\n", packet[:min(20, len(packet))])
				}

				// Try to parse as different message types
				if cMessage == gms.MSG_INVEN_RES_NUM {
					fmt.Fprintf(w, "*** THIS IS THE INVENTORY RESPONSE WE'RE LOOKING FOR! ***\n")

					// Try to parse it
					if len(packet) >= binary.Size(gms.MSG_INVEN_RES{}) {
						var invRes gms.MSG_INVEN_RES
						buf := bytes.NewReader(packet)
						err := binary.Read(buf, binary.LittleEndian, &invRes)
						if err == nil {
							charName := string(bytes.Trim(invRes.CCharacName[:], "\x00"))
							fmt.Fprintf(w, "Character: %s\n", charName)
							fmt.Fprintf(w, "Chunk Number: %d\n", invRes.CNum)
							fmt.Fprintf(w, "Chunk Size: %d\n", invRes.ISize)
						} else {
							fmt.Fprintf(w, "Failed to parse inventory response: %v\n", err)
						}
					}
				} else {
					fmt.Fprintf(w, "Unknown CMessage type: %d\n", cMessage)

					// Show the raw header
					if len(packet) >= binary.Size(gms.GmsHeader{}) {
						var header gms.GmsHeader
						buf := bytes.NewReader(packet)
						err := binary.Read(buf, binary.LittleEndian, &header)
						if err == nil {
							gmName := string(bytes.Trim(header.CGMName[:], "\x00"))
							fmt.Fprintf(w, "Header - IKey: %d, UITime: %d, GMName: '%s'\n",
								header.IKey, header.UITime, gmName)
						}
					}
				}

				fmt.Fprintf(w, "\n")

			} else {
				fmt.Printf("Connection %d MONITOR: Packet %d - Too short (%d bytes)\n", connID, packetCount, len(packet))
				fmt.Fprintf(w, "Packet %d: Too short (%d bytes)\n", packetCount, len(packet))
			}

		case <-timeout:
			fmt.Printf("Connection %d MONITOR: Timeout after 30 seconds\n", connID)
			fmt.Fprintf(w, "\nMonitoring complete.\n")
			fmt.Fprintf(w, "Total non-heartbeat packets received: %d\n", packetCount)
			if packetCount == 0 {
				fmt.Fprintf(w, "\nNO RESPONSES RECEIVED!\n")
				fmt.Fprintf(w, "This means the server either:\n")
				fmt.Fprintf(w, "1. Character 'FirefoxUpper' doesn't exist\n")
				fmt.Fprintf(w, "2. Server requires authentication first\n")
				fmt.Fprintf(w, "3. Server sends error responses that are being filtered\n")
				fmt.Fprintf(w, "4. Request format is incorrect\n")
			}
			return

		case <-time.After(5 * time.Second):
			if packetCount == 0 {
				fmt.Printf("Connection %d MONITOR: Still no non-heartbeat packets...\n", connID)
				fmt.Fprintf(w, "Still waiting for responses...\n")
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func DebugInventoryHandler(w http.ResponseWriter, r *http.Request) {
	conn, ok := r.Context().Value("conn").(net.Conn)
	if !ok {
		http.Error(w, "No connection available", http.StatusInternalServerError)
		return
	}

	connID, _ := r.Context().Value("connID").(int)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "=== DEBUG INVENTORY REQUEST ===\n")
	fmt.Fprintf(w, "Using connection %d\n\n", connID)

	// Test with a simple character name first
	characterName := "Zakari[Shao]"

	fmt.Printf("=== DEBUG INVENTORY REQUEST for '%s' ===\n", characterName)
	fmt.Fprintf(w, "Testing with character name: '%s'\n\n", characterName)

	// Build the inventory request message
	var invReqMessage gms.MSG_INVEN_REQ

	// Initialize everything to zero first
	invReqMessage = gms.MSG_INVEN_REQ{}

	// Set header
	invReqMessage.Header.IKey = gms.MSG_KEY
	invReqMessage.Header.CMessage = gms.MSG_INVEN_REQ_NUM
	invReqMessage.Header.UITime = 0 // Will be set by network.Send

	// Copy strings carefully
	copy(invReqMessage.Header.CGMName[:], "firefox")
	copy(invReqMessage.CCharacName[:], characterName)

	// Debug: Show what we built before sending
	fmt.Printf("BEFORE SEND - MSG_INVEN_REQ structure:\n")
	fmt.Printf("  Header.IKey: %d\n", invReqMessage.Header.IKey)
	fmt.Printf("  Header.CMessage: %d\n", invReqMessage.Header.CMessage)
	fmt.Printf("  Header.UITime: %d\n", invReqMessage.Header.UITime)
	fmt.Printf("  Header.CGMName: %q\n", string(invReqMessage.Header.CGMName[:]))
	fmt.Printf("  CCharacName: %q\n", string(invReqMessage.CCharacName[:]))
	fmt.Printf("  CCharacName (hex): %x\n", invReqMessage.CCharacName[:])

	fmt.Fprintf(w, "Message structure before sending:\n")
	fmt.Fprintf(w, "  CMessage: %d\n", invReqMessage.Header.CMessage)
	fmt.Fprintf(w, "  Character Name: %q\n", string(invReqMessage.CCharacName[:]))
	fmt.Fprintf(w, "  GM Name: %q\n", string(invReqMessage.Header.CGMName[:]))
	fmt.Fprintf(w, "\nCheck console for detailed hex dumps...\n")

	// Use debug send function
	result := network.DebugSend(invReqMessage, int(binary.Size(invReqMessage)), conn)
	if result != 0 {
		fmt.Fprintf(w, "ERROR: Failed to send inventory request\n")
		return
	}

	fmt.Fprintf(w, "\nDebug send completed! Check console logs for details.\n")
}
