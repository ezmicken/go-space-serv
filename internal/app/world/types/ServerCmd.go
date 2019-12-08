package world

type ServerCmd byte

const (
	SPing ServerCmd = iota + 1
	SPong
	SBlocks
	SWorldInfo
)
