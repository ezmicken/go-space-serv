package phys

import (
	"container/list"
)

type ConnectionContext struct {
	inputQueue *list.List // TODO: make this a priority queue based on seq

	// physical properties
	bod UdpBody;
	orientation float64
}

func (ctx *ConnectionContext) Init() {
	ctx.inputQueue = list.New()
}

func (ctx *ConnectionContext) AddInput(i UdpInput) {
	ctx.inputQueue.PushBack(i)
}

func (ctx *ConnectionContext) GetInput() UdpInput {
	ele := ctx.inputQueue.Front()
	ctx.inputQueue.Remove(ele)
	return ele.Value.(UdpInput)
}
