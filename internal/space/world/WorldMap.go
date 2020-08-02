package world

import (
  "github.com/ojrac/opensimplex-go"

  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/tcp"
)

type WorldMap struct {
  W int
  H int
  Seed int64
  Resolution int
  SpawnX int
  SpawnY int

  blocks [][]BlockType
}

var viewSize int = 50

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
        wm.blocks[y][x] = SOLID
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

func (wm *WorldMap) Serialize() (out *tcp.NetworkMsg) {
  var msg tcp.NetworkMsg

  msg.PutUint32(wm.W)
  msg.PutUint32(wm.H)
  msg.PutByte(byte(wm.Resolution))

  for y := 0; y < wm.H; y++ {
    for x := 0; x < wm.W; x++ {
      msg.PutByte(byte(wm.blocks[y][x]))
    }
  }

  return &msg
}


func (wm *WorldMap) Deserialize(bytes []byte, includesDimensions bool) {
  if includesDimensions {
    wm.W = snet.Read_int32(bytes[:4])
    wm.H = snet.Read_int32(bytes[4:8])
    wm.Resolution = int(bytes[8])
  }

  wm.blocks = make([][]BlockType, wm.H)
  for i := range wm.blocks {
    wm.blocks[i] = make([]BlockType, wm.W)
  }

  i := 0
  for y := 0; y < wm.H; y++ {
    for x := 0; x < wm.W; x++ {
      wm.blocks[y][x] = BlockType(bytes[i])
    }
  }
}
