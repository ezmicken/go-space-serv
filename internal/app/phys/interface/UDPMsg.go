package phys

import(
  . "go-space-serv/internal/app/phys/types"
)

type UDPMsg interface {
  GetCmd() UDPCmd
  GetSize() int
  Serialize([]byte)
  Deserialize(bytes []byte)
}
