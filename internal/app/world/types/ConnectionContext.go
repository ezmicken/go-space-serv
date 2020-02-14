package world

import (
	"github.com/teris-io/shortid"
	"github.com/akavel/polyclip-go"
	"container/list"

	. "go-space-serv/internal/app/snet/types"
	"log"
)

type ConnectionContext struct {
	explored polyclip.Polygon
	msgQueue *list.List
	playerId string
}

var sid *shortid.Shortid

func (ctx *ConnectionContext) Init() {
	ctx.msgQueue = list.New()

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

	ctx.playerId = id
	log.Printf("generated playerId: %s", id)
}

func (ctx *ConnectionContext) NumMsgs() int {
	return ctx.msgQueue.Len()
}

func (ctx *ConnectionContext) GetMsg() *NetworkMsg {
	ele := ctx.msgQueue.Front()
	ctx.msgQueue.Remove(ele)
	return ele.Value.(*NetworkMsg)
}

func (ctx *ConnectionContext) AddMsgs(m []*NetworkMsg) {
	for i := range m {
		log.Printf("pushing %d", m[i].Size)
		ctx.msgQueue.PushBack(m[i])
	}
}

func (ctx *ConnectionContext) AddMsg(m *NetworkMsg) {
	log.Printf("pushing %d", m.Size)
	ctx.msgQueue.PushBack(m)
}

func (ctx *ConnectionContext) GetPlayerId() string {
	return ctx.playerId
}

func (ctx *ConnectionContext) InitExploredPoly(xMin, xMax, yMin, yMax float64) {
	ctx.explored = polyclip.Polygon{{{X: xMin, Y: yMin}, {X: xMax, Y: yMin}, {X: xMax, Y: yMax}, {X: xMax, Y: yMin}}}
}
