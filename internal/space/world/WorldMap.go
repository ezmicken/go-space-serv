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
  chunker *Chunker
  sizeInBlocks float64

  Poly polyclip.Polygon
}

var viewSize int = 16

func NewWorldMap(name string) (*WorldMap, error) {
  metaFile, err := os.Open(fmt.Sprintf("assets/%s/meta.chunks", name))
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
  wm.info.Name = name
  wm.sizeInBlocks = float64(wm.info.Size * wm.info.ChunkSize)

  sizeInBlocks := float64(wm.info.Size * wm.info.ChunkSize)
  wm.Poly = polyclip.Polygon{{
    {0, 0},
    {0, sizeInBlocks},
    {sizeInBlocks, sizeInBlocks},
    {sizeInBlocks, 0},
  }}

  wm.chunker = NewChunker(wm.info)

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

  fileId := uint16(math.Floor(float64(uint32(id) / wm.info.ChunksPerFile)))

  serializedChunk := wm.chunker.GetChunk(id, fileId)
  blocksMsg.Data = append([]byte{}, serializedChunk...)

  return blocksMsg
}

// Assumes bb is clamped to chunks
func (wm *WorldMap) Explore(bb polyclip.Rectangle) []msg.BlocksMsg {
  msgs := []msg.BlocksMsg{}

  log.Printf("exploring: %v", bb)
  pY := bb.Min.Y;
  pX := bb.Min.X;
  chunkId := wm.chunkIdFromPoint(polyclip.Point{pX, pY})

  for pY < bb.Max.Y {
    for pX < bb.Max.X {
      msgs = append(msgs, wm.serializeChunk(int(pX), int(pY), chunkId))
      chunkId++
      pX += float64(wm.info.ChunkSize)
    }
    pX = bb.Min.X
    pY += float64(wm.info.ChunkSize)
    chunkId = wm.chunkIdFromPoint(polyclip.Point{pX, pY})
  }

  return msgs
}

func (wm *WorldMap) chunkIdFromPoint(point polyclip.Point) uint16 {
  chunkX := point.X/float64(wm.info.ChunkSize)
  chunkY := point.Y/float64(wm.info.ChunkSize)
  return uint16(chunkY * float64(wm.info.Size)) + uint16(chunkX)
}

func (wm *WorldMap) ClampToChunks(rect polyclip.Rectangle) polyclip.Rectangle {
  // expand rectangle to multiple of chunk size.
  rect.Min.X = float64((int(rect.Min.X) - 64) &^ 0x7F)
  rect.Min.Y = float64((int(rect.Min.Y) - 64) &^ 0x7F)
  rect.Max.X = float64((int(rect.Max.X) + 64) &^ 0x7F)
  rect.Max.Y = float64((int(rect.Max.Y) + 64) &^ 0x7F)

  // clamp to world bounds
  if rect.Min.X < 0 { rect.Min.X = 0 }
  if rect.Min.Y < 0 { rect.Min.Y = 0 }
  if rect.Max.X > wm.sizeInBlocks { rect.Max.X = wm.sizeInBlocks }
  if rect.Max.Y > wm.sizeInBlocks { rect.Max.Y = wm.sizeInBlocks }

  // Correct non-rectangles.
  if rect.Min.X == rect.Max.X {
    if rect.Min.X > 128 {
      rect.Min.X -= 128
    } else if rect.Max.X < (wm.sizeInBlocks - 128) {
      rect.Max.X += 128
    } else if rect.Min.X == 0 {
      rect.Max.X += 128
    } else if rect.Max.X == wm.sizeInBlocks {
      rect.Min.X -= 128
    } else {
      log.Printf("Unable to correct x of rectangle %v", rect)
    }
  }

  if rect.Min.Y == rect.Max.Y {
    if rect.Min.Y > 128 {
      rect.Min.Y -= 128
    } else if rect.Max.Y < (wm.sizeInBlocks - 128) {
      rect.Max.Y += 128
    } else if rect.Min.Y == 0 {
      rect.Max.Y += 128
    } else if rect.Max.Y == wm.sizeInBlocks {
      rect.Min.Y -= 128
    } else {
      log.Printf("Unable to correct y of rectangle %v", rect)
    }
  }

  return rect
}

func (wm *WorldMap) SerializeChunk(id uint16) msg.BlocksMsg {
  x := int(id) % int(wm.info.Size * wm.info.ChunkSize)
  y := int(id) / int(wm.info.Size * wm.info.ChunkSize)

  return wm.serializeChunk(x, y, id)
}

func (wm *WorldMap) GetWorldInfoMsg() msg.WorldInfoMsg {
  var worldInfoMsg msg.WorldInfoMsg
  worldInfoMsg.ChunksPerFile = wm.info.ChunksPerFile
  worldInfoMsg.ChunkSize = wm.info.ChunkSize
  worldInfoMsg.Size = wm.info.Size
  worldInfoMsg.Seed = wm.info.Seed
  worldInfoMsg.Threshold = wm.info.Threshold
  return worldInfoMsg
}

func (wm *WorldMap) GetSpawnPoint() (x, y float32) {
  return wm.GetCellCenter(int(SPAWNX), int(SPAWNY))
}
