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
  "runtime/pprof"
  "encoding/binary"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool/goroutine"
  "github.com/google/uuid"

  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/util"
  "go-space-serv/internal/space/world"
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

  players             SimPlayers
  simulation          Simulation
  msgFactory          SimMsgFactory
  ipsToPlayers        sync.Map

  launchTime          int64
  tick                time.Duration   // loop speed
  pool                *goroutine.Pool

  // World map data
  worldConn           *net.TCPConn
  worldLaddr          *net.TCPAddr
  worldRaddr          *net.TCPAddr
  worldConnOpen       bool
  worldMap            *world.WorldMap
  worldMapBytes       []byte
  worldMapLen         int
  toWorld       chan  []byte
}



func main() {
  cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
  flagTimestep := flag.Int64("timestep", 33, "physics timestep in milliseconds")
  flagTimestepNano := flag.Int64("timestepNano", 33000000, "physics timestep in nanoseconds")
  flagProtocolId := flag.Uint("protocolId", 3551548956, "value must match client")
  flagWorldRate := flag.Int("worldRate", 12, "physics frames passing before sending state update to world.")

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

  mapName := flag.Arg(0)
  if mapName == "" { mapName = "localMap"}

  wm, err := world.NewWorldMap(mapName)
  if err != nil { panic(err) }

  // Populate config
  // TODO: take this from flags
  var config helpers.Config
  config.TIMESTEP = *flagTimestep
  config.TIMESTEP_NANO = *flagTimestepNano
  config.NAME = "SPACE-PHYS"
  config.VERSION = "0.0.1"
  config.PROTOCOL_ID = uint32(*flagProtocolId)
  config.MAX_MSG_SIZE = 1024
  config.WORLD_RATE = *flagWorldRate
  helpers.SetConfig(&config)

  log.Printf("PROTOCOL_ID: %d", config.PROTOCOL_ID)

  // Initialize UDP server
  ps := &physicsServer{
    pool: p,
    tick: 8333333,
    launchTime: time.Now().UnixNano(),
    state: snet.DEAD,
    worldConnOpen: false,
    worldMap: wm,
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
  go ps.worldRx(ps.worldLaddr, ps.worldRaddr)
  go ps.worldTx()

  <-ps.shutdown
}

func (ps *physicsServer) serve(udpAddr string) {
  log.Printf("Listening to %s", udpAddr)
  err := gnet.Serve(ps, udpAddr, gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true))
  if err != nil {
    log.Fatal(err)
  }
}

func (ps *physicsServer) worldTx() {
  for packet := range ps.toWorld {
    ps.worldConn.Write(packet)
  }
}

// world server <-> physics server interaction
func (ps *physicsServer) worldRx(laddr, raddr *net.TCPAddr) {
  c, err := net.DialTCP("tcp", laddr, raddr)
  if err != nil {
    log.Printf("%s", err)
    panic("couldnt find world server")
  }

  c.SetNoDelay(true)

  ps.worldConn = c
  ps.worldConnOpen = true
  ps.state = snet.SETUP

  reader := bufio.NewReader(c)

  // Tell world about our port and that we're ready.
  portMsg := make([]byte, 5)
  portMsg[0] = byte(snet.IReady)
  binary.LittleEndian.PutUint32(portMsg[1:5], uint32(udpPort))
  c.Write(portMsg)

  ps.simulation.Start(ps.worldMap, &ps.players, ps.toWorld)

  ps.state = snet.ALIVE

  for ps.state <= snet.ALIVE {
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

              plr := udp.NewPlayer(ps.simulation.GetPlayerChan(), playerId, &ps.msgFactory)
              if plr != nil {
                ps.ipsToPlayers.Store(ip.String(), plr)
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
          //ps.simulation.RemoveControlledBody(playerId)
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
  ps.worldConn.Close()
  close(ps.toWorld)
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
      tmp, ok := ps.ipsToPlayers.Load(ipString)
      if ok {
        plr := tmp.(*udp.UDPPlayer)
        if valid && snet.Read_uint32(bytes[0:4]) == helpers.GetProtocolId() {
          if plr.GetState() >= udp.CONNECTED {
            plr.Unpack(bytes[4:])
          } else if plr.AuthenticateConnection(bytes, connection) {
            ps.players.Add(plr)
          }
        }
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
