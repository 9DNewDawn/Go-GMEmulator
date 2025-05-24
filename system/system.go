package system

import "net"

var GlobalSystem = &System{}

type System struct {
	ServerTime       uint32
	TimeGapBetweenDS uint32
	DSConnection     net.Conn
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
