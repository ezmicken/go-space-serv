package udp

import (
  "log"
  "math"
  //"math/rand"
  "time"
  //"encoding/binary"

  "github.com/panjf2000/gnet"
  "github.com/google/uuid"

  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/util"
)

type UDPPlayerState byte

const BUFFER_SIZE uint16 = 1024
const SHUTUP_TIME int    = 10
const TIMEOUT            = 5000000000

type PacketData struct {
  Acked     bool
  SendTime  int64
  Size      int32
}

type UDPPlayer struct {
  Id                uuid.UUID
  Outgoing    chan  UDPMsg

  incoming    chan  UDPMsg
  connection        gnet.Conn
  msgFactory        UDPMsgFactory

  spamChan          chan struct{}
  active            bool
  lastSync          int64
  lastPacketRx      int64

  // packet stuff
  clientSalt        int64
  serverSalt        int64

  seqBuffer         []uint32
  packetData        []PacketData

  txSeq             uint16
  txAck             uint16
  rxSeq             uint16
  txAckedBytes      int64
  prevAckedBytes    int64

  shutupTx          int
  shutupRx          int

  packetBuffer      []byte
  packetBufferTail  int
  packetBufferEmpty bool

  connector         *UDPConnector
}

func NewPlayer(in chan UDPMsg, id uuid.UUID, factory UDPMsgFactory) *UDPPlayer {
  var p UDPPlayer
  p.active = false
  p.lastSync = 0
  p.lastPacketRx = 0

  p.txSeq = 0
  p.txAck = 0
  p.rxSeq = 0
  p.txAckedBytes = 0
  p.prevAckedBytes = 0
  p.packetBufferTail = HEADER_SIZE
  p.packetBufferEmpty = true
  p.shutupRx = 0
  p.shutupTx = 0

  p.incoming = in
  p.Outgoing = make(chan UDPMsg, 100)
  p.Id = id
  p.msgFactory = factory

  p.seqBuffer = make([]uint32, BUFFER_SIZE)
  p.packetData = make([]PacketData, BUFFER_SIZE)
  p.packetBuffer = make([]byte, BUFFER_SIZE)

  p.connector = NewUDPConnector(500, 20)

  return &p
}

// Compares two sequence values and their difference.
func seqGreaterThan(s1 uint16, s2 uint16) bool {
  return ((s1 > s2) && (s1-s2 <= 32768)) || ((s1 < s2) && (s2-s1 > 32768))
}

func (p *UDPPlayer) getPacketData(seq uint16) PacketData {
  return p.packetData[seq % BUFFER_SIZE]
}

func (p *UDPPlayer) insertPacketData(pd PacketData, seq uint16) {
  idx := seq % BUFFER_SIZE
  p.seqBuffer[idx] = uint32(seq)
  p.packetData[idx] = pd
}

func (p *UDPPlayer) onPacketAcked(seq uint16) {
  idx := seq % BUFFER_SIZE
  if p.seqBuffer[idx] == uint32(seq) {
    p.packetData[idx].Acked = true
    p.seqBuffer[idx] = math.MaxUint32
    p.txAckedBytes += int64(p.packetData[idx].Size)
  }

  // TODO: calculate exponential moving average RTT
}

func (p *UDPPlayer) getMsg() UDPMsg {
  var tmp UDPMsg
  select {
  case tmp = <-p.Outgoing:
  default:
    return nil
  }

  return tmp
}

// Sends a packet every rate milliseconds count times
func (p *UDPPlayer) sendRepeating(msg []byte, rate, count int) {
  if p.spamChan != nil {
    close(p.spamChan)
  }

  ticker := time.NewTicker(time.Duration(rate) * time.Millisecond)
  defer ticker.Stop()

  log.Printf("sending first %d", msg[4])
  p.connection.SendTo(msg)

  // Use a local reference to the chan
  // for the case where it was closed
  // and re-opened during Sleep
  thisChan := make(chan struct{})
  p.spamChan = thisChan

  go func() {
    for i := 0; i < count; i++ {
      _, open := <-thisChan
      if !open {
        return
      }

      log.Printf("sending %d (iteration %d)", msg[4], i)
      p.connection.SendTo(msg)
      <- ticker.C
    }

    log.Printf("Player.sendRepeating finished all iterations.")
    close(thisChan)
  }()
}

