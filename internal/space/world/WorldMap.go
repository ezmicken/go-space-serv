package world

import (
  "log"
  "fmt"
  "math"
  "os"
  "errors"

  "github.com/akavel/polyclip-go"
  "github.com/go-gl/mathgl/mgl32"
  "github.com/ezmicken/spacesim"

  "go-space-serv/internal/space/world/msg"
  "go-space-serv/internal/space/geom"
)

// TODO: move to worldinfo
const RESOLUTION float32 = 32
const SPAWNX uint32 = 64
const SPAWNY uint32 = 64

const NW int = 0
const NE int = 1
const SE int = 2
const SW int = 3

type WorldMap struct {
  info WorldInfo
  chunker *Chunker
  sizeInBlocks float64

  Poly polyclip.Polygon
}

type chip struct {
  blocks []byte
  box geom.RectInt
  chunkId uint16
}

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

  serializedChunk := wm.chunker.GetZippedChunk(id, fileId)
  blocksMsg.Data = append([]byte{}, serializedChunk...)

  return blocksMsg
}

func (wm *WorldMap) PushBlockRects(seq int, cb *spacesim.ControlledBody) {
  // TODO: find a better way to do this.
  xPos := cb.GetPositionX(uint16(seq)-1)
  yPos := cb.GetPositionY(uint16(seq)-1)
  playerBox := geom.NewRect(xPos.Float() - 64, yPos.Float() - 64, 128, 128)

  // Get playerBox in block coordinates
  blockMinX, blockMinY := wm.GetCellFromPosition(playerBox.X, playerBox.Y)
  blockMaxX, blockMaxY := wm.GetCellFromPosition(playerBox.X + playerBox.W, playerBox.Y + playerBox.H)

  blockMinX -= 1
  blockMinY -= 1
  blockW := blockMaxX - blockMinX + 2
  blockH := blockMaxY - blockMinY + 2
  blockBox := geom.NewRectInt(blockMinX, blockMinY, blockW, blockH)

  playerBox.X = float32(blockMinX) * RESOLUTION
  playerBox.Y = float32(blockMinY) * RESOLUTION
  playerBox.W = float32(blockW) * RESOLUTION
  playerBox.H = float32(blockH) * RESOLUTION
  playerMaxX := playerBox.X + playerBox.W
  playerMaxY := playerBox.Y + playerBox.H

  actualWorldSize := float32(wm.info.Size * wm.info.ChunkSize) * RESOLUTION

  var invalidChip chip
  invalidChip.blocks = nil
  invalidChip.box = geom.NewRectInt(0, 0, 0, 0)

  var chipNW = invalidChip
  var chipNE = invalidChip
  var chipSE = invalidChip
  var chipSW = invalidChip

  if playerBox.X >= 0 && playerMaxY <= actualWorldSize {
    chipNW = wm.getChip(blockBox, mgl32.Vec3{playerBox.X, playerMaxY, 0}, NW)
  }

  if playerMaxX <= actualWorldSize && playerMaxY <= actualWorldSize {
    chipNE = wm.getChip(blockBox, mgl32.Vec3{playerMaxX, playerMaxY, 0}, NE)
  }

  if playerMaxX <= actualWorldSize && playerBox.X >= 0 {
    chipSE = wm.getChip(blockBox, mgl32.Vec3{playerMaxX, playerBox.Y, 0}, SE)
  }

  if playerBox.X >= 0 && playerBox.Y >= 0 {
    chipSW = wm.getChip(blockBox, mgl32.Vec3{playerBox.X, playerBox.Y, 0}, SW)
  }

  blocks := make([]byte, blockBox.W * blockBox.H)
  numSolid := len(blocks)

  // Best case -- all chips cover same chunk.
  if chipNW.chunkId == chipNE.chunkId && chipNW.chunkId == chipSE.chunkId && chipNW.chunkId == chipSW.chunkId {
    actualChip := invalidChip
    if chipNW.blocks != nil { actualChip = chipNW }
    if chipNE.blocks != nil { actualChip = chipNE }
    if chipSE.blocks != nil { actualChip = chipSE }
    if chipSW.blocks != nil { actualChip = chipSW }
    if actualChip.blocks == nil {
      return
    }

    i := 0
    for y := 0; y < actualChip.box.H; y++ {
      for x := 0; x < actualChip.box.W; x++ {
        color := actualChip.blocks[(int(wm.info.ChunkSize) * (y + actualChip.box.Y)) + (x + actualChip.box.X)]
        if color <= 0 {
          numSolid--
        }
        blocks[i] = color
        i++
      }
    }
  } else {
    if chipNW.blocks != nil {
      numSolid -= wm.transformBlocks(blockBox, chipNW, blocks, NW)
    }
    if chipNE.blocks != nil && chipNE.chunkId != chipNW.chunkId {
      numSolid -= wm.transformBlocks(blockBox, chipNE, blocks, NE)
    }
    if chipSE.blocks != nil && chipSE.chunkId != chipNW.chunkId && chipSE.chunkId != chipNE.chunkId {
      numSolid -= wm.transformBlocks(blockBox, chipSE, blocks, SE)
    }
    if chipSW.blocks != nil && chipSW.chunkId != chipNW.chunkId && chipSW.chunkId != chipNE.chunkId && chipSW.chunkId != chipSE.chunkId {
      numSolid -= wm.transformBlocks(blockBox, chipSW, blocks, SW)
    }
  }

  // Create a rect in world space for each solid block
  if numSolid > 0 {
    iBlx := 0
    worldSize := int(actualWorldSize)
    for y := 0; y < blockBox.H; y++ {
      for x := 0; x < blockBox.W; x++ {
        blockIsSolid := false       || blockBox.X + x < 0
        blockIsSolid = blockIsSolid || blockBox.Y + y < 0
        blockIsSolid = blockIsSolid || blockBox.X + x > worldSize
        blockIsSolid = blockIsSolid || blockBox.Y + y > worldSize
        blockIsSolid = blockIsSolid || blocks[iBlx] > 0
        iBlx++

        if blockIsSolid {
          cb.AddBlock(int32(blockBox.X + x), int32(blockBox.Y + y))
        }
      }
    }
  }
}

