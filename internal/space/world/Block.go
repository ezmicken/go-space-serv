package world

type BlockType byte

const (
  EMPTY BlockType = iota
  GRAY
)

type Block struct {
  Type BlockType
  X, Y int
}
