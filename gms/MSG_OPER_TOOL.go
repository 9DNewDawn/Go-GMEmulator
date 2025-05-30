package gms

type GmsHeader struct {
	IKey     int32    // 4 bytes
	CMessage uint8    // 1 byte
	UITime   uint32   // 4 bytes
	CGMName  [13]byte // 13 bytes
}

type MSG_SYSTEM_TIME_REQ struct {
	Header GmsHeader
}

type MSG_SYSTEM_TIME_RES struct {
	Header       GmsHeader
	UITime       uint32
	UServerIndex uint16 // Server index
}

type MSG_GM_ADD_INVGOLD struct {
	Header      GmsHeader
	CCharacName [13]byte // UniqueUserID
	IGold       int32
}

type MSG_GM_EDIT_LEVEL struct {
	Header      GmsHeader
	CCharacName [13]byte // UniqueUserID
	ILevel      int32
}

type MSG_GM_EDIT_VITAL struct {
	Header      GmsHeader
	CCharacName [13]byte // UniqueUserID
	SVital      int16
	UIHP        uint32 // Current HP
}

type MSG_GM_EDIT_GMCLASS struct {
	Header           GmsHeader
	CCharacName      [13]byte // UniqueUserID
	CClass           byte
	IBlockingEndTime int32
}

type MSG_INVEN_REQ struct {
	Header      GmsHeader
	CCharacName [13]byte
}

const MSG_KEY = 1003
const MSG_SYSTEM_TIME_REQ_NUM = 11
const MSG_SYSTEM_TIME_RES_NUM = 12
const MSG_GM_ADD_INVGOLD_NUM = 77
const MSG_GM_EDIT_LEVEL_NUM = 61
const MSG_GM_EDIT_VITAL_NUM = 62
const MSG_GM_EDIT_GMCLASS_NUM = 70
const MSG_INVEN_REQ_NUM = 46
const MSG_INVEN_RES_NUM = 47
