package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	ApiRoutes "gm-emulator/api/domains/game/routes"
	"gm-emulator/crypto"
	"gm-emulator/gms"
	"gm-emulator/handlers"
	"gm-emulator/system"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
)

type PacketHandler func([]byte)

// create a global jcrypto object
var jcrypto *crypto.JCrypto

var handlersMap = map[byte]PacketHandler{
	gms.MSG_SYSTEM_TIME_RES_NUM: handlers.HandleSystemTimeRes,
	// Add more handlers here...
}

func initializeConnectionPool() error {
	const poolSize = 4
	for i := 0; i < poolSize; i++ {
		conn, err := connectToServer("51.222.8.230:9998")
		if err != nil {
			return fmt.Errorf("failed to create connection %d: %v", i, err)
		}

		// Add to pool
		system.GlobalSystem.DSConnectionPool <- conn
		fmt.Printf("Added connection %d to pool\n", i+1)
	}
	return nil
}

func main() {
	// initialize the conneciton pool
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
			},
		}

		copy(data.Header.CGMName[:], "Go-GMEmulator")
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
			newConn, connErr := connectToServer(addr)
			if connErr != nil {
				fmt.Println("Reconnection failed:", connErr)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Println("Reconnected to server.")
			conn = newConn
			system.GlobalSystem.DSConnection = conn
			// Restart the receive loop
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

	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("[%s]: '%s' has %d middlewares\n", method, route, len(middlewares))
		return nil
	})

	fmt.Println("Starting web server on :3000")
	http.ListenAndServe(":3000", r)
}

// Create a middleware that checks if we have a free connection in the system.GlobalSystem.DSConnectionPool. If it doesn't have one, it'll create a new one and add it to the pool.
func ConnectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Current connection pool size: %d\n", len(system.GlobalSystem.DSConnectionPool))

		var conn net.Conn
		select {
		case conn = <-system.GlobalSystem.DSConnectionPool:
			fmt.Println("Got connection from pool")
		default:
			// No connection available - refuse request
			fmt.Println("No connections available, refusing request")
			http.Error(w, "No connections available", http.StatusServiceUnavailable)
			return
		}

		// Pass connection in context
		ctx := context.WithValue(r.Context(), "conn", conn)

		// Return connection to pool after request
		defer func() {
			system.GlobalSystem.DSConnectionPool <- conn
			fmt.Println("Returned connection to pool")
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
