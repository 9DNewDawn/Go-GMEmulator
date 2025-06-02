package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	ApiRoutes "gm-emulator/api/domains/game/routes"
	"gm-emulator/gms"
	"gm-emulator/handlers"
	"gm-emulator/system"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
)

type PacketHandler func([]byte)

var handlersMap = map[byte]PacketHandler{
	gms.MSG_SYSTEM_TIME_RES_NUM: handlers.HandleSystemTimeRes,
}

var (
	connectionChannels = make(map[net.Conn]chan []byte)
	channelsMutex      = sync.RWMutex{}
)

func initializeConnectionPool() error {
	const poolSize = 4
	for i := 0; i < poolSize; i++ {
		conn, err := connectToServer("51.222.8.230:9998")
		if err != nil {
			return fmt.Errorf("failed to create connection %d: %v", i, err)
		}

		packetChannel := make(chan []byte, 100)

		channelsMutex.Lock()
		connectionChannels[conn] = packetChannel
		channelsMutex.Unlock()

		go receiveLoopForConnection(conn, i+1)
		go heartbeatLoopForConnection(conn, i+1)

		connWithChannel := &system.ConnectionWithChannel{
			Conn:    conn,
			Channel: packetChannel,
			ID:      i + 1,
		}

		system.GlobalSystem.DSConnectionPool <- connWithChannel
	}
	return nil
}

func heartbeatLoopForConnection(conn net.Conn, connID int) {
	sendSingleSystemTimeReq(conn)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := sendSingleSystemTimeReq(conn)
			if err != nil {
				return
			}
		}
	}
}

func sendSingleSystemTimeReq(conn net.Conn) error {
	data := gms.MSG_SYSTEM_TIME_REQ{
		Header: gms.GmsHeader{
			IKey:     gms.MSG_KEY,
			CMessage: gms.MSG_SYSTEM_TIME_REQ_NUM,
			UITime:   0,
		},
	}

	copy(data.Header.CGMName[:], "Go-GMEmulator")
	structBuf := new(bytes.Buffer)
	err := binary.Write(structBuf, binary.LittleEndian, data)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, uint16(22))
	if err != nil {
		return err
	}
	buf.Write(structBuf.Bytes())

	_, err = conn.Write(buf.Bytes())
	return err
}

func main() {
	err := initializeConnectionPool()
	if err != nil {
		fmt.Println("Error initializing connection pool:", err)
		return
	}

	addr := "51.222.8.230:9998"
	conn, err := connectToServer(addr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	system.GlobalSystem.DSConnection = conn
	defer conn.Close()

	go startWebServer()
	go receiveLoop(conn)
	go sendSystemTimeReq(conn, addr)

	select {}
}

func receiveLoopForConnection(conn net.Conn, connID int) {
	channelsMutex.RLock()
	packetChannel, exists := connectionChannels[conn]
	channelsMutex.RUnlock()

	if !exists {
		return
	}

	const MAX_OF_RECV_BUFFER = 32768
	recvBuffer := make([]byte, MAX_OF_RECV_BUFFER)
	curStartPos := 0
	curEndPos := 0

	defer func() {
		channelsMutex.Lock()
		delete(connectionChannels, conn)
		close(packetChannel)
		channelsMutex.Unlock()
	}()

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		n, err := conn.Read(recvBuffer[curEndPos:])
		if err != nil {
			return
		}

		curEndPos += n

		for {
			if curStartPos < curEndPos-2 {
				packetLen := binary.LittleEndian.Uint16(recvBuffer[curStartPos : curStartPos+2])
				if packetLen > MAX_OF_RECV_BUFFER {
					return
				}
				if curEndPos-curStartPos >= int(packetLen)+2 {
					packet := make([]byte, packetLen)
					copy(packet, recvBuffer[curStartPos+2:curStartPos+2+int(packetLen)])
					curStartPos += int(packetLen) + 2

					if len(packet) >= 5 {
						msgKey := int(binary.LittleEndian.Uint32(packet[0:4]))
						if msgKey == gms.MSG_KEY {
							cMessage := packet[4]

							if cMessage == gms.MSG_SYSTEM_TIME_RES_NUM {
								handlers.HandleSystemTimeRes(packet)
							} else {
								select {
								case packetChannel <- packet:
								default:
									// Channel full, drop packet
								}
							}
						}
					}
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

func sendSystemTimeReq(conn net.Conn, addr string) {
	for {
		data := gms.MSG_SYSTEM_TIME_REQ{
			Header: gms.GmsHeader{
				IKey:     gms.MSG_KEY,
				CMessage: gms.MSG_SYSTEM_TIME_REQ_NUM,
				UITime:   0,
			},
		}

		copy(data.Header.CGMName[:], "Go-GMEmulator")
		structBuf := new(bytes.Buffer)
		err := binary.Write(structBuf, binary.LittleEndian, data)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		buf := new(bytes.Buffer)
		err = binary.Write(buf, binary.LittleEndian, uint16(22))
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		buf.Write(structBuf.Bytes())

		_, err = conn.Write(buf.Bytes())
		if err != nil {
			newConn, connErr := connectToServer(addr)
			if connErr != nil {
				time.Sleep(10 * time.Second)
				continue
			}
			conn = newConn
			system.GlobalSystem.DSConnection = conn
			go receiveLoop(conn)
		}

		time.Sleep(10 * time.Second)
	}
}

func connectToServer(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}

func receiveLoop(conn net.Conn) {
	const MAX_OF_RECV_BUFFER = 32768
	recvBuffer := make([]byte, MAX_OF_RECV_BUFFER)
	curStartPos := 0
	curEndPos := 0

	for {
		n, err := conn.Read(recvBuffer[curEndPos:])
		if err != nil {
			return
		}
		curEndPos += n

		for {
			if curStartPos < curEndPos-2 {
				packetLen := binary.LittleEndian.Uint16(recvBuffer[curStartPos : curStartPos+2])
				if packetLen > MAX_OF_RECV_BUFFER {
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
							}
						}
					}
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
	r := chi.NewRouter()

	r.Use(ConnectionMiddleware)
	r.Route("/api", func(r chi.Router) {
		ApiRoutes.GetGameRoutes(r)
	})

	fmt.Println("Starting web server on :3000")
	http.ListenAndServe(":3000", r)
}

func ConnectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var connWithChannel *system.ConnectionWithChannel
		select {
		case connWithChannel = <-system.GlobalSystem.DSConnectionPool:
		default:
			http.Error(w, "No connections available", http.StatusServiceUnavailable)
			return
		}

		ctx := context.WithValue(r.Context(), "conn", connWithChannel.Conn)
		ctx = context.WithValue(ctx, "channel", connWithChannel.Channel)
		ctx = context.WithValue(ctx, "connID", connWithChannel.ID)

		next.ServeHTTP(w, r.WithContext(ctx))

		system.GlobalSystem.DSConnectionPool <- connWithChannel
	})
}
