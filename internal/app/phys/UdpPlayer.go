package phys

import (
  "log"
  "math"
  "time"
  "encoding/binary"

  "github.com/panjf2000/gnet"

  "go-space-serv/internal/app/snet"
  "go-space-serv/internal/app/util"

  . "go-space-serv/internal/app/phys/interface"
  . "go-space-serv/internal/app/phys/types"
  . "go-space-serv/internal/app/player/types"
)

type UdpPlayerState byte

const (
  DISCONNECTED UdpPlayerState = iota
  CHALLENGED
  CONNECTED
  PLAYING
  SPECTATING
)

const BUFFER_SIZE uint16 = 1024
const SHUTUP_TIME int    = 10

type PacketData struct {
  Acked     bool
  SendTime  int64
  Size      int32
}

type UdpPlayer struct {
  spamChan          chan struct{}
  name              string
  active            bool
  lastSync          int64

  // packet stuff
  clientSalt        int64
  serverSalt        int64
  seqBuffer         []uint32
  packetData        []PacketData
  txSeq             uint16
  txAck             uint16
  rxSeq             uint16

  shutupTx          int
  shutupRx          int

  packetBuffer      []byte
  packetBufferTail  int
  packetBufferEmpty bool

  connection        gnet.Conn
  toSim             chan UDPMsg
  toClient          chan UDPMsg
  state             UdpPlayerState
  stats             *PlayerStats
}

func NewUdpPlayer(n string) *UdpPlayer {
  var p UdpPlayer
  p.name = n
  p.active = false
  p.state = DISCONNECTED
  p.lastSync = 0

  p.txSeq = 0
  p.txAck = 0
  p.rxSeq = 0
  p.packetBufferTail = 0
  p.packetBufferEmpty = true
  p.shutupRx = 0
  p.shutupTx = 0

  p.toClient = make(chan UDPMsg, 100)

  p.seqBuffer = make([]uint32, BUFFER_SIZE)
  p.packetData = make([]PacketData, BUFFER_SIZE)
  p.packetBuffer = make([]byte, BUFFER_SIZE)

  return &p
}

func (p *UdpPlayer) SetSimChan(ch chan UDPMsg) {
  p.toSim = ch
}

// Sends a packet every rate milliseconds count times
func (p *UdpPlayer) SendRepeating(msg []byte, rate, count int) {
  if p.spamChan != nil {
    close(p.spamChan)
  }

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
      time.Sleep(time.Duration(rate) * time.Millisecond)
    }

    log.Printf("Player.SendRepeating finished all iterations.")
    close(thisChan)
  }()
}

// Compares two sequence values and their difference.
func seqGreaterThan(s1 uint16, s2 uint16) bool {
  return ((s1 > s2) && (s1-s2 <= 32768)) || ((s1 < s2) && (s2-s1 > 32768))
}

func (p *UdpPlayer) getPacketData(seq uint16) PacketData {
  return p.packetData[seq % BUFFER_SIZE]
}

func (p *UdpPlayer) insertPacketData(pd PacketData, seq uint16) {
  idx := seq % BUFFER_SIZE
  p.seqBuffer[idx] = uint32(seq)
  p.packetData[idx] = pd
}

func (p *UdpPlayer) onPacketAcked(seq uint16) {
  idx := seq % BUFFER_SIZE
  if p.seqBuffer[idx] == uint32(seq) {
    p.packetData[idx].Acked = true
    p.seqBuffer[idx] = math.MaxUint32
  }

  // TODO: calculate exponential moving average RTT
}

func (p *UdpPlayer) serializeHeader(header UDPHeader) {
  binary.LittleEndian.PutUint32(p.packetBuffer[:4], header.ProtocolId)
  binary.LittleEndian.PutUint64(p.packetBuffer[4:12], uint64(header.Salt))
  binary.LittleEndian.PutUint16(p.packetBuffer[12:14], header.Seq)
  binary.LittleEndian.PutUint16(p.packetBuffer[14:16], header.Ack)
}

