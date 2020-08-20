package main

import (
  "log"
  "sync"
  "time"
  "net"
  "context"
  "os"
  "os/signal"

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
  addrToId          sync.Map

  // lifecycle
  ctx               context.Context
  lifeWG            sync.WaitGroup
  life        chan  struct{}
  cancel            func()
  physicsOpen       bool
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
    physicsOpen: false,
  }

  ws.ctx, ws.cancel = context.WithCancel(context.Background())
  ws.life = make(chan struct{})
  go ws.sig()
  go ws.live()

  <-ws.ctx.Done()
  log.Printf("Shutting down...")
  ws.state = snet.SHUTDOWN
  <-ws.life

  log.Printf("end")
}

func (ws *worldServer) live() {
  ws.lifeWG.Add(1)
  go func() {
    err := gnet.Serve(ws, "tcp://:9494", gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true))
    if err != nil {
      ws.lifeWG.Done()
      ws.cancel()
      log.Fatal(err)
    }
    log.Printf("server finish")
  }()

  ws.lifeWG.Wait()
  ws.state = snet.DEAD
  ws.cancel()
  close(ws.life)
}

func (ws *worldServer) sig() {
  signalChan := make(chan os.Signal, 1)
  signal.Notify(signalChan, os.Interrupt)

  oscall := <-signalChan
  log.Printf("system call: %+v", oscall)
  ws.cancel()
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
  ws.physicsOpen = true
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
  } else {
    ws.physicsOpen = false
  }

  return
}

func (ws *worldServer) React(data []byte, c gnet.Conn) (out []byte, action gnet.Action) {
  bytes := append([]byte{}, data...)
  c.ResetBuffer()
  if isPhysicsConnection(c) {
    _ = ws.pool.Submit(func() {
      if ws.state == snet.SETUP && len(bytes) >= 5 && bytes[0] == byte(snet.IReady) {
        ws.physicsPort = snet.Read_uint32(bytes[1:])
        log.Printf("Accepting player connections...")
        ws.state = snet.ALIVE
      } else {
        // _ = ws.pool.Submit(func() {
        //   if len(data) >= 4 {
        //     msg := snet.GetNetworkMsgFromData(data)
        //     if msg != nil {
        //       responseMsg := interpret(*msg, c)
        //       c.ResetBuffer();

        //       if responseMsg != nil {
        //         response := snet.GetDataFromNetworkMsg(responseMsg)
        //         if response != nil {
        //           log.Printf("[%s] React <- %d", c.RemoteAddr().String(), responseMsg.Size)
        //           c.AsyncWrite(responseMsg.SizeBytes())
        //           c.AsyncWrite(responseMsg.Data)
        //         }
        //       }
        //     }
        //   }
        // })
      }
    })
  }

  return
}

func (ws *worldServer) OnShutdown(s gnet.Server) {
  ws.lifeWG.Done()
}

func (ws *worldServer) Tick() (delay time.Duration, action gnet.Action) {
  delay = ws.tick

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


