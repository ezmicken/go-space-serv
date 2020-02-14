package world

import(
	"log"

	. "go-space-serv/internal/app/snet/types"
)

type cmdMapInstance map[ClientCmd]func(input []byte)(*NetworkMsg)

var cmdMap cmdMapInstance

func handleClientPing(input []byte)(*NetworkMsg) {
	return &NetworkMsg{Size: 1, Data: []byte{byte(SPong)}}
}

func handleClientPong(input []byte)(*NetworkMsg) {
	return nil
}

func getCmdMap() cmdMapInstance {
	if cmdMap == nil {
		log.Printf("Creating client command map")
		cmdMap = make(map[ClientCmd]func(input []byte)(*NetworkMsg))

		cmdMap[0] 		= nil
		cmdMap[CPing] = handleClientPing
		cmdMap[CPong] = handleClientPong
	}

	return cmdMap
}

func HandleCmd(input []byte) (*NetworkMsg) {
	fn := getCmdMap()[ClientCmd(input[0])]

	if fn != nil {
		return fn(input)
	}

	return nil
}
