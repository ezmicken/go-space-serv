package main

import (
  "log"
  "time"
  "net"
  "bufio"
  "math/rand"
  "encoding/binary"
  "sync"
  "context"
  "os"
  "os/signal"
  //"runtime"
  //"runtime/debug"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool/goroutine"

  "go-space-serv/internal/app/snet"
  "go-space-serv/internal/app/util"
  "go-space-serv/internal/app/world"

  . "go-space-serv/internal/app/player/types"
  . "go-space-serv/internal/app/phys"
  . "go-space-serv/internal/app/phys/types"
  . "go-space-serv/internal/app/snet/types"
)

const cmdLen            int     = 1
const maxMsgSize        int     = 1024
const udpPort           int     = 9495

func SDBMHash(str string) uint32 {
  var hash uint32 = 0;
  var i uint32 = 0;
  var length uint32 = uint32(len(str))

  for i = 0; i < length; i++ {
    hash = uint32(byte(str[int(i)])) + (hash << 6) + (hash << 16) - hash
  }

  return hash;
}

type physicsServer struct {
  *gnet.EventServer

  // Lifecycle
  ctx                 context.Context
  lifeWG              sync.WaitGroup
  life          chan  struct{}
  state               ServerState
  cancel              func()

  players             Players
  sim                 Simulation
  ipsToPlayers        sync.Map

  launchTime          int64
  tick                time.Duration   // loop speed
  pool                *goroutine.Pool

  // World map data
  worldConn           net.Conn
  worldConnOpen       bool
  worldMap            world.WorldMap
  worldMapBytes       []byte
  worldMapLen         int
}

func main() {
  p := goroutine.Default()
  defer p.Release()

  // Populate config
  // TODO: take this from flags
  var config helpers.Config
  config.TIMESTEP = 33
  config.TIMESTEP_NANO = 33000000
  config.NAME = "SPACE-PHYS"
  config.VERSION = "0.0.1"
  config.PROTOCOL_ID = SDBMHash(config.NAME + config.VERSION)
  helpers.SetConfig(&config)

  log.Printf("PROTOCOL_ID: %d", config.PROTOCOL_ID)

  // Initialize UDP server
  ps := &physicsServer{
    pool: p,
    tick: 8333333,
    launchTime: time.Now().UnixNano(),
    state: DEAD,
    worldConnOpen: false,
  }

  rand.Seed(time.Now().UnixNano())

  // Start listening for operating system signals
  ps.ctx, ps.cancel = context.WithCancel(context.Background())
  ps.life = make(chan struct{})
  go ps.sig()
  go ps.live()

  // Wait for context to finish
  <-ps.ctx.Done()
  log.Printf("Shuttind down...");
  ps.state = SHUTDOWN
  <-ps.life

  log.Printf("end")
}

func (ps *physicsServer) live() {
  var waitWorld sync.WaitGroup
  ps.state = WAIT_WORLD

  waitWorld.Add(1)
  go ps.world(&waitWorld)
  waitWorld.Wait()

  go ps.serve("udp://:9495")

  ps.state = ALIVE
  ps.lifeWG.Wait()

  close(ps.life)
}

func (ps *physicsServer) sig() {
  signalChan := make(chan os.Signal, 1)
  signal.Notify(signalChan, os.Interrupt)

  oscall := <-signalChan
  log.Printf("system call: %+v", oscall)

  ps.cancel()
}

func (ps *physicsServer) serve(udpAddr string) {
  ps.lifeWG.Add(1)
  defer ps.lifeWG.Done()

  log.Printf("Listening to %s", udpAddr)
  err := gnet.Serve(ps, udpAddr, gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true))
  if err != nil {
    log.Fatal(err)
  }
}

// world server <-> physics server interaction
func (ps *physicsServer) world(wg *sync.WaitGroup) {
  ps.lifeWG.Add(1)
  defer ps.lifeWG.Done()

  // connect via TCP to the world server
  dialer := &net.Dialer{
    LocalAddr: &net.TCPAddr{
      IP:   net.ParseIP("127.0.0.1"),
      Port: 9499,
    },
  }
  c, err := dialer.Dial("tcp", "127.0.0.1:9494")
  if err != nil {
    log.Printf("%s", err)
    panic("couldnt find world server")
  }

  ps.worldConn = c
  ps.worldConnOpen = true
  ps.state = SETUP

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

  for ps.state <= ALIVE {
    select {
    case <-ps.ctx.Done():
      log.Printf("Closing world connection...")
      return
    default:
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

            ps.sim.Start(&ps.worldMap, &ps.players);
            wg.Done()
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
              } else if eventContent[0] == byte(IShutdown) {
                ps.cancel()
              }
            }
            reader.Discard(eventSize)
          } else if err.Error() == "EOF" {
            ps.cancel() // TODO: retry connecting to world
            return
          } else if ps.state == SHUTDOWN {
            return
          } else {
            log.Printf("worldConn err: " + err.Error())
          }
      }
      time.Sleep(250 * time.Millisecond)
    }
  }
}

func (ps *physicsServer) OnShutdown(srv gnet.Server) {
  if ps.worldConnOpen == true {
    log.Printf("IShutdown -> world")
    // TODO: do this from world() when we actually
    // write something other than this so we can flush.
    ps.worldConn.Write([]byte{1, 0, 0, 0, byte(IShutdown)})
    ps.worldConn.Close()
  }

  ps.state = DEAD
}

func (ps *physicsServer) Tick() (delay time.Duration, action gnet.Action) {
  delay = ps.tick
  if ps.state == SHUTDOWN {
    action = gnet.Shutdown
  }
  return
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
    bytes := data
    _ = ps.pool.Submit(func() {
      var player *UdpPlayer
      playerId, ok := ps.ipsToPlayers.Load(ipString)
      if ok {
        player = ps.players.GetPlayer(playerId.(string))
      }

      valid = valid && snet.Read_uint32(bytes[0:4]) == helpers.GetProtocolId()
      if valid && player != nil {

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
            player.SetClientSalt(clientSalt)
            serverSalt := rand.Int63()
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
              player.SetState(CONNECTED)
              player.SetSimChan(ps.sim.GetPlayerChan())
              log.Printf("%s passed challenge, welcoming.", ipString)
              msgBytes := make([]byte, 13)
              binary.LittleEndian.PutUint32(msgBytes[0:4], helpers.GetProtocolId())
              msgBytes[4] = byte(WELCOME)
              binary.LittleEndian.PutUint64(msgBytes[5:13], uint64(clientSalt ^ serverSalt))
              player.SendRepeating(msgBytes, 500, 20)
              player.SetState(SPECTATING)
            } else {
              log.Printf("%s failed challenge with %d", clientSalt ^ serverSalt)
            }
          }

          return
        }

        // state == CONNECTED
        player.Unpack(bytes[4:])
      }
    })
  } else {
    log.Printf("Invalid IP %s", ipString)
  }

  return
}

func (ps *physicsServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
  log.Printf("UDP server is listening on %s (multi-cores: %t, loops: %d)\n",
    srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
  return
}
