package world

import(
  "encoding/binary"
)

type BlockType byte

const (
  EMPTY BlockType = iota
  SOLID
  GAS
)

type Block struct {
  Type BlockType
  X, Y int
}

func (b *Block) Serialize() (out []byte) {
  var xBytes = make([]byte, 4)
  var yBytes = make([]byte, 4)

  binary.LittleEndian.PutUint32(xBytes, uint32(b.X))
  binary.LittleEndian.PutUint32(yBytes, uint32(b.Y))

  out = append([]byte{}, byte(b.Type))
  out = append(out, xBytes...)
  out = append(out, yBytes...)

  return
}
