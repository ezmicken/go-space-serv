package world

import(
  "log"
  "fmt"
  "os"
  "bytes"
  "io/ioutil"
  "compress/zlib"

  "go-space-serv/internal/space/util"
)

type Chunker struct {
  info WorldInfo
  files map[uint16][]byte
  access []int64
  writer *zlib.Writer
}

func NewChunker(info WorldInfo) *Chunker {
  var c Chunker
  c.access = make([]int64, info.NumFiles)
  c.files = make(map[uint16][]byte)
  c.info = info
  c.writer = zlib.NewWriter(nil)

  return &c
}

func (c *Chunker) GetChunk(chunkId, fileId uint16) []byte {
  file := c.files[fileId]
  if file == nil {
    file = c.loadFile(fileId)
  }

  c.access[fileId] = helpers.NowMillis()

  chunkStart := uint32(chunkId) - (uint32(fileId) * c.info.ChunksPerFile)
  chunkEnd := chunkStart + c.info.BlocksPerChunk
  chunkSlice := file[chunkStart:chunkEnd]

  var buf bytes.Buffer
  c.writer.Reset(&buf)
  c.writer.Write(chunkSlice)
  c.writer.Close()

  return buf.Bytes()
}

func (c *Chunker) loadFile(fileId uint16) []byte {
  log.Printf("Loading file %s/%03d", c.info.Name, fileId)
  fileName := fmt.Sprintf("assets/%s/%03d.chunks", c.info.Name, fileId)
  file, err := os.Open(fileName)
  if err != nil {
    panic(err)
  }
  defer file.Close()

  zr, errz := zlib.NewReader(file)
  if errz != nil {
    panic(err)
  }

  c.files[fileId], err = ioutil.ReadAll(zr)
  if err != nil {
    panic(err)
  }

  return c.files[fileId]
}
