package world

import(
  "log"

  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/tcp"
)

type cmdMapInstance map[snet.ClientCmd]func(input []byte)(*tcp.NetworkMsg)

var cmdMap cmdMapInstance

func handleClientPing(input []byte)(*tcp.NetworkMsg) {
  return &tcp.NetworkMsg{Size: 1, Data: []byte{byte(snet.SPong)}}
}

func handleClientPong(input []byte)(*tcp.NetworkMsg) {
  return nil
}

func getCmdMap() cmdMapInstance {
  if cmdMap == nil {
    log.Printf("Creating client command map")
    cmdMap = make(map[snet.ClientCmd]func(input []byte)(*tcp.NetworkMsg))

    cmdMap[0]     = nil
    cmdMap[snet.CPing] = handleClientPing
    cmdMap[snet.CPong] = handleClientPong
  }

  return cmdMap
}

func HandleCmd(input []byte) (*tcp.NetworkMsg) {
  fn := getCmdMap()[snet.ClientCmd(input[0])]

  if fn != nil {
    return fn(input)
  }

  return nil
}
