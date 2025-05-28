package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gm-emulator/crypto"
	"gm-emulator/gms"
	"gm-emulator/handlers"
	"gm-emulator/system"
	"net"
	"net/http"
	"os"
	"time"
)

type PacketHandler func([]byte)

// create a global jcrypto object
var jcrypto *crypto.JCrypto

var handlersMap = map[byte]PacketHandler{
	gms.MSG_SYSTEM_TIME_RES_NUM: handlers.HandleSystemTimeRes,
	// Add more handlers here...
}

func main() {
	addr := "51.222.8.230:9998"
	conn, err := connectToServer(addr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	system.GlobalSystem.DSConnection = conn
	defer conn.Close()

	jcrypto = crypto.NewJCrypto(4096)
	err2 := jcrypto.Init("D:\\Dev\\NewDawn9D\\Go-GMEmulator\\lump.dat")
	if err2 != nil {
		fmt.Printf("Warning: Could not load key file: %v\n", err2)
		os.Exit(1)
	}

	go startWebServer()
	go receiveLoop(conn)

	go func() {
		sendSystemTimeReq(conn, addr)
	}()

	// Block main forever (or use a better mechanism)
	select {}
}

func sendSystemTimeReq(conn net.Conn, addr string) {
	for {
		data := gms.MSG_SYSTEM_TIME_REQ{
			Header: gms.GmsHeader{
				IKey:     gms.MSG_KEY,
				CMessage: gms.MSG_SYSTEM_TIME_REQ_NUM,
				UITime:   0,
				CGMName:  [13]byte{'f', 'i', 'r', 'e', 'f', 'o', 'x', 0, 0, 0, 0, 0, 0},
			},
		}

		structBuf := new(bytes.Buffer)
		err := binary.Write(structBuf, binary.LittleEndian, data)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		buf := new(bytes.Buffer)
		err = binary.Write(buf, binary.LittleEndian, uint16(22))
		if err != nil {
			fmt.Println("Failed to write length prefix:", err)
			time.Sleep(10 * time.Second)
			continue
		}
		buf.Write(structBuf.Bytes())

		_, err = conn.Write(buf.Bytes())
		if err != nil {
			fmt.Println("Error sending data:", err)
			// Try to reconnect
			newConn, connErr := connectToServer(addr)
			if connErr != nil {
				fmt.Println("Reconnection failed:", connErr)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Println("Reconnected to server.")
			conn = newConn
			system.GlobalSystem.DSConnection = conn
		}

		time.Sleep(10 * time.Second)
	}
}

func connectToServer(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}

// Modified to handle errors and reconnect if needed

func receiveLoop(conn net.Conn) {
	const MAX_OF_RECV_BUFFER = 4096
	recvBuffer := make([]byte, MAX_OF_RECV_BUFFER)
	curStartPos := 0
	curEndPos := 0

	for {
		n, err := conn.Read(recvBuffer[curEndPos:])
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}
		curEndPos += n

		for {
			if curStartPos < curEndPos-2 {
				packetLen := binary.LittleEndian.Uint16(recvBuffer[curStartPos : curStartPos+2])
				if packetLen > MAX_OF_RECV_BUFFER {
					fmt.Println("[RecvBuffer] Exception: Packet Length Overflow")
					return
				}
				if curEndPos-curStartPos >= int(packetLen)+2 {
					packet := recvBuffer[curStartPos+2 : curStartPos+2+int(packetLen)]
					curStartPos += int(packetLen) + 2

					if len(packet) >= 5 {
						msgKey := int(binary.LittleEndian.Uint32(packet[0:4]))
						if msgKey == gms.MSG_KEY {
							cMessage := packet[4]
							if handler, ok := handlersMap[cMessage]; ok {
								handler(packet)
							} else {
								fmt.Printf("Unknown CMessage: %d\n", cMessage)
							}
						} else {
							fmt.Printf("MSG_KEY mismatch: got %d, expected %d\n", msgKey, gms.MSG_KEY)
						}
					} else {
						fmt.Println("Packet too short to parse MSG_KEY and CMessage")
					}
					fmt.Println("Packet received:", packet)
				} else {
					if curStartPos != 0 {
						copy(recvBuffer[0:], recvBuffer[curStartPos:curEndPos])
						curEndPos -= curStartPos
						curStartPos = 0
						for i := curEndPos; i < MAX_OF_RECV_BUFFER; i++ {
							recvBuffer[i] = 0
						}
					}
					break
				}
			} else {
				break
			}
		}
	}
}

func startWebServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := gms.MSG_GM_EDIT_LEVEL{
			Header: gms.GmsHeader{
				IKey:     gms.MSG_KEY,
				CMessage: gms.MSG_GM_EDIT_LEVEL_NUM,
			},
			ILevel: 200,
		}
		copy(msg.CCharacName[:], []byte("FirefoxTest"))

		ret := Send(msg, int(binary.Size(msg)))
		if ret == 0 {
			fmt.Fprintf(w, "Message sent successfully!")
		} else {
			fmt.Fprintf(w, "Failed to send message.")
		}
	})
	http.ListenAndServe(":8080", nil)
}

func Send(msg interface{}, size int) int {

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
	copy(header.CGMName[:], []byte("firefox"))

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

	fmt.Printf("sent at: time: %d, key: %d\n", header.UITime, header.UITime%100)

	n, err := system.GlobalSystem.DSConnection.Write(finalBuf.Bytes())
	fmt.Printf("len: %d\n", n)
	fmt.Printf("buffer: %v\n", finalBuf.Bytes())
	if err != nil {
		fmt.Println("Error sending data:", err)
		return -1
	}
	return 0
}
