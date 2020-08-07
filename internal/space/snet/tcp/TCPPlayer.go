package tcp

import (
  "log"
  "github.com/teris-io/shortid"
  "github.com/akavel/polyclip-go"
  "github.com/panjf2000/gnet"

  "go-space-serv/internal/space/player"
)

type TCPPlayerState byte

const (
  DISCONNECTED TCPPlayerState = iota
  AUTH
  CONNECTED
)

type TCPPlayer struct {
  explored        polyclip.Polygon
  playerId        string
  stats           *player.PlayerStats
  connection      gnet.Conn
  factory         TCPMsgFactory
  state           TCPPlayerState
  toClient chan   TCPMsg
  toWorld  chan   TCPMsg
}

var sid *shortid.Shortid

func NewTCPPlayer(conn gnet.Conn) *TCPPlayer {
  var p TCPPlayer
  p.toClient = make(chan TCPMsg, 100)
  p.stats = player.NewPlayerStats(); // TODO: get this from db
  p.connection = conn
  p.state = DISCONNECTED

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

  return &p
}

func (p *TCPPlayer) AddMsg(m TCPMsg) {
  select {
  case p.toClient <- m:
  default:
    log.Printf("%s msg chan full. Discarding...", p.playerId)
  }
}

func (p *TCPPlayer) SetWorldChan(c chan TCPMsg) {
  p.toWorld = c
}

func (p *TCPPlayer) SetMsgFactory(f TCPMsgFactory) {
  p.factory = f
}

func (p *TCPPlayer) GetPlayerId() string {
  return p.playerId
}

func (p *TCPPlayer) GetState() TCPPlayerState {
  return p.state
}

func (p *TCPPlayer) GetPlayerStats() *player.PlayerStats {
  return p.stats;
}

func (p *TCPPlayer) InitExploredPoly(xMin, xMax, yMin, yMax float64) {
  p.explored = polyclip.Polygon{{{X: xMin, Y: yMin}, {X: xMax, Y: yMin}, {X: xMax, Y: yMax}, {X: xMax, Y: yMin}}}
}

func (p *TCPPlayer) GetConnection() gnet.Conn {
  return p.connection
}
