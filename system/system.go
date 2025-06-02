package system

import "net"

var GlobalSystem = &System{
	DSConnectionPool: make(chan net.Conn, 4),
}

type System struct {
	ServerTime       uint32
	TimeGapBetweenDS uint32
	DSConnection     net.Conn
	DSConnectionPool chan net.Conn // Channel pool for DSConnections
}

func (s *System) GetServerTime() uint32 {
	return s.ServerTime
}

func (s *System) SetServerTime(time uint32) {
	s.ServerTime = time
}

func (s *System) GetTimeGapBetweenDS() uint32 {
	return s.TimeGapBetweenDS
}

func (s *System) SetTimeGapBetweenDS(time uint32) {
	s.TimeGapBetweenDS = time
}

// Add a DSConnection to the pool
func (s *System) AddDSConnection(conn net.Conn) {
	s.DSConnectionPool <- conn
}

// Get a DSConnection from the pool
func (s *System) GetDSConnection() net.Conn {
	return <-s.DSConnectionPool
}
