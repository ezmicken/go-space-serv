package world

import (
  "log"
  "fmt"
  "math"
  "os"
  "errors"

  "github.com/akavel/polyclip-go"

  "go-space-serv/internal/space/world/msg"
)

// TODO: move to worldinfo
const RESOLUTION float32 = 32
const SPAWNX uint32 = 1600
const SPAWNY uint32 = 0

type WorldMap struct {
  info WorldInfo

  Poly polyclip.Polygon
}

var viewSize int = 16

func NewWorldMap(name string) (*WorldMap, error) {
  metaFile, err := os.OpenFile(fmt.Sprintf("assets/%s/meta.chunks", name), os.O_RDONLY, 0644)
  if err != nil {
    return nil, err
  }

  stat, err2 := metaFile.Stat()
  if err != nil {
    return nil, err2
  }
  metaFileSize := stat.Size()
  bytes := make([]byte, metaFileSize)
  bytesRead, err3 := metaFile.Read(bytes)
  if err3 != nil {
    return nil, err3
  }

  if int64(bytesRead) != metaFileSize {
    return nil, errors.New(fmt.Sprintf("failed to read meta file %d/%d", bytesRead, metaFileSize))
  }

  var wm WorldMap
  wm.info = DeserializeWorldInfo(bytes)

  sizeInBlocks := float64(wm.info.Size * wm.info.ChunkSize)
  wm.Poly = polyclip.Polygon{{
    {0, 0},
    {0, sizeInBlocks},
    {sizeInBlocks, sizeInBlocks},
    {sizeInBlocks, 0},
  }}

  log.Printf("Loaded map %s\n%v\n", name, wm.info)

  return &wm, nil
}

func (wm *WorldMap) GetCellCenter(x, y int) (xPos, yPos float32) {
  halfRes := RESOLUTION / 2
  xPos = RESOLUTION * float32(x) + halfRes
  yPos = RESOLUTION * float32(y) + halfRes
  return
}

func (wm *WorldMap) GetCellFromPosition(xPos, yPos float32) (x, y int) {
  x = int(math.Floor(float64(xPos / float32(RESOLUTION))))
  y = int(math.Floor(float64(yPos / float32(RESOLUTION))))
  return
}

func (wm *WorldMap) serializeChunk(x, y int, id uint16) msg.BlocksMsg {
  var blocksMsg msg.BlocksMsg
  blocksMsg.Id = id

  currentState := EMPTY//wm.blocks[y][x]
  var newState BlockType
  currentCount := 0
  blocksMsg.Data = []byte{byte(currentState)}

  for yi := 0; yi < int(wm.info.ChunkSize); yi++ {
    for xi := 0; xi < int(wm.info.ChunkSize); xi++ {
      newState = EMPTY//wm.blocks[y+yi][x+xi]
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
  return uint16(point.Y * float64(wm.info.Size)) + uint16(point.X)
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
  x := int(id) % int(wm.info.Size * wm.info.ChunkSize)
  y := int(id) / int(wm.info.Size * wm.info.ChunkSize)

  return wm.serializeChunk(x, y, id)
}

func (wm *WorldMap) GetWorldInfoMsg() msg.WorldInfoMsg {
  var worldInfoMsg msg.WorldInfoMsg
  worldInfoMsg.Size = uint32(wm.info.Size)
  worldInfoMsg.Res = byte(RESOLUTION)
  return worldInfoMsg
}

func (wm *WorldMap) GetSpawnPoint() (x, y float32) {
  return wm.GetCellCenter(int(SPAWNX), int(SPAWNY))
}