func (p *UdpPlayer) PackAndSend() {
  numMsgs := len(p.toClient)
  shouldStop := numMsgs == 0
  shouldStop = shouldStop && p.txSeq == p.txAck
  shouldStop = shouldStop && p.shutupTx > SHUTUP_TIME
  shouldStop = shouldStop && p.shutupRx > SHUTUP_TIME
  shouldStop = shouldStop || p.state < CONNECTED

  if shouldStop {
    return
  }

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
    p.serializeHeader(header)
    p.connection.SendTo(p.packetBuffer[:HEADER_SIZE+1])
    return
  } else {
    p.shutupTx = 0
  }

  // Move the tail based on newest ack
  p.packetBufferTail = HEADER_SIZE
  for i := p.txSeq; i <= p.txAck; i-- {
    pd := p.getPacketData(i)
    p.packetBufferTail += int(pd.Size)
  }

  for j := 0; j < 50; j++ {
    tmp := p.getMsg()
    if tmp == nil {
      break
    }

    msg := tmp.(UDPMsg)
    msgSize := msg.GetSize()
    if p.packetBufferTail + msgSize >= int(BUFFER_SIZE) {
      p.toClient <- msg
      log.Printf("%s packetBuffer overflow.", p.name)
    }

    if !p.packetBufferEmpty {
      for k := p.packetBufferTail + msgSize; k >= (HEADER_SIZE+msgSize); k-- {
        p.packetBuffer[k] = p.packetBuffer[k-msgSize]
      }
    }

    p.txSeq++
    header.Seq = p.txSeq
    p.packetBufferTail += msgSize
    msg.Serialize(p.packetBuffer[HEADER_SIZE : HEADER_SIZE+msgSize])
    p.packetBufferEmpty = false;

    var pd PacketData
    pd.Acked = false
    pd.SendTime = helpers.NowMillis()
    pd.Size = int32(msgSize)

    msg = nil
  }

  p.serializeHeader(header)
  p.connection.SendTo(p.packetBuffer[:p.packetBufferTail])
}

func (p *UdpPlayer) Unpack(packet []byte) {
  head := 8
  tail := 0
  salt := snet.Read_int64(packet[tail:head])
  tail = head
  msgLen := len(packet)
  if salt != (p.clientSalt ^ p.serverSalt) {
    return
  }

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

  cmd := UDPCmd(packet[head])

  if cmd == SHUTUP {
    p.shutupRx++
    return
  } else {
    p.shutupRx = 0
  }

  if seqGreaterThan(seq, p.rxSeq) {
    p.rxSeq = seq

    for head < msgLen {
      head = CreateAndPublishMsg(packet, head, p.toSim, p.name)
    }
  }
}

func (p *UdpPlayer) AddMsg(msg UDPMsg) {
  select {
  case p.toClient <- msg:
  default:
    log.Printf("%s msg queue full. Discarding...", p.name)
  }
}

func (p *UdpPlayer) getMsg() UDPMsg {
  var tmp UDPMsg
  select {
  case tmp = <-p.toClient:
  default:
    return nil
  }

  return tmp
}

func (p *UdpPlayer) GetName() string {
  return p.name
}

func (p *UdpPlayer) Activate() {
  p.active = true
}

func (p *UdpPlayer) Deactivate() {
  p.active = false
}

func (p *UdpPlayer) IsActive() bool {
  return p.active
}

func (p *UdpPlayer) SetState(s UdpPlayerState) {
  p.state = s
}

func (p *UdpPlayer) SetStats(s *PlayerStats) {
  p.stats = s
}

func (p *UdpPlayer) GetStats() *PlayerStats {
  return p.stats
}

func (p *UdpPlayer) GetState() UdpPlayerState {
  return p.state
}

func (p *UdpPlayer) SetConnection(c gnet.Conn) {
  p.connection = c
}

func (p *UdpPlayer) SetClientSalt(salt int64) {
  p.clientSalt = salt
}
func (p *UdpPlayer) GetClientSalt() int64 {
  return p.clientSalt
}
func (p *UdpPlayer) SetServerSalt(salt int64) {
  p.serverSalt = salt
}
func (p *UdpPlayer) GetServerSalt() int64 {
  return p.serverSalt
}

func (p *UdpPlayer) GetConnection() gnet.Conn {
  return p.connection
}

func (p *UdpPlayer) Sync(time int64) {
  p.lastSync = time
}

func (p *UdpPlayer) GetLastSync() int64 {
  return p.lastSync
}
