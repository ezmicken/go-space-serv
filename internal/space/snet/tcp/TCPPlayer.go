package tcp

import (
  "log"
  "github.com/teris-io/shortid"
  "github.com/akavel/polyclip-go"
  "github.com/panjf2000/gnet"
  "container/list"

  "go-space-serv/internal/space/player"
)

type TCPPlayer struct {
  explored  polyclip.Polygon
  msgQueue  *list.List
  playerId  string
  stats     *player.PlayerStats
  c         gnet.Conn
}

var sid *shortid.Shortid

func (p *TCPPlayer) Init(conn gnet.Conn) {
  p.msgQueue = list.New()

  if sid == nil {
    tempsid, err := shortid.New(0, shortid.DefaultABC, 447)
    if err != nil {
      panic(err)
    }
    sid = tempsid
  }

  id, err2 := sid.Generate()
  if err2 != nil {
    panic(err2)
  }

  p.playerId = id
  log.Printf("generated playerId: %s", id)

  p.stats = player.NewPlayerStats(); // TODO: get this from db
  p.c = conn
}

func (p *TCPPlayer) NumMsgs() int {
  return p.msgQueue.Len()
}

func (p *TCPPlayer) GetMsg() *NetworkMsg {
  ele := p.msgQueue.Front()
  p.msgQueue.Remove(ele)
  return ele.Value.(*NetworkMsg)
}

func (p *TCPPlayer) AddMsgs(m []*NetworkMsg) {
  for i := range m {
    log.Printf("pushing %d", m[i].Size)
    p.msgQueue.PushBack(m[i])
  }
}

func (p *TCPPlayer) AddMsg(m *NetworkMsg) {
  log.Printf("pushing %d", m.Size)
  p.msgQueue.PushBack(m)
}

func (p *TCPPlayer) GetPlayerId() string {
  return p.playerId
}

func (p *TCPPlayer) GetPlayerStats() *player.PlayerStats {
  return p.stats;
}

func (p *TCPPlayer) InitExploredPoly(xMin, xMax, yMin, yMax float64) {
  p.explored = polyclip.Polygon{{{X: xMin, Y: yMin}, {X: xMax, Y: yMin}, {X: xMax, Y: yMax}, {X: xMax, Y: yMin}}}
}

func (p *TCPPlayer) GetConnection() gnet.Conn {
  return p.c
}
