package main

import (
  "flag"
  "log"
  "time"
  "net"
  "bufio"
  "math/rand"
  "sync"
  "os"
  "encoding/binary"
  "runtime/pprof"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool/goroutine"
  "github.com/google/uuid"

  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/util"
  "go-space-serv/internal/space/world"
  "go-space-serv/internal/space/player"
  "go-space-serv/internal/space/sim"
)

const cmdLen            int     = 1
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
  life          chan  struct{}
  shutdown      chan  struct{}
  state               snet.ServerState

  players             sim.SimPlayers
  simulation          sim.Simulation
  msgFactory          sim.SimMsgFactory
  ipsToPlayers        sync.Map

  launchTime          int64
  tick                time.Duration   // loop speed
  pool                *goroutine.Pool

  // World map data
  worldConn           *net.TCPConn
  worldLaddr          *net.TCPAddr
  worldRaddr          *net.TCPAddr
  worldConnOpen       bool
  worldMap            world.WorldMap
  worldMapBytes       []byte
  worldMapLen         int
  toWorld       chan  []byte
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
  p := goroutine.Default()
  defer p.Release()

  flag.Parse()
  if *cpuprofile != "" {
    f, err := os.Create(*cpuprofile)
    if err != nil {
      log.Fatal(err)
    }
    defer f.Close()
    if err = pprof.StartCPUProfile(f); err != nil {
      log.Fatal("could not start CPU profile: ", err)
    }
    defer pprof.StopCPUProfile()
  }

  // Populate config
  // TODO: take this from flags
  var config helpers.Config
  config.TIMESTEP = 33
  config.TIMESTEP_NANO = 33000000
  config.NAME = "SPACE-PHYS"
  config.VERSION = "0.0.1"
  config.PROTOCOL_ID = SDBMHash(config.NAME + config.VERSION)
  config.MAX_MSG_SIZE = 1024
  config.WORLD_RATE = 8
  helpers.SetConfig(&config)

  log.Printf("PROTOCOL_ID: %d", config.PROTOCOL_ID)

  // Initialize UDP server
  ps := &physicsServer{
    pool: p,
    tick: 8333333,
    launchTime: time.Now().UnixNano(),
    state: snet.DEAD,
    worldConnOpen: false,
    toWorld: make(chan []byte, 32),
  }

  ps.worldLaddr = &net.TCPAddr{
    IP: net.ParseIP("127.0.0.1"),
    Port: 9499,
  }
  ps.worldRaddr = &net.TCPAddr{
    IP: net.ParseIP("127.0.0.1"),
    Port: 9494,
  }

  rand.Seed(time.Now().UnixNano())

  ps.life = make(chan struct{})
  ps.shutdown = make(chan struct{})

  go ps.live()
  <-ps.life

  log.Printf("end")
}

func (ps *physicsServer) live() {
  defer close(ps.life)

  ps.state = snet.WAIT_WORLD

  go ps.serve("udp://:9495")
  go ps.world(ps.worldLaddr, ps.worldRaddr)

  <-ps.shutdown
}

func (ps *physicsServer) serve(udpAddr string) {
  log.Printf("Listening to %s", udpAddr)
  err := gnet.Serve(ps, udpAddr, gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true))
  if err != nil {
    log.Fatal(err)
  }
}

