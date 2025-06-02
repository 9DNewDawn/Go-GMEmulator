package system

import (
	"net"
	"sync"
)

type ConnectionWithChannel struct {
	Conn    net.Conn
	Channel chan []byte
	ID      int
}

type GlobalSystemType struct {
	DSConnection     net.Conn
	DSConnectionPool chan *ConnectionWithChannel
	timeGapBetweenDS int64
	timeGapMutex     sync.RWMutex
}

var GlobalSystem = &GlobalSystemType{
	DSConnectionPool: make(chan *ConnectionWithChannel, 4),
}

func (gs *GlobalSystemType) SetTimeGapBetweenDS(gap uint32) {
	gs.timeGapMutex.Lock()
	defer gs.timeGapMutex.Unlock()
	gs.timeGapBetweenDS = int64(gap)
}

func (gs *GlobalSystemType) GetTimeGapBetweenDS() int64 {
	gs.timeGapMutex.RLock()
	defer gs.timeGapMutex.RUnlock()
	return gs.timeGapBetweenDS
}

func (gs *GlobalSystemType) GetTimeGapBetweenDSAsUint32() uint32 {
	gs.timeGapMutex.RLock()
	defer gs.timeGapMutex.RUnlock()
	return uint32(gs.timeGapBetweenDS)
}
