package snet

// one-byte indicator of client->server intent

type ClientCmd byte

const (
  CPing ClientCmd = iota + 1
  CPong
)