func (wm *WorldMap) transformBlocks(blockBox geom.RectInt, ch chip, blocks []byte, quad int) int {
  numSolid := 0

  for y := 0; y < ch.box.H; y++ {
    for x := 0; x < ch.box.W; x++ {
      color := ch.blocks[(int(wm.info.ChunkSize) * (y + ch.box.Y)) + (x + ch.box.X)]
      if color <= 0 {
        numSolid--
      }

      oX := blockBox.W - ch.box.W
      oY := blockBox.H - ch.box.H
      idx := -1

      if quad == NW {
        idx = (blockBox.W * oY) + (blockBox.W * y) + x
      } else if quad == NE {
        idx = (blockBox.W * oY) + (blockBox.W * y) + oX + x
      } else if quad == SE {
        idx = (blockBox.W * y) + oX + x
      } else if quad == SW {
        idx = (blockBox.W * y) + x
      } else {
        panic("transformBlocks encountered invalid Quadrant")
      }

      blocks[idx] = color
    }
  }

  return numSolid
}

func (wm *WorldMap) getChip(blockBox geom.RectInt, worldPos mgl32.Vec3, quad int) chip {
  chunkCellX := int(math.Floor(float64(worldPos.X() / (float32(wm.info.ChunkSize) * RESOLUTION))))
  chunkCellY := int(math.Floor(float64(worldPos.Y() / (float32(wm.info.ChunkSize) * RESOLUTION))))
  var c chip
  c.chunkId = uint16((chunkCellY * int(wm.info.Size)) + chunkCellX)
  fileId := uint16(math.Floor(float64(uint32(c.chunkId) / wm.info.ChunksPerFile)))
  c.blocks = wm.chunker.GetChunk(c.chunkId, fileId)

  var texX int
  var texY int
  var texW int
  var texH int
  chunkSize := int(wm.info.ChunkSize)

  if quad == NW || quad == SW {
    texX = blockBox.X - (chunkCellX * chunkSize)
    texW = blockBox.W
    if texX + blockBox.W > chunkSize {
      texW = chunkSize - texX
    }
  } else {
    texX = (blockBox.X + blockBox.W) - (chunkCellX * chunkSize) - blockBox.W
    texW = blockBox.W
    if texX < 0 {
      texW += texX
      texX = 0
    }
  }

  if quad == SW || quad == SE {
    texY = blockBox.Y - (chunkCellY * chunkSize)
    texH = blockBox.H
    if texY + blockBox.H > chunkSize {
      texH = chunkSize - texY
    }
  } else {
    texY = (blockBox.Y + blockBox.H) - (chunkCellY * chunkSize) - blockBox.H
    texH = blockBox.H
    if texY < 0 {
      texH += texY
      texY = 0
    }
  }
  c.box = geom.NewRectInt(texX, texY, texW, texH)

  return c
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
