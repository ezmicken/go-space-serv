package main

import(
  "os"
  "path/filepath"
  "compress/gzip"
  "fmt"
  "flag"
  "sync"

  "github.com/ojrac/opensimplex-go"
  "go-space-serv/internal/space/world"
)

var wg sync.WaitGroup

func main() {
  flagCPF := flag.Uint("cpf", 256, "Chunks Per File")
  flagCSize := flag.Uint("csize", 128, "Chunk Size")
  flagSize := flag.Uint("size", 256,"Map Size")
  flagSeed := flag.Uint64("seed", 209323094, "Seed for noise generation")
  flagThreshold := flag.Float64("threshold", 0.36, "Threshold for empty blocks")

  flag.Parse()

  dir := flag.Arg(0)

  if dir == "" {
    flag.PrintDefaults()
    return
  }

  var info world.WorldInfo
  info.ChunksPerFile = uint32(*flagCPF)
  info.ChunkSize = uint32(*flagCSize)
  info.Size = uint32(*flagSize)
  info.BlocksPerChunk = info.ChunkSize * info.ChunkSize
  info.BlocksPerFile = info.BlocksPerChunk * info.Size
  info.NumFiles = (info.Size * info.Size) / info.ChunksPerFile
  info.Seed = *flagSeed
  info.Threshold = *flagThreshold

  fmt.Printf("%v", info)

  fmt.Printf("\nCleaning out %s...", dir)
  cleanFiles(dir)

  fmt.Printf("\nGenerating noise...")
  noise := opensimplex.New(int64(info.Seed))

  fmt.Printf("\nGenerating map")
  var fileId uint32

  // chunk coordinate in world space
  var chunkX uint32
  var chunkY uint32
  var chunkId uint32
  var fileChunkId uint32

  // block coordinate in chunk space
  var x uint32
  var y uint32

  // block coordinate in world space
  var xCoord uint32
  var yCoord uint32

  var fileIdx uint32

  chunkId = 0

  fileBytes := make([]byte, info.BlocksPerFile)

  for fileId = 0; fileId < info.NumFiles; fileId++ {
    fileIdx = 0
    for fileChunkId = 0; fileChunkId < info.ChunksPerFile; fileChunkId++ {
      chunkX = chunkId % info.Size
      chunkY = chunkId / info.Size

      for y = 0; y < info.ChunkSize; y++ {
        for x = 0; x < info.ChunkSize; x++ {
          xCoord = (chunkX * info.ChunkSize) + x
          yCoord = (chunkY * info.ChunkSize) + y
          noiseVal := noise.Eval2(float64(xCoord) * 0.05, float64(yCoord) * 0.05)
          if noiseVal > info.Threshold {
            fileBytes[fileIdx] = 1
          } else {
            fileBytes[fileIdx] = 0
          }
          fileIdx++
        }
      }
      chunkId++
    }

    fileBytesCopy := fileBytes
    go writeChunksToFile(fileBytesCopy, fmt.Sprintf("%s/%03d.chunks", dir, fileId))
  }

  wg.Wait()

  metaFile, err := os.OpenFile(fmt.Sprintf("%s/meta.chunks", dir), os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Printf("\nSaving info...")
  metaBytes := world.SerializeWorldInfo(info)
  _, wErr := metaFile.Write(metaBytes)
  if wErr != nil {
    fmt.Println(wErr)
    return
  }

  fmt.Printf("\nDone.\n")
}

func cleanFiles(dir string) {
  files, err := filepath.Glob(filepath.Join(dir, "*.chunks"))
  if err != nil {
    fmt.Println(err)
    return
  }

  for _, file := range files {
    err = os.RemoveAll(file)
    if err != nil {
      fmt.Println(err)
      return
    }
  }
}

func writeChunksToFile(fileBytes []byte, fileName string) {
  wg.Add(1)

  f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    fmt.Println(err)
    wg.Done()
    return
  }

  zw := gzip.NewWriter(f)
  _, errG := zw.Write(fileBytes)
  if errG != nil {
    fmt.Println(err)
    wg.Done()
    return
  }
  fmt.Printf(".")

  if err = zw.Close(); err != nil {
    fmt.Println(err)
    wg.Done()
    return
  }

  wg.Done()
}
