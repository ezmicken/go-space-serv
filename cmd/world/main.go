package main

import (
  "log"
  "sync"
  "time"
  "net"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool/goroutine"

  //"github.com/akavel/polyclip-go"

  // integer math
  //"github.com/bxcodec/saint"

  "go-space-serv/internal/app/world"
  "go-space-serv/internal/app/snet"
  . "go-space-serv/internal/app/snet/types"
  . "go-space-serv/internal/app/world/types"
)

type worldServer struct {
  *gnet.EventServer
  pool            *goroutine.Pool
  tick             time.Duration
  connectedSockets sync.Map
  state            ServerState
  physics          gnet.Conn
  physicsIP        net.IP
  physicsPort      int
  worldMap         *world.WorldMap
}

// protocol constants
// message structure is as follows:
// [ length, command, content ]
//     4b       1b    <= 4091b
const prefixLen int = 4
const cmdLen int = 1
const blockLen int = 9
const maxMsgSize int = 4096
const maxBlocksPerMsg = (maxMsgSize - prefixLen - cmdLen) / blockLen

var tcpPort = "tcp://:9494";
var spawnX int = 500;
var spawnY int = 500;
var viewSize = 50; // todo: make this a player stat

func (ws *worldServer) propagateMsg(playerId string, msg *NetworkMsg) {
  ws.connectedSockets.Range(func(key, value interface{}) bool {
    c := value.(gnet.Conn)
    ctx := c.Context().(ConnectionContext)

    if ctx.GetPlayerId() != playerId {
      ctx.AddMsg(msg);
    }

    return true
  })
}

func (ws *worldServer) initPlayerConnection(c gnet.Conn) {
  var ctx ConnectionContext
  ctx.Init()
  playerId := ctx.GetPlayerId()

  // Let other clients know
  joinMsg := new(NetworkMsg)
  joinMsg.PutByte(byte(SPlayerJoin))
  joinMsg.PutString(playerId)
  stats := ctx.GetPlayerStats()
  joinMsg.PutFloat32(stats.Thrust)
  joinMsg.PutFloat32(stats.MaxSpeed)
  joinMsg.PutFloat32(stats.Rotation)
  ws.propagateMsg(playerId, joinMsg)

  // Let this client know about connected players
  ws.connectedSockets.Range(func(key, value interface{}) bool {
    connected := value.(gnet.Conn)
    connectedCtx := connected.Context().(ConnectionContext)

    if connectedCtx.GetPlayerId() != playerId {
      connectedMsg := new(NetworkMsg)
      connectedMsg.PutByte(byte(SPlayerJoin))
      connectedMsg.PutString(playerId)
      stats = connectedCtx.GetPlayerStats()
      connectedMsg.PutFloat32(stats.Thrust)
      connectedMsg.PutFloat32(stats.MaxSpeed)
      connectedMsg.PutFloat32(stats.Rotation)
      ctx.AddMsg(connectedMsg)
    }

    return true
  })

  // Let this client know their player stats
  statsMsg := new(NetworkMsg)
  statsMsg.PutByte(byte(SPlayerStats))
  statsMsg.PutString(ctx.GetPlayerId())
  statsMsg.PutFloat32(stats.Thrust)
  statsMsg.PutFloat32(stats.MaxSpeed)
  statsMsg.PutFloat32(stats.Rotation)
  ctx.AddMsg(statsMsg);

  // queue world data msg
  worldInfoNetworkMsg := ws.worldMap.SerializeInfo()
  ctx.AddMsg(worldInfoNetworkMsg)

  // gather info about surrounding blocks and queue msgs
  blocks := ws.worldMap.GetBlocksAroundPoint(spawnX, spawnY)
  ctx.InitExploredPoly(float64(spawnX - viewSize), float64(spawnX + viewSize), float64(spawnY - viewSize), float64(spawnY + viewSize))

  blocksLen := len(blocks)
  blocksMsgs := []*NetworkMsg{}
  currentMsg := new(NetworkMsg)
  currentMsg.PutByte(byte(SBlocks));
  for i, j := 0, 0; i < blocksLen; i++ {
    if j >= maxBlocksPerMsg {
      blocksMsgs = append(blocksMsgs, currentMsg)
      currentMsg = new(NetworkMsg)
      currentMsg.PutByte(byte(SBlocks))
      j = 0
    }
    currentMsg.PutBytes(blocks[i].Serialize())
    j++
  }

  ctx.AddMsgs(blocksMsgs)
  c.SetContext(ctx)

  ws.connectedSockets.Store(c.RemoteAddr().String(), c)

  // notify physics server
  var physMsg NetworkMsg
  physMsg.PutByte(byte(IJoin))
  physMsg.PutString(ctx.GetPlayerId())
  log.Printf("Sending connection event to physics");
  ws.physics.AsyncWrite(snet.GetDataFromNetworkMsg(&physMsg))

  // send physics connection details to this client
  var physConnectionMsg NetworkMsg
  physConnectionMsg.PutByte(byte(SConnectionInfo))
  physConnectionMsg.PutBytes(ws.physicsIP)
  physConnectionMsg.PutUint32(ws.physicsPort)
  physConnectionMsg.PutString(ctx.GetPlayerId())
  ctx.AddMsg(&physConnectionMsg)

  return
}