// world server <-> physics server interaction
func (ps *physicsServer) world(laddr, raddr *net.TCPAddr) {
  c, err := net.DialTCP("tcp", laddr, raddr)
  if err != nil {
    log.Printf("%s", err)
    panic("couldnt find world server")
  }

  c.SetNoDelay(true)

  ps.worldConn = c
  ps.worldConnOpen = true
  ps.state = snet.SETUP

  const(
    readingSize byte = iota
    readingResolution
    readingMap
    readingEvent
  )

  var state byte = readingSize
  var numBlockBytes uint32 = 0
  var size uint32 = 0
  var mapBuf []byte = nil
  var mapBufRead uint32 = 0
  var totalBytesRead int = 0
  reader := bufio.NewReader(c)

  for ps.state <= snet.ALIVE {
    switch state {
      case readingSize:
        sizeBytes, err := reader.Peek(4)
        if err == nil {
          numBlockBytes = snet.Read_uint32(sizeBytes)
          size = helpers.Sqrt_uint32(numBlockBytes * 8)
          mapBuf = make([]byte, numBlockBytes)

          ps.worldMap.W = int(size)
          ps.worldMap.H = int(size)

          reader.Discard(4)
          totalBytesRead += 4

          state = readingResolution
          log.Printf("Read worldMap size=%d", size)
        } else {
          ps.state = snet.SHUTDOWN
          log.Printf(err.Error())
        }
      case readingResolution:
        resBytes, err := reader.Peek(1)
        if err == nil {
          ps.worldMap.Resolution = int(resBytes[0])

          reader.Discard(1)
          totalBytesRead += 1

          state = readingMap
          log.Printf("Read worldMap resolution=%d", ps.worldMap.Resolution)
        } else {
          ps.state = snet.SHUTDOWN
          log.Printf(err.Error())
        }
      case readingMap:
        log.Printf("reading map bytes")
        for b, err := reader.Peek(1); err == nil && mapBufRead < numBlockBytes; mapBufRead++ {
          mapBuf[mapBufRead] = b[0]
          totalBytesRead += 1
          reader.Discard(1)
        }
        log.Printf("total bytes %d", totalBytesRead)
        if err == nil {
          log.Printf("finished reading %d/%d blocks", mapBufRead, numBlockBytes)
          ps.worldMap.Deserialize(mapBuf)
          reader.Reset(c)

          state = readingEvent

          // Tell world about our port and that we're ready.
          portMsg := make([]byte, 5)
          portMsg[0] = byte(snet.IReady)
          binary.LittleEndian.PutUint32(portMsg[1:5], uint32(udpPort))
          c.Write(portMsg)

          // Start the simulation
          ps.simulation.Start(&ps.worldMap, &ps.players, ps.toWorld);

          // Signal that world is setup.
          ps.state = snet.ALIVE
        } else {
          log.Printf("%s", err.Error())
          ps.state = snet.SHUTDOWN
        }
      case readingEvent:
        event, err := reader.Peek(1)
        if err == nil {
          log.Printf("Received event from world %d", event)
          reader.Discard(1)

          if event[0] == byte(snet.IJoin) {
            var ipLen []byte
            var ipBytes []byte
            var idBytes []byte
            idBytes, err = reader.Peek(16)
            if err == nil {
              playerId, _ := uuid.FromBytes(idBytes)
              reader.Discard(16)
              ipLen, err = reader.Peek(1)
              if err == nil {
                reader.Discard(1)
                ipBytes, err = reader.Peek(int(ipLen[0]))
                if err == nil {
                  ip := net.IP(ipBytes)
                  reader.Discard(int(ipLen[0]))

                  var info player.Player
                  info.Id = playerId
                  info.Stats = player.DefaultPlayerStats()
                  plr := ps.players.Add(&info)
                  if plr != nil {
                    plr.SetSimChan(ps.simulation.GetPlayerChan())
                    plr.SetMsgFactory(&ps.msgFactory)
                    ps.ipsToPlayers.Store(ip.String(), playerId)
                    log.Printf("Storing %s <-> %v", ip.String(), playerId)
                  }
                }
              }
            }
          } else if event[0] == byte(snet.ILeave) {
            var idBytes []byte
            idBytes, err = reader.Peek(16)
            if err == nil {
              reader.Discard(16)
              playerId, _ := uuid.FromBytes(idBytes)
              ps.players.Remove(playerId)
              ps.simulation.RemoveControlledBody(playerId)
            }
          } else if event[0] == byte(snet.IShutdown) {
            ps.state = snet.SHUTDOWN
          }
        }

        if err != nil && err.Error() == "EOF" {
          ps.state = snet.SHUTDOWN // TODO: retry connecting to world
          return
        } else if err != nil && ps.state == snet.SHUTDOWN {
          return
        } else if err != nil {
          log.Printf("worldConn err: " + err.Error())
        }
    }
  }
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

// Event Handler

func (ps *physicsServer) OnShutdown(srv gnet.Server) {
  ps.state = snet.DEAD
  close(ps.shutdown)
}

func (ps *physicsServer) Tick() (delay time.Duration, action gnet.Action) {
  delay = ps.tick
  if ps.state == snet.SHUTDOWN {
    action = gnet.Shutdown
  }
  return
}

func (ps *physicsServer) React(data []byte, connection gnet.Conn) (out []byte, action gnet.Action) {
  valid := true
  ipString := connection.RemoteAddr().(*net.UDPAddr).IP.String()
  valid = valid && ps.ipIsValid(ipString)

  if valid {
    bytes := data
    _ = ps.pool.Submit(func() {
      var plr *udp.UDPPlayer
      playerId, ok := ps.ipsToPlayers.Load(ipString)
      if ok {
        plr = ps.players.GetPlayer(playerId.(uuid.UUID))
      }

      valid = valid && snet.Read_uint32(bytes[0:4]) == helpers.GetProtocolId()
      if valid && plr != nil {
        plr.PacketReceive(bytes, connection)
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
