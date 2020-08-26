package world

import (
  //"log"
  "math"
  "encoding/binary"

  "github.com/ojrac/opensimplex-go"
  "github.com/akavel/polyclip-go"

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
  Poly polyclip.Polygon

  blocks [][]BlockType
}

var viewSize int = 16

func (wm *WorldMap) Generate() {
  w := float64(wm.W)
  h := float64(wm.H)

  wm.Poly = polyclip.Polygon{{
    {0, 0},
    {w, 0},
    {w, h},
    {0, h},
  }}

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

func (wm *WorldMap) serializeChunk(x, y int, id uint16) msg.BlocksMsg {
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

func (wm *WorldMap) SerializeChunks(poly polyclip.Polygon) []msg.BlocksMsg {
  numContours := len(poly)
  msgs := []msg.BlocksMsg{}
  for i := 0; i < numContours; i++ {
    bb := clampToChunks(poly[i].BoundingBox())
    pY := bb.Min.Y;
    pX := bb.Min.X;
    chunkId := wm.chunkIdFromPoint(bb.Min)
    for pY < bb.Max.Y {
      for pX < bb.Max.X {
        msgs = append(msgs, wm.serializeChunk(int(pX), int(pY), chunkId))
        chunkId++
        pX += 16
      }
      pX = bb.Min.X
      pY += 16
      chunkId = wm.chunkIdFromPoint(polyclip.Point{pX, pY})
    }
  }
  return msgs
}

func (wm *WorldMap) chunkIdFromPoint(point polyclip.Point) uint16 {
  return uint16(point.Y * float64(wm.W)) + uint16(point.X)
}
func clampTo16(val int) int {
  return (val + 8) &^ 0xF
}

func clampToChunks(rect polyclip.Rectangle) polyclip.Rectangle {
  rect.Min.X = float64(clampTo16(int(rect.Min.X)))
  rect.Min.Y = float64(clampTo16(int(rect.Min.Y)))
  rect.Max.X = float64(clampTo16(int(rect.Max.X)))
  rect.Max.Y = float64(clampTo16(int(rect.Max.Y)))
  return rect
}

func (wm *WorldMap) SerializeChunk(id uint16) msg.BlocksMsg {
  x := int(id) % wm.W * wm.ChunkSize
  y := int(id) / wm.W * wm.ChunkSize

  return wm.serializeChunk(x, y, id)
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
