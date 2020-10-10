package main

import (
  "log"
  "fmt"
  "flag"
  "sync"
  "time"
  "net"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool/goroutine"
  "github.com/google/uuid"

  "go-space-serv/internal/space/world"
  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/tcp"
)

type worldServer struct {
  *gnet.EventServer

  pool              *goroutine.Pool
  tick              time.Duration
  state             snet.ServerState

  physics           gnet.Conn
  physicsIP         net.IP
  physicsPort       uint32

  wld               *world.World
  players           *world.WorldPlayers
  msgFactory        world.WorldMsgFactory
  addrToId          sync.Map

  // lifecycle
  life        chan  struct{}
  shutdown    chan  struct{}
}

const maxMsgSize int = 1024

var spawnX uint16 = 1600;
var spawnY uint16 = 0;

func main() {
  flagPort := flag.Uint("port", 9494, "Port to listen on")

  flag.Parse()

  mapName := flag.Arg(0)

  if mapName == "" { mapName = "localMap" }

  var plrs world.WorldPlayers
  plrs.Count = 0

  w, err := world.NewWorld(&plrs, mapName)
  if err != nil {
    panic(err)
  }
  p := goroutine.Default()
  defer p.Release()

  ws := &worldServer{
    wld: w,
    pool: p,
    tick: 100000000,
    state: snet.WAIT_PHYS,
    players: &plrs,
    life: make(chan struct{}),
    shutdown: make(chan struct{}),
  }

  go ws.live(*flagPort)
  <-ws.life

  log.Printf("end")
}

func (ws *worldServer) live(port uint) {
  defer close(ws.life)

  go func() {
    log.Printf("Awaiting simulation...")
    addr := fmt.Sprintf("tcp://:%d", port);
    err := gnet.Serve(ws, addr, gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true))
    if err != nil {
      log.Fatal(err)
    }
  }()

  <-ws.shutdown
}

func (ws *worldServer) initPlayerConnection(c gnet.Conn) {
  // TODO: auth
  // TODO: get this from db via auth token
  id := uuid.New()
  tcpPlr := tcp.NewPlayer(c, id, &ws.msgFactory)
  plr := ws.players.Add(tcpPlr)
  plr.X = spawnX
  plr.Y = spawnY

  addr := c.RemoteAddr().(*net.TCPAddr).IP
  ws.addrToId.Store(addr.String(), id)

  ws.wld.PlayerJoin(plr, ws.physicsIP, ws.physicsPort)

  // Tell the physics server about this client
  packet := []byte{byte(snet.IJoin)}
  packet = append(packet, id[0:]...)
  packet = append(packet, byte(len(addr)))
  packet = append(packet, addr...)
  ws.physics.AsyncWrite(packet)

  plr.Tcp.Connected()

  return
}

func (ws *worldServer) closePlayerConnection(c gnet.Conn) {
  id, ok := ws.addrToId.Load(c.RemoteAddr().(*net.TCPAddr).IP.String())
  if ok {
    playerId := id.(uuid.UUID)

    ws.wld.PlayerLeave(playerId)

    // Tell physics about this
    packet := []byte{byte(snet.ILeave)}
    packet = append(packet, playerId[0:]...)
    ws.physics.AsyncWrite(packet)
  }
}

func (ws *worldServer) initPhysicsConnection(c gnet.Conn) (out []byte) {
  ws.physics = c
  ws.physicsIP = snet.GetOutboundIP()
  ws.state = snet.SETUP
  log.Printf("Physics server connected, sending blocks.")

  return
}

func (ws *worldServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
  log.Printf("[%s] o", c.RemoteAddr().String())

  switch ws.state {
    case snet.WAIT_PHYS:
      // check if this is the physics server
      if isPhysicsConnection(c) {
        out = ws.initPhysicsConnection(c)
      } else {
        action = gnet.Close
      }
    case snet.SETUP:
      // deny connections
      action = gnet.Close
    case snet.ALIVE:
      // accept connections
      ws.initPlayerConnection(c)
    case snet.SHUTDOWN:
      // deny connections
      action = gnet.Close
    default:
      // deny connections
      action = gnet.Close
  }

  if action == gnet.Close {
    log.Printf("[%s] rejected due to state=%d", c.RemoteAddr().String(), ws.state)
  }

  return
}

func (ws *worldServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
  log.Printf("[%s] x", c.RemoteAddr().String())

  if !isPhysicsConnection(c) {
    ws.closePlayerConnection(c);
  }

  return
}

func (ws *worldServer) React(data []byte, c gnet.Conn) (out []byte, action gnet.Action) {
  bytes := append([]byte{}, data...)
  c.ResetBuffer()
  if isPhysicsConnection(c) {
    if ws.state == snet.SETUP && len(bytes) >= 5 && bytes[0] == byte(snet.IReady) {
      ws.physicsPort = snet.Read_uint32(bytes[1:])
      log.Printf("Accepting player connections...")
      ws.state = snet.ALIVE
    } else {
      _ = ws.pool.Submit(func() {
        ws.wld.InterpretPhysics(bytes)
      })
    }
  }

  return
}

func (ws *worldServer) OnShutdown(s gnet.Server) {
  ws.state = snet.DEAD
  close(ws.shutdown)
}

func (ws *worldServer) Tick() (delay time.Duration, action gnet.Action) {
  delay = ws.tick

  if ws.state == snet.SHUTDOWN {
    action = gnet.Shutdown
  }

  return
}

// TODO: find some real way to validate this.
func isPhysicsConnection(c gnet.Conn) bool {
  tcpAddr := c.RemoteAddr().(*net.TCPAddr)
  return tcpAddr.Port == 9499
}


