package main

import(
  "os"
  "path/filepath"
  "compress/gzip"
  "fmt"
  "flag"
  "sync"

  "github.com/ojrac/opensimplex-go"
)

var wg sync.WaitGroup

func main() {
  flagCPF := flag.Int("cpf", 256, "Chunks Per File")
  flagCSize := flag.Int("csize", 128, "Chunk Size")
  flagSize := flag.Int("size", 256,"Map Size")
  flagSeed := flag.Int64("seed", 209323094, "Seed for noise generation")
  flagThreshold := flag.Float64("threshold", 0.36, "Threshold for empty blocks")

  flag.Parse()

  dir := flag.Arg(0)

  if dir == "" {
    flag.PrintDefaults()
    return
  }

  chunksPerFile := *flagCPF
  chunkSize := *flagCSize
  size := *flagSize
  blocksPerChunk := chunkSize * chunkSize
  blocksPerFile := blocksPerChunk * size
  numFiles := (size * size) / chunksPerFile
  seed := *flagSeed
  fmt.Printf("%d\t\tChunks Per File\n", chunksPerFile)
  fmt.Printf("%d\t\tChunk Size\n", chunkSize)
  fmt.Printf("%d\t\tSize\n", size)
  fmt.Printf("%d\t\tBlocks Per Chunk\n", blocksPerChunk)
  fmt.Printf("%d\t\tBlocks PerFile\n", blocksPerFile)
  fmt.Printf("%d\t\tFiles", numFiles)

  fmt.Printf("\nCleaning out %s...", dir)
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

  fmt.Printf("\nGenerating noise...")
  noise := opensimplex.New(seed)

  fmt.Printf("\nGenerating map")
  var fileId int

  // chunk coordinate in world space
  var chunkX int
  var chunkY int
  var chunkId int
  var fileChunkId int

  // block coordinate in chunk space
  var x int
  var y int

  // block coordinate in world space
  var xCoord int
  var yCoord int

  var fileIdx int

  chunkId = 0

  fileBytes := make([]byte, blocksPerFile)

  for fileId = 0; fileId < numFiles; fileId++ {
    fileIdx = 0
    for fileChunkId = 0; fileChunkId < chunksPerFile; fileChunkId++ {
      chunkX = chunkId % size
      chunkY = chunkId / size

      for y = 0; y < chunkSize; y++ {
        for x = 0; x < chunkSize; x++ {
          xCoord = (chunkX * chunkSize) + x
          yCoord = (chunkY * chunkSize) + y
          noiseVal := noise.Eval2(float64(xCoord) * 0.05, float64(yCoord) * 0.05)
          if noiseVal > *flagThreshold {
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
  fmt.Sprintf("\nDone.")
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
