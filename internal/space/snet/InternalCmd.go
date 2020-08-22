package snet

// one-byte indicator of server->server intent

type InternalCmd byte

const(
  IReady InternalCmd = iota + 1
  IJoin
  ILeave
  ISpec
  ISpawn
  IState
  IShutdown
)
