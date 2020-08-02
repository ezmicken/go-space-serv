package udp

import(
  "encoding/binary"
)

type UDPHeader struct {
  ProtocolId  uint32
  Salt        int64
  Seq         uint16
  Ack         uint16
}

const HEADER_SIZE = 16

func (h UDPHeader) Serialize(packet []byte) {
  binary.LittleEndian.PutUint32(packet[:4], h.ProtocolId)
  binary.LittleEndian.PutUint64(packet[4:12], uint64(h.Salt))
  binary.LittleEndian.PutUint16(packet[12:14], h.Seq)
  binary.LittleEndian.PutUint16(packet[14:16], h.Ack)
}
