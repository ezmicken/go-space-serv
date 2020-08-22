package world

import (
  "math"
  "encoding/binary"

  "github.com/ojrac/opensimplex-go"

  "go-space-serv/internal/space/util"
  "go-space-serv/internal/space/world/msg"
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

func (wm *WorldMap) GetCellCenter(x, y int) (xPos, yPos float32) {
  xPos = float32(wm.Resolution * x + (wm.Resolution / 2))
  yPos = float32(wm.Resolution * y + (wm.Resolution / 2))
  return
}

func (wm *WorldMap) GetCellFromPosition(xPos, yPos float32) (x, y int) {
  x = int(math.Floor(float64(xPos / float32(wm.Resolution))))
  y = int(math.Floor(float64(yPos / float32(wm.Resolution))))
  return
}

func (wm *WorldMap) SerializeChunk(id uint16) msg.BlocksMsg {
  x := int(id) % wm.W * wm.ChunkSize
  y := int(id) / wm.W * wm.ChunkSize

  var blocksMsg msg.BlocksMsg
  blocksMsg.Id = id

  currentState := wm.blocks[y][x]
  var newState BlockType
  currentCount := 0
  blocksMsg.Data = []byte{byte(currentState)}

  for yi := 0; yi < wm.ChunkSize; yi++ {
    for xi := 0; xi < wm.ChunkSize; xi++ {
      newState = wm.blocks[y+yi][x+xi]
      if newState == currentState {
        currentCount++
      } else {
        currentState = newState
        blocksMsg.Data = append(blocksMsg.Data, []byte{byte(currentCount), byte(currentState)}...)
        currentCount = 1
      }
    }
  }

  if currentCount > 0 {
    blocksMsg.Data = append(blocksMsg.Data, byte(currentCount))
  }

  return blocksMsg
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
