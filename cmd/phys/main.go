package main

import (
  "log"
  "time"
  "net"
  "bufio"
  "math/rand"
  "encoding/binary"
  "sync"
  //"runtime"
  //"runtime/debug"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool/goroutine"

  "go-space-serv/internal/app/snet"
  "go-space-serv/internal/app/helpers"
  "go-space-serv/internal/app/world"
  "go-space-serv/internal/app/player"
  "go-space-serv/internal/app/phys"

  . "go-space-serv/internal/app/player/types"
  . "go-space-serv/internal/app/phys/types"
  . "go-space-serv/internal/app/snet/types"
)

// physicsServer's job is to run the physics simulation
// and propagate input between clients.
// The simulation runs at 30fps and the loop runs
// at or near 120fps.
// server chooses a time to start simulation
// and updates it every 256 frames.
// server sends this time to clients so that they
// can simulate from the same time as well. This
// keeps the numbers from growing too large.

type physicsServer struct {
  *gnet.EventServer
  pool                *goroutine.Pool

  players             player.Players
  sim                 phys.Simulation
  ipsToPlayers        sync.Map

  launchTime          int64
  tick                time.Duration // loop speed

  // World map data
  worldConn           net.Conn
  worldMap            world.WorldMap
  worldMapBytes       []byte
  worldMapLen         int
}

type UDPCmd byte

const (
  HELLO UDPCmd = iota + 1
  SHUTUP
  DISCONNECT
  CHALLENGE
  WELCOME
)

// protocol constants
// message structure is as follows:
// [ length, command, content ]
//     4b       1b    <= 4091b
const prefixLen         int     = 4
const cmdLen            int     = 1
const maxMsgSize        int     = 1024
const udpAddr           string  = "udp://:9495";
const udpPort           int     = 9495
const maxPacketRetries  int     = 20

func SDBMHash(str string) uint32 {
  var hash uint32 = 0;
  var i uint32 = 0;
  var length uint32 = uint32(len(str))

  for i = 0; i < length; i++ {
    hash = uint32(byte(str[int(i)])) + (hash << 6) + (hash << 16) - hash
  }

  return hash;
}

func main() {
  p := goroutine.Default()
  defer p.Release()

  // Populate config
  // TODO: take this from flags
  var config helpers.Config
  config.TIMESTEP = 34
  config.TIMESTEP_NANO = 34000000
  config.NAME = "SPACE-PHYS"
  config.VERSION = "0.0.1"
  config.PROTOCOL_ID = SDBMHash(config.NAME + config.VERSION)
  helpers.SetConfig(&config)

  log.Printf("PROTOCOL_ID: %d", config.PROTOCOL_ID)

  // connect via TCP to the world server
  dialer := &net.Dialer{
    LocalAddr: &net.TCPAddr{
      IP:   net.ParseIP("127.0.0.1"),
      Port: 9499,
    },
  }
  conn, err := dialer.Dial("tcp", "127.0.0.1:9494")
  if err != nil {
    log.Printf("%s", err)
    panic("couldnt find world server")
  }
  defer conn.Close()

  // Initialize UDP server
  ps := &physicsServer{
    pool: p,
    tick: 8333333,
    launchTime: time.Now().UnixNano(),
    worldConn: conn,
  }

  // react to events from world server
  go ps.ReactWorld(conn)

  // propagate sim output to all players
  go ps.SimToPlayers()

  // consume queue of outgoing messages to all clients.
  go ps.Output()

  rand.Seed(time.Now().UnixNano())

  log.Fatal(gnet.Serve(ps, udpAddr, gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true)))
}

func (ps *physicsServer) Output() {
  for {
    ps.players.SendQueuedMsgs()
    time.Sleep(32)
  }
}

func (ps *physicsServer) SimToPlayers() {
  for {
    ps.players.QueueMsgAll(ps.sim.Pull())
    time.Sleep(32)
  }
}

func (ps *physicsServer) SyncPlayer(p *UdpPlayer, syncTime int64) {
  log.Printf("Syncing %s @ %d", p.GetName(), helpers.NanosToMillis(syncTime))
  var syncMsg NetworkMsg
  syncMsg.PutByte(byte(SSync))
  syncMsg.PutUint16(ps.sim.GetSeq())
  syncMsg.PutUint64(uint64(helpers.NanosToMillis(syncTime)))

  if p.IsActive() {
    p.Sync(syncTime)
    p.AddMsg(&syncMsg)
    log.Printf("Synced %s", p.GetName())
  }
}

