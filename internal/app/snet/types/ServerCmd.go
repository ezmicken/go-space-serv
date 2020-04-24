package snet

// one-byte indicator of server->client intent

type ServerCmd byte

const (
  SPing ServerCmd = iota + 1
  SPong
  SBlocks
  SWorldInfo
  SFrame
  SSpawn
  SSpec
  SConnectionInfo
  SSync
)