func (p *UDPPlayer) Unpack(packet []byte) {
  head := 8
  tail := 0
  salt := snet.Read_int64(packet[tail:head])
  tail = head
  msgLen := len(packet)
  if salt != (p.clientSalt ^ p.serverSalt) {
    return
  }

  p.lastPacketRx = time.Now().UnixNano()

  // handle seq/ack
  head += 2
  seq := snet.Read_uint16(packet[tail:head])
  tail = head
  head += 2
  ack := snet.Read_uint16(packet[tail:head])
  tail = head

  if seqGreaterThan(ack, p.txAck) {
    p.txAck = ack
    p.onPacketAcked(seq)
  }

  if (head < msgLen) {
    cmd := UDPCmd(packet[head])

    if cmd == SHUTUP {
      p.shutupRx++
      return
    } else {
      p.shutupRx = 0
    }

    if seqGreaterThan(seq, p.rxSeq) {
      p.rxSeq = seq
      log.Printf("rx: %v", cmd)

      for head < msgLen {
        head = p.msgFactory.CreateAndPublishMsg(packet, head, p.incoming, p.Id)
      }
    }
  }
}

func (p *UDPPlayer) PackAndSend() {
  numMsgs := len(p.Outgoing)

  shouldStop := numMsgs == 0
  shouldStop = shouldStop && p.txSeq == p.txAck
  shouldStop = shouldStop && p.shutupTx > SHUTUP_TIME
  shouldStop = shouldStop && p.shutupRx > SHUTUP_TIME
  shouldStop = shouldStop || p.connector.GetState() < CONNECTED

  if shouldStop && p.active {
    p.active = false
    return
  }

  if shouldStop {
    return
  }

  p.active = true

  var header UDPHeader
  header.ProtocolId = helpers.GetProtocolId()
  header.Salt = p.clientSalt ^ p.serverSalt
  header.Ack = p.rxSeq

  // Client has acknowledged all our messages
  if numMsgs == 0 && p.txSeq == p.txAck {
    header.Seq = p.txSeq
    var pd PacketData
    pd.Acked = false
    pd.SendTime = helpers.NowMillis()
    pd.Size = 1
    p.insertPacketData(pd, p.txSeq)
    p.packetBuffer[HEADER_SIZE] = byte(SHUTUP)
    p.packetBufferEmpty = true
    p.shutupTx++
    header.Serialize(p.packetBuffer)
    p.connection.SendTo(p.packetBuffer[:HEADER_SIZE+1])
    return
  } else {
    p.shutupTx = 0
  }

  // Move the tail based on newest ack
  // p.packetBufferTail = HEADER_SIZE
  // for i := p.txSeq; i < p.txAck; i-- {
  //   pd := p.getPacketData(i)
  //   p.packetBufferTail += int(pd.Size)
  // }

  if p.txAckedBytes > p.prevAckedBytes {
    ackedBytes := int(p.txAckedBytes - p.prevAckedBytes)
    for i := 0; i < ackedBytes; i++ {
      // TODO: stop this from overflowing.
      p.packetBuffer[HEADER_SIZE + i] = p.packetBuffer[HEADER_SIZE + ackedBytes + i];
    }
    p.packetBufferTail -= ackedBytes
    p.prevAckedBytes = p.txAckedBytes
    log.Printf("tail moved left %v(%v)", ackedBytes, p.packetBufferTail)
  }

  for j := 0; j < 50; j++ {
    tmp := p.getMsg()
    if tmp == nil {
      break
    }

    msg := tmp.(UDPMsg)
    msgSize := msg.GetSize()
    if p.packetBufferTail + msgSize >= int(BUFFER_SIZE) {
      p.Outgoing <- msg
      log.Printf("%v packetBuffer overflow.", p.Id)
    }

    p.txSeq++
    header.Seq = p.txSeq
    log.Printf("tx: %v", msg.GetCmd())
    msg.Serialize(p.packetBuffer[p.packetBufferTail : p.packetBufferTail+msgSize])
    p.packetBufferTail += msgSize
    p.packetBufferEmpty = false;
    log.Printf("tail moved right %v(%v)", msgSize, p.packetBufferTail)

    var pd PacketData
    pd.Acked = false
    pd.SendTime = helpers.NowMillis()
    pd.Size = int32(msgSize)
    p.insertPacketData(pd, p.txSeq)

    msg = nil
  }
  header.Serialize(p.packetBuffer)
  p.connection.SendTo(p.packetBuffer[:p.packetBufferTail])
}

// TODO: add sequence to this
func (p *UDPPlayer) AuthenticateConnection(bytes []byte, conn gnet.Conn, ip string) bool {
  return p.connector.Authenticate(bytes, conn, ip)
}

func (p *UDPPlayer) Disconnect() {
  close(p.Outgoing)
  p.connector.Disconnect()
}

func (p *UDPPlayer) GetState() UDPConnectorState {
  return p.connector.GetState()
}

func (p *UDPPlayer) IsActive() bool {
  return p.active
}

func (p *UDPPlayer) IsTimedOut() bool {
  return time.Now().UnixNano() - p.lastPacketRx > TIMEOUT
}
