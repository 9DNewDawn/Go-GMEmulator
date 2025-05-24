package main

type GmsHeader struct {
	IKey     int32    // 4 bytes
	CMessage uint8    // 1 byte
	UITime   uint32   // 4 bytes
	CGMName  [15]byte // 13 bytes
}

type MSG_SYSTEM_TIME_REQ struct {
	Header GmsHeader
}

const MSG_KEY = 1003
const MSG_SYSTEM_TIME_REQ_NUM = 11
