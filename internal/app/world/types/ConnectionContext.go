package world

import (
	"github.com/akavel/polyclip-go"
	"container/list"

	. "go-space-serv/internal/app/net/types"
)

type ConnectionContext struct {
	explored polyclip.Polygon
	msgQueue *list.List
}

func (ctx *ConnectionContext) Init() {
	ctx.msgQueue = list.New()
}

func (ctx *ConnectionContext) NumMsgs() int {
	return ctx.msgQueue.Len()
}

func (ctx *ConnectionContext) GetMsg() NetworkMsg {
	ele := ctx.msgQueue.Front()
	ctx.msgQueue.Remove(ele)
	return ele.Value.(NetworkMsg)
}

func (ctx *ConnectionContext) AddMsgs(m []NetworkMsg) {
	for i := range m {
		ctx.msgQueue.PushBack(m[i])
	}
}

func (ctx *ConnectionContext) AddMsg(m NetworkMsg) {
	ctx.msgQueue.PushBack(m)
}

func (ctx *ConnectionContext) InitExploredPoly(xMin, xMax, yMin, yMax float64) {
	ctx.explored = polyclip.Polygon{{{X: xMin, Y: yMin}, {X: xMax, Y: yMin}, {X: xMax, Y: yMax}, {X: xMax, Y: yMin}}}
}
