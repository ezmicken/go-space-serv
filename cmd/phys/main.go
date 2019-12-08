package main

import (
  "log"
  "bytes"
  "encoding/binary"
  "sync"
  "time"
  //"runtime/debug"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool"

  //"github.com/bxcodec/saint" // integer math

  //"go-space-serv/internal/app/phys"
  "go-space-serv/internal/app/net"
  . "go-space-serv/internal/app/phys/types"
  //. "go-space-serv/internal/app/net/types"
)

type physicsServer struct {
  *gnet.EventServer
  pool            	*pool.WorkerPool
  tick             	time.Duration
  connectedSockets 	sync.Map
  bodies 						[]Body
  seq								byte
}

// protocol constants
// message structure is as follows:
// [ length, command, content ]
//     4b       1b    <= 4091b
const prefixLen int = 4
const cmdLen int = 1
const maxMsgSize int = 1500

var udpPort = "udp://:9495";

func read_int(data []byte) (ret int) {
  buf := bytes.NewBuffer(data)
  binary.Read(buf, binary.LittleEndian, &ret)
  return
}

func (ps *physicsServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
  log.Printf("[%s] o", c.RemoteAddr().String())
  ps.connectedSockets.Store(c.RemoteAddr().String(), c)

  var ctx ConnectionContext
  ctx.Init()
  c.SetContext(ctx)

  return
}

func (ps *physicsServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
  log.Printf("[%s] x", c.RemoteAddr().String())
  ps.connectedSockets.Delete(c.RemoteAddr().String())
  return
}

func (ps *physicsServer) React(c gnet.Conn) (out []byte, action gnet.Action) {
  data := append([]byte{}, c.Read()...)
  _ = ps.pool.Submit(func() {
    if len(data) >= 4 {
      msg := net.GetNetworkMsgFromData(data)
      if msg != nil {
        // add inputs to the input buffer
        c.ResetBuffer();
      }
    }
  })

  return
}

func (ps *physicsServer) Tick() (delay time.Duration, action gnet.Action) {
  // update sequence -- go automatically roll it over from 255 -> 0
  ps.seq++;

	// initialize the frame
	//frame := UdpFrame.New(seq)

	// gather inputs and update frame
	// ps.connectedSockets.Range(func(key, value interface{}) {

	// })
  ps.connectedSockets.Range(func(key, value interface{}) bool {
    //addr := key.(string)
    //c := value.(gnet.Conn)
    //ctx := c.Context().(ConnectionContext)

    // consume an input from the buffer

    // numMsgs := ctx.NumMsgs()
    // if numMsgs > 0 {
    //   tcpMsg := ctx.GetMsg();
    //   msgBytes := getDataFromUdpMsg(&tcpMsg);
    //   log.Printf("[%s] Tick <- %d", addr, tcpMsg.Cmd)
    //   c.AsyncWrite(msgBytes)
    // }
    return true
  })

  delay = ps.tick
  return
}

func main() {
  p := pool.NewWorkerPool()
  defer p.Release()

  ps := &physicsServer{
  	pool: p,
  	tick: 8333333,
  	seq: 0,
  	bodies: []Body{},
  }

  log.Printf("Listening to UDP on port %s", udpPort)
  log.Fatal(gnet.Serve(ps, udpPort, gnet.WithMulticore(true), gnet.WithTicker(true)))
}