func (ps *physicsServer) interpret(i UdpInput, c gnet.Conn) (out []byte) {
  // playerName := i.GetName()
  // player := ps.players.GetPlayer(playerName)

  // if player != nil {
  //   // TODO: authenticate
  //   if i.GetType() == HELLO && !player.IsActive() {
  //     log.Printf("%s connected.", playerName)
  //     log.Printf("%s", c.RemoteAddr())
  //     player.Activate()
  //     player.SetState(SPECTATING)
  //     player.SetConnection(c)
  //     ps.SyncPlayer(player, ps.sim.GetLastSync())
  //   }

  //   switch cmd := i.GetType(); cmd {
  //     case MOVE:
  //       ps.sim.AddMove(&i)
  //     case SPAWN:
  //       player := ps.players.GetPlayer(i.GetName())
  //       if player != nil {
  //         ps.sim.SpawnPlayer(&i, player)
  //       }
  //   }
  // }

  return
}

type UDPHeader struct {
  protocolId uint32
  seq uint16
  ack uint16
}

func (ps *physicsServer) ipIsValid(ip string) bool {
  found := false
  ps.ipsToPlayers.Range(func(key, value interface{}) bool {
    if key.(string) == ip {
      found = true
      return false
    }
    return true
  })

  return found
}

func (ps *physicsServer) React(data []byte, connection gnet.Conn) (out []byte, action gnet.Action) {
  valid := true
  ipString := connection.RemoteAddr().(*net.UDPAddr).IP.String()
  valid = valid && ps.ipIsValid(ipString)

  if valid {
    log.Printf("Received packet from valid IP")
    bytes := append([]byte{}, data...)
    _ = ps.pool.Submit(func() {
      var player *UdpPlayer
      playerId, ok := ps.ipsToPlayers.Load(ipString)
      if ok {
        log.Printf("Found player")
        player = ps.players.GetPlayer(playerId.(string))
      }

      valid = valid && snet.Read_uint32(bytes[0:4]) == helpers.GetProtocolId()
      if valid && player != nil {
        log.Printf("ProtocolId is valid")

        // Respond to HELLO with CHALLENGE
        if player.GetState() == DISCONNECTED {
          // enforce padding to avoid participating in DDoS minification
          if len(bytes) != maxMsgSize {
            log.Printf("Rejecting packet due to lack of padding.")
            return
          }

          cmd := UDPCmd(bytes[4])
          if cmd == HELLO {
            log.Printf("Received HELLO");
            clientSalt := snet.Read_int64(bytes[5:13])
            log.Printf("client salt: %d", clientSalt)
            player.SetClientSalt(clientSalt)
            serverSalt := rand.Int63()
            log.Printf("server salt: %d", serverSalt)
            player.SetServerSalt(serverSalt)
            player.SetConnection(connection)

            msgBytes := make([]byte, maxMsgSize)
            binary.LittleEndian.PutUint32(msgBytes[0:4], helpers.GetProtocolId())
            msgBytes[4] = byte(CHALLENGE)
            binary.LittleEndian.PutUint64(msgBytes[5:13], uint64(clientSalt))
            binary.LittleEndian.PutUint64(msgBytes[13:21], uint64(serverSalt))
            player.SendRepeating(msgBytes, 500, 20)
            player.SetState(CHALLENGED)
          }

          return
        }
        if player.GetState() == CHALLENGED {
          // enforce padding to avoid participating in DDoS minification
          if len(bytes) != maxMsgSize {
            log.Printf("Rejecting packet due to lack of padding.")
            return
          }

          cmd := UDPCmd(bytes[4])
          if cmd == CHALLENGE {
            log.Printf("Received CHALLENGE");
            clientSalt := player.GetClientSalt()
            serverSalt := player.GetServerSalt()
            challengeResponse := snet.Read_int64(bytes[5:13])
            if challengeResponse == clientSalt ^ serverSalt {
              log.Printf("SUCCESS");
              player.SetState(CONNECTED)
              log.Printf("%s passed challenge, welcoming.", ipString)
              msgBytes := make([]byte, 13)
              binary.LittleEndian.PutUint32(msgBytes[0:4], helpers.GetProtocolId())
              msgBytes[4] = byte(WELCOME)
              binary.LittleEndian.PutUint64(msgBytes[5:13], uint64(clientSalt ^ serverSalt))
              player.SendRepeating(msgBytes, 500, 20)
            } else {
              log.Printf("%s failed challenge with %d", clientSalt ^ serverSalt)
            }
          }

          return
        }

        // state == CONNECTED
      } else {
        log.Printf("Invalid protocolId")
      }
    })
  } else {
    log.Printf("Invalid IP %s", ipString)
  }

  return
}

