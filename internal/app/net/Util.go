package net

import(
	"encoding/binary"
	"bytes"
	"log"

	. "go-space-serv/internal/app/net/types"
)

const PrefixLength int = 4

func read_int(data []byte) (ret int) {
  buf := bytes.NewBuffer(data)
  binary.Read(buf, binary.LittleEndian, &ret)
  return
}

func GetNetworkMsgFromData(data [] byte) (*NetworkMsg) {
  dataLen := len(data)
  if dataLen >= PrefixLength {
    msgLen := read_int(data)
    if dataLen - PrefixLength == msgLen {
      msgData := data[4:]
      return &NetworkMsg{Size: msgLen, Data: msgData}
    }
  }

  return nil
}

func GetDataFromNetworkMsg(msg *NetworkMsg) (out []byte) {
  // size
  sizeBuf := new(bytes.Buffer)
  err := binary.Write(sizeBuf, binary.LittleEndian, int32(msg.Size))
  if err != nil {
    log.Printf("Unable to convert msg size to byte. err = %s", err)
    return nil
  }
  out = append([]byte{}, sizeBuf.Bytes()...)

  // content
  if msg.Data != nil {
    out = append(out, msg.Data...)
  }

  return
}
