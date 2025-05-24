package crypto

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"unsafe"
)

const (
	KEYVALLENTH      = 16   // Maximum key length
	KEY_RANGE        = 0xff // Key range
	CAPSULE_BUF_SIZE = 4096
	TAIL_SIZE        = 2
)

// Header structure equivalent to _sHeader
type Header struct {
	// In Go, we need to handle bit fields differently
	// We'll pack this into a single uint16 and use bit operations
	packed uint16
}

// NewHeader creates a new header with the specified values
func NewHeader(length int16, crypto, compressed uint8) *Header {
	h := &Header{}
	h.SetLength(length)
	h.SetCrypto(crypto)
	h.SetCompressed(compressed)
	return h
}

// SetLength sets the length field (12 bits, -2048 to 2047)
func (h *Header) SetLength(length int16) {
	h.packed = (h.packed & 0xF000) | (uint16(length) & 0x0FFF)
}

// GetLength gets the length field
func (h *Header) GetLength() int16 {
	val := h.packed & 0x0FFF
	// Sign extend if negative (bit 11 is set)
	if val&0x0800 != 0 {
		return int16(val | 0xF000)
	}
	return int16(val)
}

// SetCrypto sets the crypto field (2 bits)
func (h *Header) SetCrypto(crypto uint8) {
	h.packed = (h.packed & 0xCFFF) | ((uint16(crypto) & 0x03) << 12)
}

// GetCrypto gets the crypto field
func (h *Header) GetCrypto() uint8 {
	return uint8((h.packed >> 12) & 0x03)
}

// SetCompressed sets the compressed field (2 bits)
func (h *Header) SetCompressed(compressed uint8) {
	h.packed = (h.packed & 0x3FFF) | ((uint16(compressed) & 0x03) << 14)
}

// GetCompressed gets the compressed field
func (h *Header) GetCompressed() uint8 {
	return uint8((h.packed >> 14) & 0x03)
}

// Tail structure equivalent to _Tail
type Tail struct {
	CRC uint8
	Seq uint8
}

// EncapsuleInfo equivalent to _Encapsule_info
type EncapsuleInfo struct {
	Buf    []byte
	Length uint16
}

// DecapsuleInfo equivalent to _Decapsule_info
type DecapsuleInfo struct {
	Buf    []byte
	Length uint16
	Seq    uint8
}

// JCrypto equivalent to _j_Crypto class
type JCrypto struct {
	keyBox      [KEY_RANGE][KEYVALLENTH]byte
	bufMaxSize  int
	bufCurSet   int
	dwKeyLength uint32
}

// NewJCrypto creates a new JCrypto instance
func NewJCrypto(bufSize int) *JCrypto {
	return &JCrypto{
		bufMaxSize: bufSize,
		bufCurSet:  0,
	}
}

