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

type MSG_INVEN_RES struct {
	Header      GmsHeader
	CCharacName [13]byte
	ISize       uint32 // Change to uint32 instead of int32
	CNum        byte
	PInvData    [300]byte
}

type MSG_CHARAC_REQ struct {
	Header      GmsHeader
	CCharacName [13]byte
}

type CHARAC_BASIC struct {
	IUniqueID     int32
	CAccount      [18]byte // en_max_lil+1, assuming en_max_lil=17
	CChrName      [13]byte
	CChrNic       [13]byte
	CSex          byte
	CParty        byte
	CGamete       [13]byte
	CHair         byte
	CFace         byte
	CLuck         byte
	CClass        byte
	CClassGrade   byte
	IContribution int32
	CGMCheck      byte
	DwPlayTime    uint32
	UcChangeName  byte
	// CharacCreateDate omitted (ifdef)
}

type CHARAC_CUR_BASIC struct {
	SZone               int16
	SY                  int16
	FX                  float32
	FZ                  float32
	SLifePower          int16
	SForcePower         int16
	SConcentrationPower int16
	CRespawnServerNo    byte
	CRespawnPosName     [13]byte
	FRespawnPosX        float32
	FRespawnPosZ        float32
}

type CHARAC_LEVEL struct {
	SMaxLifePower          int16
	SMaxForcePower         int16
	SMaxConcentrationPower int16
	SConstitution          int16
	SZen                   int16
	SIntelligence          int16
	SDexterity             int16
	SStr                   int16
	SLeftPoint             int16
	SMasteryPoint          int16
}

type CHARAC_STATE struct {
	SInnerLevel        int16
	UIJin              uint32
	IGong              int32
	SRetribution       int16
	IHonor             int32
	SShowdown          int16
	USFatigue          uint16
	SWoundValue        int16
	SInsideWoundValue  int16
	SFuryParameter     int16
	SLevelUpGameSecond int32
	IORIndex           int32
	SPeaceMode         int16
	IMuteTime          int32
	IHonorGaveDate     int32
	IHonorTakeDate     int32
	IHiding            int32
	IBlockingEndDate   int32
	SPkPrevDeadMode    int16
	SPkDeadCount       int16
	SPkKillCount       int16
	SMonsterKill       int16
	SPartyIndex        int16
	SPartySlotNo       int16
	// Omitted conditional fields
}

type MSG_CHARAC_RES struct {
	Header         GmsHeader
	CChrName       [13]byte
	CharacBasic    CHARAC_BASIC
	CharacCurBasic CHARAC_CUR_BASIC
	CharacLevel    CHARAC_LEVEL
	CharacState    CHARAC_STATE
}

type MSG_GM_ADD_INVITEM struct {
	Header       GmsHeader
	CCharacName  [13]byte
	CFirstType   uint8
	CSecondType  uint8
	SItemID      int16
	UCItemCount  uint8
	USDurability uint16
	UCSlotCount  uint8
	UCInchant    uint8
	// Conditional fields for _PD_GM_ADDITEM_MODIFY_
	// CCashCheck   int8   // Use int8 for char
	// USTimeValue  uint16 // Use uint16 for u_short
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
const MSG_CHARAC_REQ_NUM = 41
const MSG_CHARAC_RES_NUM = 42
const MSG_GM_ADD_INVITEM_NUM = 73
