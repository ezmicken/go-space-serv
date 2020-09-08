package world

import(
  "fmt"
  "math"
  "encoding/binary"
)

type WorldInfo struct {
  ChunksPerFile   uint32
  ChunkSize       uint32
  Size            uint32
  BlocksPerChunk  uint32
  BlocksPerFile   uint32
  NumFiles        uint32
  Seed            uint64
  Threshold       float64
}

func SerializeWorldInfo(info WorldInfo) []byte {
  result := make([]byte, 40)

  binary.LittleEndian.PutUint32(result[:4], info.ChunksPerFile)
  binary.LittleEndian.PutUint32(result[4:8], info.ChunkSize)
  binary.LittleEndian.PutUint32(result[8:12], info.Size)
  binary.LittleEndian.PutUint32(result[12:16], info.BlocksPerChunk)
  binary.LittleEndian.PutUint32(result[16:20], info.BlocksPerFile)
  binary.LittleEndian.PutUint32(result[20:24], info.NumFiles)
  binary.LittleEndian.PutUint64(result[24:32], info.Seed)
  binary.LittleEndian.PutUint64(result[32:40], math.Float64bits(info.Threshold))

  return result
}

func DeserializeWorldInfo(data []byte) WorldInfo {
  var result WorldInfo

  result.ChunksPerFile = binary.LittleEndian.Uint32(data[:4])
  result.ChunkSize = binary.LittleEndian.Uint32(data[4:8])
  result.Size = binary.LittleEndian.Uint32(data[8:12])
  result.BlocksPerChunk = binary.LittleEndian.Uint32(data[12:16])
  result.BlocksPerFile = binary.LittleEndian.Uint32(data[16:20])
  result.NumFiles = binary.LittleEndian.Uint32(data[20:24])
  result.Seed = binary.LittleEndian.Uint64(data[24:32])
  result.Threshold = math.Float64frombits(binary.LittleEndian.Uint64(data[32:40]))

  return result
}

func (info WorldInfo) String() string {
  var result string

  result = fmt.Sprintf("%s%d\t\tChunks Per File\n", result, info.ChunksPerFile)
  result = fmt.Sprintf("%s%d\t\tChunk Size\n", result, info.ChunkSize)
  result = fmt.Sprintf("%s%d\t\tSize\n", result, info.Size)
  result = fmt.Sprintf("%s%d\t\tBlocks Per Chunk\n", result, info.BlocksPerChunk)
  result = fmt.Sprintf("%s%d\t\tBlocks Per File\n", result, info.BlocksPerFile)
  result = fmt.Sprintf("%s%d\t\tFiles\n", result, info.NumFiles)
  result = fmt.Sprintf("%s%d\tSeed\n", result, info.Seed)
  result = fmt.Sprintf("%s%f\tThreshold\n", result, info.Threshold)

  return result
}
