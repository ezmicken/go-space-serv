package snet

import(
  "encoding/binary"
  "unicode/utf8"
  "bytes"
  "log"
  "strings"
  "net"

  . "go-space-serv/internal/app/snet/types"
)

const PrefixLength int = 4

// auto-incrementing id
var nextId uint16 = 0
func GetNextId() uint16 {
  nextId += 1
  return nextId
}

// Get preferred outbound ip of this machine
func GetOutboundIP() (net.IP) {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr).IP

    return localAddr
}

func Read_int64(data []byte) int64 {
  var ret64 int64
  buf := bytes.NewBuffer(data)
  binary.Read(buf, binary.LittleEndian, &ret64)
  return ret64
}

func Read_int32(data []byte) int {
  var ret32 int32
  buf := bytes.NewBuffer(data)
  binary.Read(buf, binary.LittleEndian, &ret32)
  return int(ret32)
}

func Read_uint16(data []byte) uint16 {
  var ret16 uint16
  buf := bytes.NewBuffer(data)
  binary.Read(buf, binary.LittleEndian, &ret16)
  return ret16
}

func Read_utf8(data []byte) string {
  var utf8Name strings.Builder
  for len(data) > 0 {
    r, size := utf8.DecodeRune(data)
    utf8Name.WriteRune(r)
    data = data[size:]
  }

  return utf8Name.String()
}

// Gets the whole message from the stream
func GetNetworkMsgFromData(data [] byte) (*NetworkMsg) {
  dataLen := len(data)
  if dataLen >= PrefixLength {
    msgLen := Read_int32(data[:4])
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