// Init initializes the crypto with a key file
func (j *JCrypto) Init(pathName string) error {
	if pathName == "" {
		return nil
	}

	file, err := os.Open(pathName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Clear the key box
	for i := range j.keyBox {
		for k := range j.keyBox[i] {
			j.keyBox[i][k] = 0
		}
	}

	// Read the key data
	_, err = io.ReadFull(file, (*(*[KEY_RANGE * KEYVALLENTH]byte)(unsafe.Pointer(&j.keyBox[0][0])))[:])
	if err != nil {
		return err
	}

	return nil
}

// GetKey returns the key for the given index
func (j *JCrypto) GetKey(key uint8) []byte {
	return j.keyBox[key%KEY_RANGE][:]
}

// Encryption encrypts the data using the specified key
func (j *JCrypto) Encryption(data []byte, key uint8) error {
	cValKey := j.GetKey(key)

	dataLen := len(data)
	k := dataLen / KEYVALLENTH // line count
	l := dataLen % KEYVALLENTH // remaining bytes

	// Convert to uint32 slices for block processing
	uiValKey := (*(*[KEYVALLENTH / 4]uint32)(unsafe.Pointer(&cValKey[0])))[:]

	// Process full blocks
	for i, x, jIdx := 0, 0, 0; i < k; i, x, jIdx = i+1, x+4, jIdx+KEYVALLENTH {
		puiData := (*(*[4]uint32)(unsafe.Pointer(&data[jIdx])))[:]

		puiData[0] = uiValKey[0] ^ puiData[0]
		puiData[1] = ^(uiValKey[2] ^ puiData[1])
		puiData[2] = uiValKey[1] ^ puiData[2]
		puiData[3] = ^(uiValKey[3] ^ puiData[3])
	}

	// Process remaining bytes
	jIdx := k * KEYVALLENTH
	for h := 0; h < l; h++ {
		data[jIdx+h] = ^(cValKey[h] ^ data[jIdx+h])
	}

	return nil
}

// Decryption decrypts the data using the specified key
func (j *JCrypto) Decryption(data []byte, key uint8) error {
	cValKey := j.GetKey(key)
	uiValKey := (*(*[KEYVALLENTH / 4]uint32)(unsafe.Pointer(&cValKey[0])))[:]

	dataLen := len(data)
	k := dataLen / KEYVALLENTH // line count
	l := dataLen % KEYVALLENTH // char count (remaining bytes)

	x := 0
	jIdx := 0
	puiData := (*(*[1 << 20]uint32)(unsafe.Pointer(&data[0]))) // large enough for safety

	// Process full blocks
	for i := 0; i < k; i, x, jIdx = i+1, x+4, jIdx+16 {
		puiData[x+0] = uiValKey[0] ^ puiData[x+0]
		puiData[x+1] = ^(uiValKey[2] ^ puiData[x+1])
		puiData[x+2] = uiValKey[1] ^ puiData[x+2]
		puiData[x+3] = ^(uiValKey[3] ^ puiData[x+3])
	}

	// Process remaining bytes
	for h := 0; h < l; h++ {
		data[jIdx+h] = ^(cValKey[h] ^ data[jIdx+h])
	}

	return nil
}

// Xor performs XOR operation on data
func (j *JCrypto) Xor(data []byte, key uint8) {
	cValKey := j.GetKey(key)

	for i := 0; i < len(data); i++ {
		data[i] = ^data[i] ^ cValKey[i%KEYVALLENTH]
	}
}

// Checksum calculates checksum for the buffer
func (j *JCrypto) Checksum(buffer []byte) uint8 {
	var cksum uint32

	i := 0
	for i < len(buffer) {
		cksum += uint32(buffer[i])
		i++
	}

	cksum = (cksum >> 16) + (cksum & 0xffff)
	cksum += (cksum >> 16)
	return uint8(^cksum)
}

// CCapsulateCrypto equivalent to CCapsulateCrypto class
type CCapsulateCrypto struct {
	*JCrypto
	capsuleBuf []byte
	bufMaxSize int
	bufCurSet  int
	seq        uint8
}

// NewCCapsulateCrypto creates a new CCapsulateCrypto instance
func NewCCapsulateCrypto(bufSize ...int) *CCapsulateCrypto {
	size := CAPSULE_BUF_SIZE
	if len(bufSize) > 0 {
		size = bufSize[0]
	}

	return &CCapsulateCrypto{
		JCrypto:    NewJCrypto(size),
		capsuleBuf: make([]byte, size),
		bufMaxSize: size,
		bufCurSet:  0,
		seq:        0,
	}
}

// InitSeqNum initializes the sequence number
func (c *CCapsulateCrypto) InitSeqNum(seq uint8) {
	c.seq = seq
}

// Encapsulate encapsulates a packet with encryption and checksum
func (c *CCapsulateCrypto) Encapsulate(packet []byte, key ...uint8) EncapsuleInfo {
	var keyVal uint8
	if len(key) > 0 {
		keyVal = key[0]
	} else {
		keyVal = c.seq
		c.seq++
	}

	// Parse header
	if len(packet) < 2 {
		return EncapsuleInfo{Buf: nil, Length: 0}
	}

	header := (*Header)(unsafe.Pointer(&packet[0]))
	length := int(header.GetLength())

	if length+TAIL_SIZE > c.bufMaxSize {
		return EncapsuleInfo{Buf: nil, Length: 0}
	}

	// Copy packet to capsule buffer
	copy(c.capsuleBuf[:length], packet[:length])

	// Update header pointer to capsule buffer
	header = (*Header)(unsafe.Pointer(&c.capsuleBuf[0]))

	// Add tail
	tail := (*Tail)(unsafe.Pointer(&c.capsuleBuf[length]))
	tail.CRC = c.Checksum(c.capsuleBuf[:length])
	tail.Seq = keyVal
	header.SetLength(int16(length + TAIL_SIZE))

	newLength := length + TAIL_SIZE

	// Encrypt if needed
	if header.GetCrypto() == 1 {
		if err := c.Encryption(c.capsuleBuf[2:newLength], keyVal); err != nil {
			return EncapsuleInfo{Buf: nil, Length: 0}
		}
	}

	return EncapsuleInfo{
		Buf:    c.capsuleBuf[:newLength],
		Length: uint16(newLength),
	}
}

// Decapsulate decapsulates a packet with decryption and checksum verification
func (c *CCapsulateCrypto) Decapsulate(packet []byte, key ...uint8) DecapsuleInfo {
	var keyVal uint8
	if len(key) > 0 {
		keyVal = key[0]
	} else {
		keyVal = c.seq
	}

	if len(packet) < 2 {
		return DecapsuleInfo{Buf: nil, Length: 0, Seq: 0}
	}

	header := (*Header)(unsafe.Pointer(&packet[0]))
	length := int(header.GetLength())

	// Decrypt if needed
	if header.GetCrypto() == 1 {
		if err := c.Decryption(packet[2:length], keyVal); err != nil {
			return DecapsuleInfo{Buf: nil, Length: 0, Seq: 0}
		}
	}

	// Remove tail and verify checksum
	newLength := length - TAIL_SIZE
	header.SetLength(int16(newLength))

	tail := (*Tail)(unsafe.Pointer(&packet[newLength]))

	if tail.CRC == c.Checksum(packet[:newLength]) {
		return DecapsuleInfo{
			Buf:    packet[:newLength],
			Length: uint16(newLength),
			Seq:    tail.Seq,
		}
	}

	return DecapsuleInfo{Buf: nil, Length: 0, Seq: tail.Seq}
}

// Example usage
func main() {
	// Create crypto instance
	crypto := NewCCapsulateCrypto()

	// Initialize with key file (optional)
	err := crypto.Init("keyfile.dat")
	if err != nil {
		fmt.Printf("Warning: Could not load key file: %v\n", err)
	}

	// Create a test packet
	testData := make([]byte, 10)
	for i := range testData {
		testData[i] = byte(i)
	}

	// Create packet with header
	packet := make([]byte, 2+len(testData))
	header := NewHeader(int16(2+len(testData)), 1, 0) // crypto enabled
	binary.LittleEndian.PutUint16(packet[0:2], header.packed)
	copy(packet[2:], testData)

	// Encapsulate
	encResult := crypto.Encapsulate(packet)
	if encResult.Buf != nil {
		fmt.Printf("Encapsulation successful, length: %d\n", encResult.Length)

		// Decapsulate
		decResult := crypto.Decapsulate(encResult.Buf)
		if decResult.Buf != nil {
			fmt.Printf("Decapsulation successful, length: %d, seq: %d\n",
				decResult.Length, decResult.Seq)
		} else {
			fmt.Println("Decapsulation failed")
		}
	} else {
		fmt.Println("Encapsulation failed")
	}
}