func (ws *worldServer) closePlayerConnection(c gnet.Conn) {
  disconnectedCtx := c.Context().(ConnectionContext)
  disconnectedPlayerId := disconnectedCtx.GetPlayerId();

  var physMsg NetworkMsg
  physMsg.PutByte(byte(ILeave))
  physMsg.PutString(disconnectedCtx.GetPlayerId())
  log.Printf("Sending disconnection event to physics");
  ws.physics.AsyncWrite(snet.GetDataFromNetworkMsg(&physMsg))

  // Let other clients know about this event
  ws.connectedSockets.Range(func(key, value interface{}) bool {
    ctx := value.(gnet.Conn).Context().(ConnectionContext)
    if ctx.GetPlayerId() != disconnectedPlayerId {
      disconnectedMsg := new(NetworkMsg)
      disconnectedMsg.PutByte(byte(SPlayerLeave))
      disconnectedMsg.PutString(disconnectedPlayerId)
      ctx.AddMsg(disconnectedMsg)
    }

    return true
  })
}

func (ws *worldServer) initPhysicsConnection(c gnet.Conn) (out []byte) {
  ws.physics = c
  ws.physicsIP = snet.GetOutboundIP()
  ws.state = SETUP_PHYS
  log.Printf("Physics server connected.")

  // send physics server the map details
  msg := ws.worldMap.Serialize()
  out = snet.GetDataFromNetworkMsg(msg)
  log.Printf("Sending world map (%d) to physics server...", len(out))

  return
}

func (ws *worldServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
  log.Printf("[%s] o", c.RemoteAddr().String())

  switch ws.state {
    case WAIT_PHYS:
      // check if this is the physics server
      if isPhysicsConnection(c) {
        out = ws.initPhysicsConnection(c)
      } else {
        action = gnet.Close
      }
    case SETUP_PHYS:
      // deny connections
      action = gnet.Close
    case RUNNING:
      // accept connections
      ws.initPlayerConnection(c)
    case SHUTDOWN:
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

  ws.connectedSockets.Delete(c.RemoteAddr().String())

  return
}

func (ws *worldServer) React(data []byte, c gnet.Conn) (out []byte, action gnet.Action) {
  //data := append([]byte{}, c.Read()...)
  if isPhysicsConnection(c) {
    _ = ws.pool.Submit(func() {
      if ws.state == SETUP_PHYS && len(data) >= 4 {
        if (len(data) >= 4) {
          msg := snet.GetNetworkMsgFromData(data)
          if msg != nil {
            c.ResetBuffer()
            if msg.Data[0] == byte(IReady) {
              ws.physicsPort = snet.Read_int32(msg.Data[1:])
              ws.state = RUNNING
              log.Printf("Accepting player connections...")
            }
          }
        }
      }
    })
  } else {
    _ = ws.pool.Submit(func() {
      if len(data) >= 4 {
        msg := snet.GetNetworkMsgFromData(data)
        if msg != nil {
          responseMsg := interpret(*msg, c)
          c.ResetBuffer();

          if responseMsg != nil {
            response := snet.GetDataFromNetworkMsg(responseMsg)
            if response != nil {
              log.Printf("[%s] React <- %d", c.RemoteAddr().String(), responseMsg.Size)
              c.AsyncWrite(responseMsg.SizeBytes())
              c.AsyncWrite(responseMsg.Data)
            }
          }
        }
      }
    })
  }

  return
}

func (ws *worldServer) Tick() (delay time.Duration, action gnet.Action) {
  if ws.state == RUNNING {
    ws.connectedSockets.Range(func(key, value interface{}) bool {
      addr := key.(string)
      c := value.(gnet.Conn)
      ctx := c.Context().(ConnectionContext)

      numMsgs := ctx.NumMsgs()
      if numMsgs > 0 {
        netMsg := ctx.GetMsg();
        log.Printf("[%s] Tick <- %d", addr, netMsg.Size)
        c.AsyncWrite(netMsg.SizeBytes())
        c.AsyncWrite(netMsg.Data)
      }
      return true
    })
  }
  delay = ws.tick
  return
}

func interpret(msg NetworkMsg, c gnet.Conn) (*NetworkMsg){
  log.Printf("[%s] -> %d", c.RemoteAddr().String(), msg.Size)
  return world.HandleCmd(msg.Data)
}

// TODO: find some real way to validate this.
func isPhysicsConnection(c gnet.Conn) bool {
  tcpAddr := c.RemoteAddr().(*net.TCPAddr)
  return tcpAddr.Port == 9499
}

func main() {
  log.Printf("Generating worldMap...")
  var wm world.WorldMap
  wm.W = 1000
  wm.H = 1000
  wm.Resolution = 32
  wm.Seed = 209323094
  wm.Generate()
  log.Printf("worldMap generated.")

  p := goroutine.Default()
  defer p.Release()

  ws := &worldServer{
    pool: p,
    tick: 100000000,
    state: WAIT_PHYS,
    worldMap: &wm,
  }

  log.Printf("Listening to TCP on port %s", tcpPort)
  log.Fatal(gnet.Serve(ws, tcpPort, gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true)))
}
