package main

import (
  "log"
  //"bytes"
  "encoding/binary"
  "sync"
  "time"
  //"runtime/debug"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool"

  //"github.com/akavel/polyclip-go"

  //"github.com/bxcodec/saint"

  "go-space-serv/internal/app/world"
  "go-space-serv/internal/app/net"
  . "go-space-serv/internal/app/net/types"
  . "go-space-serv/internal/app/world/types"
)

type worldServer struct {
  *gnet.EventServer
  pool            *pool.WorkerPool
  tick             time.Duration
  connectedSockets sync.Map
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

var worldMap [][]BlockType = nil
var tcpPort = "tcp://:9494";
var spawnX int = 500;
var spawnY int = 500;
var mapW int = 1000;
var mapH int = 1000;
var viewSize int = 50;

func getBlocksAroundPoint(x, y int) (out []Block) {
  // TODO: optimize this using polygons
  out = []Block{}
  for xI := 0; xI < viewSize; xI++ {
    for yI := 0; yI < viewSize; yI++ {
      var b Block;
      b.Type = worldMap[yI][xI]
      b.X = xI
      b.Y = yI
      out = append(out, b)
    }
  }

  return
}

func getWorldInfoNetworkMsg() (NetworkMsg) {
  var msg NetworkMsg
  msg.Size = 9 // cmd, int, int
  msg.Data = []byte{byte(SWorldInfo)}
  var wBytes = make([]byte, 4)
  var hBytes = make([]byte, 4)

  binary.LittleEndian.PutUint32(wBytes, uint32(mapW))
  binary.LittleEndian.PutUint32(hBytes, uint32(mapH))

  msg.Data = append(msg.Data, wBytes...)
  msg.Data = append(msg.Data, hBytes...)

  return msg
}

func interpret(msg NetworkMsg, c gnet.Conn) (*NetworkMsg){
  log.Printf("[%s] -> %d", c.RemoteAddr().String(), msg.Size)
  return world.HandleCmd(msg.Data)
}

func (ws *worldServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
  log.Printf("[%s] o", c.RemoteAddr().String())
  ws.connectedSockets.Store(c.RemoteAddr().String(), c)

  var ctx ConnectionContext
  ctx.Init()

  // queue world data msg
  worldInfoTcpMsg := getWorldInfoNetworkMsg()
  ctx.AddMsg(worldInfoTcpMsg)

  // gather info about surrounding blocks and queue msgs
  blocks := getBlocksAroundPoint(spawnX, spawnY)
  ctx.InitExploredPoly(float64(spawnX - viewSize), float64(spawnX + viewSize), float64(spawnY - viewSize), float64(spawnY + viewSize))

  blocksLen := len(blocks)
  blocksMsgs := []NetworkMsg{}
  currentMsg := new(NetworkMsg)
  currentMsg.Data = []byte{byte(SBlocks)}
  for i, j := 0, 0; i < blocksLen; i++ {
    if j >= maxBlocksPerMsg {
      currentMsg.Size = len(currentMsg.Data)
      blocksMsgs = append(blocksMsgs, *currentMsg)
      currentMsg = new(NetworkMsg)
      currentMsg.Data = []byte{byte(SBlocks)}
      j = 0
    }
    currentMsg.Data = append(currentMsg.Data, blocks[i].Serialize()...)
    j++
  }

  ctx.AddMsgs(blocksMsgs)
  c.SetContext(ctx)

  return
}

func (ws *worldServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
  log.Printf("[%s] x", c.RemoteAddr().String())
  ws.connectedSockets.Delete(c.RemoteAddr().String())
  return
}

func (ws *worldServer) React(c gnet.Conn) (out []byte, action gnet.Action) {
  data := append([]byte{}, c.Read()...)
  _ = ws.pool.Submit(func() {
    if len(data) >= 4 {
      msg := net.GetNetworkMsgFromData(data)
      if msg != nil {
        responseMsg := interpret(*msg, c)
        c.ResetBuffer();

        if responseMsg != nil {
          response := net.GetDataFromNetworkMsg(responseMsg)
          if response != nil {
            log.Printf("[%s] React <- %d", c.RemoteAddr().String(), responseMsg.Size)
            c.AsyncWrite(response)
          }
        }
      }
    }
  })

  return
}

func (ws *worldServer) Tick() (delay time.Duration, action gnet.Action) {
  ws.connectedSockets.Range(func(key, value interface{}) bool {
    addr := key.(string)
    c := value.(gnet.Conn)
    ctx := c.Context().(ConnectionContext)

    numMsgs := ctx.NumMsgs()
    if numMsgs > 0 {
      netMsg := ctx.GetMsg();
      msgBytes := net.GetDataFromNetworkMsg(&netMsg);
      log.Printf("[%s] Tick <- %d", addr, netMsg.Size)
      c.AsyncWrite(msgBytes)
    }
    return true
  })

  delay = ws.tick
  return
}

func main() {
  log.Printf("Generating worldMap...");
  worldMap = world.GenerateMap(209323094, mapW, mapH)

  p := pool.NewWorkerPool()
  defer p.Release()

  ws := &worldServer{pool: p, tick: 200000000}
  log.Printf("Listening to TCP on port %s", tcpPort)
  log.Fatal(gnet.Serve(ws, tcpPort, gnet.WithMulticore(true), gnet.WithTicker(true)))
}
