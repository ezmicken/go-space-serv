package phys

type UDPHeader struct {
  ProtocolId  uint32
  Salt        int64
  Seq         uint16
  Ack         uint16
}

const HEADER_SIZE = 16
