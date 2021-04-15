package udp

import(
  "encoding/binary"
)

type UDPHeader struct {
  ProtocolId  uint32    // ID of this protocol
  Salt        int64     // client salt ^ server salt
  Ack         uint16    // Sequence that this packet acknowledges
  Seq         uint16    // Seqeuence number of this packet
  Redundant   byte      // # of redundant messages in the buffer
}

const HEADER_SIZE = 17

func (h UDPHeader) Serialize(packet []byte) {
  binary.LittleEndian.PutUint32(packet[:4], h.ProtocolId)
  binary.LittleEndian.PutUint64(packet[4:12], uint64(h.Salt))
  binary.LittleEndian.PutUint16(packet[12:14], h.Ack)
  binary.LittleEndian.PutUint16(packet[14:16], h.Seq)
  packet[16] = h.Redundant
}