func (ps *physicsServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
  log.Printf("UDP server is listening on %s (multi-cores: %t, loops: %d)\n",
    srv.Addr.String(), srv.Multicore, srv.NumLoops)
  return
}

// world server <-> physics server interaction
func (ps *physicsServer) ReactWorld(c net.Conn) {

  const(
    readingSize byte = iota
    readingWidth
    readingHeight
    readingResolution
    readingMap
    readingEvent
  )

  var state byte = readingSize
  var size int = 0
  var mapBuf []byte = nil
  var mapBufRead int = 0
  reader := bufio.NewReader(c)

  for {
    switch state {
      case readingSize:
        sizeBytes, err := reader.Peek(4)
        if err == nil {
          size = snet.Read_int32(sizeBytes)
          mapBuf = make([]byte, size - 4 - 4)

          reader.Discard(4)

          state = readingWidth
          log.Printf("Read worldMap data length=%d", size)
        } else {
          log.Printf(err.Error())
        }
      case readingWidth:
        widthBytes, err := reader.Peek(4)
        if err == nil {
          ps.worldMap.W = snet.Read_int32(widthBytes)

          reader.Discard(4)

          state = readingHeight
          log.Printf("Read worldMap width=%d", ps.worldMap.W)
        } else {
          log.Printf(err.Error())
        }
      case readingHeight:
        heightBytes, err := reader.Peek(4)
        if err == nil {
          ps.worldMap.H = snet.Read_int32(heightBytes)

          reader.Discard(4)

          state = readingResolution
          log.Printf("Read worldMap height=%d", ps.worldMap.H)
        } else {
          log.Printf(err.Error())
        }
      case readingResolution:
        resBytes, err := reader.Peek(1)
        if err == nil {
          ps.worldMap.Resolution = int(resBytes[0])

          reader.Discard(1)

          state = readingMap
          log.Printf("Read worldMap resolution=%d", ps.worldMap.Resolution)
        } else {
          log.Printf(err.Error())
        }
      case readingMap:
        for b, err := reader.Peek(1); err == nil && mapBufRead < size - 4 - 4 - 1; {
          mapBuf[mapBufRead] = b[0]
          mapBufRead++
          reader.Discard(1)
        }

        if mapBufRead == size - 4 - 4 - 1 {
          ps.worldMap.Deserialize(mapBuf, false)

          reader.Reset(c)

          state = readingEvent
          log.Printf("Read worldMap bytes")

          var msg NetworkMsg
          msg.PutByte(byte(IReady))
          msg.PutUint32(udpPort)
          c.Write(snet.GetDataFromNetworkMsg(&msg))

          ps.sim.Start(&ps.worldMap);
        }
      case readingEvent:
        eventSizeBytes, err := reader.Peek(4)
        if err == nil {
          log.Printf("Received event from world")
          eventSize := snet.Read_int32(eventSizeBytes)
          reader.Discard(4)

          eventContent, err2 := reader.Peek(eventSize)
          if err2 == nil {
            if eventContent[0] == byte(IJoin) {
              playerIdLenBytes := eventContent[1:3]
              playerIdLen := snet.Read_uint16(playerIdLenBytes)
              playerId := snet.Read_utf8(eventContent[3:3+playerIdLen])
              ipLen := snet.Read_uint16(eventContent[3+playerIdLen:3+playerIdLen+2])
              ipString := snet.Read_utf8(eventContent[3+playerIdLen+2:3+playerIdLen+2+ipLen])
              ps.players.Add(playerId, NewPlayerStats())
              log.Printf("Storing %s <-> %s", ipString, playerId)
              ps.ipsToPlayers.Store(ipString, playerId);
            } else if eventContent[0] == byte(ILeave) {
              playerIdLenBytes := eventContent[1:3]
              playerIdLen := snet.Read_uint16(playerIdLenBytes)
              playerId := snet.Read_utf8(eventContent[3:3+playerIdLen])
              ps.players.Remove(playerId)
              ps.sim.RemoveControlledBody(playerId)
            }
          }
          reader.Discard(eventSize)
        } else {
          log.Printf("%s", err.Error())
        }
    }
    time.Sleep(1 * time.Second)
  }
}
