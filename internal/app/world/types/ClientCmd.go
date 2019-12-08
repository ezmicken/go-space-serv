package world

type ClientCmd byte

const (
	CPing ClientCmd = iota + 1
	CPong
)
