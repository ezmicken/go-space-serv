package world

import (
  "encoding/binary"

  "github.com/ojrac/opensimplex-go"

  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/tcp"
  "go-space-serv/internal/space/util"
)

type WorldMap struct {
  W int
  H int
  Seed int64
  Resolution int
  ChunkSize int
  SpawnX int
  SpawnY int

  blocks [][]BlockType
}

var viewSize int = 16

func (wm *WorldMap) Generate() {
  noise := opensimplex.New(wm.Seed)

  // initialize multidimensional array
  wm.blocks = make([][]BlockType, wm.H)
  for i := range wm.blocks {
    wm.blocks[i] = make([]BlockType, wm.W)
  }

  for y := 0; y < wm.H; y++ {
    for x := 0; x < wm.W; x++ {
      floatVal := noise.Eval2(float64(x) * 0.05, float64(y) * 0.05)
      if floatVal > 0.36 {
        wm.blocks[y][x] = GRAY
      } else {
        wm.blocks[y][x] = EMPTY
      }
    }
  }

  return
}

func (wm *WorldMap) GetBlock(x, y int) BlockType {
  return wm.blocks[x][y]
}

func (wm *WorldMap) GetBlocksAroundPoint(x, y int) (out []Block) {
  out = []Block{}
  for xI := x - viewSize; xI < x + viewSize; xI++ {
    for yI := y - viewSize; yI < y + viewSize; yI++ {
      var b Block;
      b.Type = wm.blocks[yI][xI]
      b.X = xI
      b.Y = yI
      out = append(out, b)
    }
  }

  return
}

func (wm *WorldMap) GetCellCenter(x, y int) (xPos, yPos float32) {
  xPos = float32(wm.Resolution * x + (wm.Resolution / 2))
  yPos = float32(wm.Resolution * y + (wm.Resolution / 2))
  return
}

func (wm *WorldMap) SerializeInfo() (out *tcp.NetworkMsg) {
  var msg tcp.NetworkMsg

  msg.PutByte(byte(snet.SWorldInfo))
  msg.PutUint32(wm.W)
  msg.PutUint32(wm.H)

  return &msg
}

func (wm *WorldMap) Serialize() []byte {
  numBlocks := wm.W * wm.H

  // map + bytes.length(uint32) + width(uint32) + height(uint32) + resolution(byte)
  bytes := make([]byte, (numBlocks/8) + 4 + 1)
  binary.LittleEndian.PutUint32(bytes[:4], uint32(numBlocks/8))
  bytes[4] = byte(wm.Resolution)

  x := 0
  y := 0
  currentByte := 5
  currentBit := 0
  for i := 0; i < numBlocks; i++ {
    if wm.blocks[y][x] != EMPTY {
      bytes[currentByte] |= (1 << currentBit)
    }

    currentBit++
    if currentBit > 7 {
      currentBit = 0
      currentByte++
    }

    x++
    if x >= wm.W {
      x = 0
      y++
    }
  }

  return bytes
}


func (wm *WorldMap) Deserialize(bytes []byte) {
  wm.blocks = make([][]BlockType, wm.H)
  for i := range wm.blocks {
    wm.blocks[i] = make([]BlockType, wm.W)
  }

  x := 0
  y := 0
  for currentByte := 0; y < wm.H; currentByte++ {
    for currentBit := 0; currentBit < 8; currentBit++ {
      if helpers.BitOn(bytes[currentByte], currentBit) {
        wm.blocks[y][x] = GRAY
      }

      x++
      if x >= wm.W {
        x = 0
        y++
      }
    }
  }
}
