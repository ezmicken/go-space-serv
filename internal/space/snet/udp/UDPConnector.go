package udp

import (
  "log"
  "math/rand"
  "encoding/binary"
  "time"
  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/util"
  "github.com/panjf2000/gnet"
)

type UDPConnectorState byte

const (
  DISCONNECTED UDPConnectorState = iota
  CHALLENGED
  CONNECTED
)

type UDPConnector struct {
  spamChan chan struct{}
  clientSalt int64
  serverSalt int64
  rate int
  count int
  state UDPConnectorState
  connection gnet.Conn

  IP string
}

func NewUDPConnector(rate, count int) *UDPConnector {
  var c UDPConnector
  c.clientSalt = -1
  c.serverSalt = -1
  c.rate = rate
  c.count = count
  c.spamChan = nil
  c.state = DISCONNECTED

  return &c
}

func (this *UDPConnector) GetState() UDPConnectorState {
  return this.state
}

func (this *UDPConnector) Disconnect() {
  this.state = DISCONNECTED
}

// TODO: add sequence to this.
func (this *UDPConnector) Authenticate(bytes []byte, conn gnet.Conn, ip string) bool {
  this.IP = ip
  // respond to HELLO with CHALLENGE
  if this.state == DISCONNECTED {
    if !packetIsFull(bytes) {
      log.Printf("Rejecting packet due to lack of padding.") // TODO: log ip.
      return false
    }

    cmd := UDPCmd(bytes[4])
    if cmd == HELLO {
      log.Printf("Received HELLO")
      this.clientSalt = snet.Read_int64(bytes[5:13])
      this.serverSalt = rand.Int63()
      this.connection = conn

      msgBytes := make([]byte, helpers.GetConfig().MAX_MSG_SIZE)
      binary.LittleEndian.PutUint32(msgBytes[0:4], helpers.GetProtocolId())
      msgBytes[4] = byte(CHALLENGE)
      binary.LittleEndian.PutUint64(msgBytes[5:13], uint64(this.clientSalt))
      binary.LittleEndian.PutUint64(msgBytes[13:21], uint64(this.serverSalt))
      this.sendRepeating(msgBytes)
      this.state = CHALLENGED
    }

    return false
  }

  if this.state == CHALLENGED {
    if !packetIsFull(bytes) {
      log.Printf("Rejecting packet due to lack of padding.") // TODO: log ip.
      return false
    }

    cmd := UDPCmd(bytes[4])
    if cmd == CHALLENGE {
      challengeResponse := snet.Read_int64(bytes[5:13])
      if challengeResponse == this.clientSalt ^ this.serverSalt {
        this.state = CONNECTED
        log.Printf("%v is welcome.", this.IP)
        msgBytes := make([]byte, 13)
        binary.LittleEndian.PutUint32(msgBytes[0:4], helpers.GetProtocolId())
        msgBytes[4] = byte(WELCOME)
        binary.LittleEndian.PutUint64(msgBytes[5:13], uint64(this.clientSalt ^ this.serverSalt))
        this.sendRepeating(msgBytes)
        return true
      }
    }

    return false
  }

  return false
}

func (this *UDPConnector) sendRepeating(msg []byte) {
  if this.spamChan != nil {
    close(this.spamChan)
  }

  ticker := time.NewTicker(time.Duration(this.rate) * time.Millisecond)
  defer ticker.Stop()

  log.Printf("sending first %d", msg[4])
  this.connection.SendTo(msg)

  // Use a local reference to the chan
  // for the case where it was closed
  // and re-opened during Sleep
  currentChan := make(chan struct{})
  this.spamChan = currentChan

  go func() {
    for i := 0; i < this.count; i++ {
      _, open := <-currentChan
      if !open {
        return
      }

      log.Printf("sending %d (iteration %d)", msg[4], i)
      this.connection.SendTo(msg)
      <- ticker.C
    }

    log.Printf("Player.sendRepeating finished all iterations.")
    close(currentChan)
  }()
}

func packetIsFull(bytes []byte) bool {
  return len(bytes) == helpers.GetConfig().MAX_MSG_SIZE
}
