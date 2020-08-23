package main

import (
  "log"
  "sync"
  "time"
  "net"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool/goroutine"
  "github.com/google/uuid"

  // integer math
  //"github.com/bxcodec/saint"

  "go-space-serv/internal/space/player"
  "go-space-serv/internal/space/world"
  "go-space-serv/internal/space/world/msg"
  "go-space-serv/internal/space/snet"
)

type worldServer struct {
  *gnet.EventServer

  pool              *goroutine.Pool
  tick              time.Duration
  state             snet.ServerState

  physics           gnet.Conn
  physicsIP         net.IP
  physicsPort       uint32

  worldMap          *world.WorldMap
  players           *world.WorldPlayers
  bodyToPlayer      map[uint16]uuid.UUID
  addrToId          sync.Map

  // lifecycle
  life        chan  struct{}
  shutdown    chan  struct{}
}

const maxMsgSize int = 1024

var spawnX int = 1600;
var spawnY int = 0;

func main() {
  log.Printf("Generating worldMap...")
  var wm world.WorldMap
  wm.W = 4096
  wm.H = 4096
  wm.ChunkSize = 16
  wm.Resolution = 32
  wm.Seed = 209323094
  wm.Generate()
  log.Printf("worldMap generated.")

  var plrs world.WorldPlayers
  plrs.Count = 0

  p := goroutine.Default()
  defer p.Release()
  ws := &worldServer{
    pool: p,
    tick: 100000000,
    state: snet.WAIT_PHYS,
    worldMap: &wm,
    players: &plrs,
    bodyToPlayer: make(map[uint16]uuid.UUID),
  }

  ws.life = make(chan struct{})
  ws.shutdown = make(chan struct{})

  go ws.live()
  <-ws.life

  log.Printf("end")
}

func (ws *worldServer) live() {
  defer close(ws.life)

  go func() {
    err := gnet.Serve(ws, "tcp://:9494", gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true))
    if err != nil {
      log.Fatal(err)
    }
  }()

  <-ws.shutdown
}

func (ws *worldServer) initPlayerConnection(c gnet.Conn) {
  // TODO: auth
  // TODO: get this from db via auth token
  var stats player.PlayerStats
  stats.Thrust = 12
  stats.MaxSpeed = 20
  stats.Rotation = 172
  plr, playerId := ws.players.Add(c)

  addr := c.RemoteAddr().(*net.TCPAddr).IP
  ws.addrToId.Store(addr.String(), playerId)

  // Tell this client his stats
  var playerInfoMsg msg.PlayerInfoMsg
  playerInfoMsg.Id = playerId
  playerInfoMsg.Stats = stats
  plr.Push(&playerInfoMsg)

  // Tell this client about the world
  var worldInfoMsg msg.WorldInfoMsg
  worldInfoMsg.Size = uint32(ws.worldMap.W)
  worldInfoMsg.Res = byte(ws.worldMap.Resolution)
  plr.Push(&worldInfoMsg)

  // Tell this client about the physics server
  var simInfoMsg msg.SimInfoMsg
  simInfoMsg.Ip = ws.physicsIP
  simInfoMsg.Port = ws.physicsPort
  plr.Push(&simInfoMsg)

  // Tell the physics server about this client
  packet := []byte{byte(snet.IJoin)}
  packet = append(packet, playerId[0:]...)
  packet = append(packet, byte(len(addr)))
  packet = append(packet, addr...)
  ws.physics.AsyncWrite(packet)

  plr.Connected()

  blocksMsg := ws.worldMap.SerializeChunk(100)
  plr.Push(&blocksMsg)

  return
}

func (ws *worldServer) closePlayerConnection(c gnet.Conn) {
  id, ok := ws.addrToId.Load(c.RemoteAddr().(*net.TCPAddr).IP.String())
  if ok {
    playerId := id.(uuid.UUID)
    packet := []byte{byte(snet.ILeave)}
    packet = append(packet, playerId[0:]...)
    ws.physics.AsyncWrite(packet)

    ws.players.Remove(playerId)
  }
}

func (ws *worldServer) initPhysicsConnection(c gnet.Conn) (out []byte) {
  ws.physics = c
  ws.physicsIP = snet.GetOutboundIP()
  ws.state = snet.SETUP
  log.Printf("Physics server connected, sending blocks.")

  out = ws.worldMap.Serialize()

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
    _ = ws.pool.Submit(func(){ws.interpretPhysics(bytes)})
  }

  return
}

func (ws *worldServer) interpretPhysics(bytes []byte) {
  if ws.state == snet.SETUP && len(bytes) >= 5 && bytes[0] == byte(snet.IReady) {
    ws.physicsPort = snet.Read_uint32(bytes[1:])
    log.Printf("Accepting player connections...")
    ws.state = snet.ALIVE
  } else if ws.state >= snet.ALIVE {
    if bytes[0] == byte(snet.ISpawn) {
      bodyId := snet.Read_uint16(bytes[1:3])
      playerId, err := uuid.FromBytes(bytes[3:19])
      if err == nil {
        ws.bodyToPlayer[bodyId] = playerId
        log.Printf("%v spawned", playerId, )
      }
    } else if bytes[0] == byte(snet.ISpec) {
      bodyId := snet.Read_uint16(bytes[1:3])
      log.Printf("%v specced", ws.bodyToPlayer[bodyId])
      delete(ws.bodyToPlayer, bodyId)
    } else if bytes[0] == byte(snet.IState) {
      head := 1
      l := len(bytes)
      for head < l {
        bodyId, x, y := ws.deserializeState(bytes[head:head+6])
        head += 6

        plr := ws.players.GetPlayer(ws.bodyToPlayer[bodyId])
        if plr != nil {
          // TODO: update position and do polygon boolean stuff
          //plr.Update(x, y, ws.worldMap.Poly)
        }
      }
    }
  }
}

func (ws *worldServer) deserializeState(bytes []byte) (id, x, y uint16) {
  id = snet.Read_uint16(bytes[:2])
  x = snet.Read_uint16(bytes[2:4])
  y = snet.Read_uint16(bytes[4:6])
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

  // if ws.state == snet.ALIVE {
  //   ws.players.Range(func(key, value interface{}) bool {
  //     addr := key.(string)
  //     plr := value.(*tcp.TCPPlayer)
  //     c := plr.GetConnection()

  //     numMsgs := plr.NumMsgs()
  //     if numMsgs > 0 {
  //       netMsg := plr.GetMsg();
  //       log.Printf("[%s] Tick <- %d", addr, netMsg.Size)
  //       c.AsyncWrite(netMsg.SizeBytes())
  //       c.AsyncWrite(netMsg.Data)
  //     }
  //     return true
  //   })
  // }

  return
}

// TODO: find some real way to validate this.
func isPhysicsConnection(c gnet.Conn) bool {
  tcpAddr := c.RemoteAddr().(*net.TCPAddr)
  return tcpAddr.Port == 9499
}


