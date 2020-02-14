package snet

import(
	"encoding/binary"
)

type NetworkMsg struct {
	Size int
	Data []byte
}

func initData(msg *NetworkMsg) {
	if msg.Data == nil {
		msg.Data = []byte{}
		msg.Size = 0
	}
}

func (msg *NetworkMsg) PutByte(val byte) {
	initData(msg)
	msg.Size += 1
	msg.Data = append(msg.Data, val)
}

func (msg *NetworkMsg) PutBytes(val []byte) {
	initData(msg)
	msg.Size += len(val)
	msg.Data = append(msg.Data, val...)
}

func (msg *NetworkMsg) PutUint32(val int) {
	initData(msg)

  var valBytes = make([]byte, 4)
  binary.LittleEndian.PutUint32(valBytes, uint32(val))

  msg.Size += 4
  msg.Data = append(msg.Data, valBytes...)
}

func (msg *NetworkMsg) PutUint16(val uint16) {
	initData(msg)

	var valBytes = make([]byte, 2)
	binary.LittleEndian.PutUint16(valBytes, val)

	msg.Size += 2
	msg.Data = append(msg.Data, valBytes...)
}

func (msg *NetworkMsg) PutUint64(val uint64) {
	initData(msg)

	var valBytes = make([]byte, 8)
	binary.LittleEndian.PutUint64(valBytes, val)

	msg.Size += 8;
	msg.Data = append(msg.Data, valBytes...)
}

func (msg *NetworkMsg) PutString(val string) {
	initData(msg)

	stringBytes := []byte(val)

	stringBytesLen := len(stringBytes)
	msg.PutUint16(uint16(stringBytesLen))

	msg.Size += len(stringBytes)
	msg.Data = append(msg.Data, stringBytes...)
}
