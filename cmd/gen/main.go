package main

import(
  "os"
  "path/filepath"
  "compress/zlib"
  "fmt"
  "flag"

  "github.com/ojrac/opensimplex-go"
  "go-space-serv/internal/space/world"
)

// 512 x 512 means chunk id can be uint16
// if this is too small upgrade to uint32 :(

func main() {
  flagCPF := flag.Uint("cpf", 512, "Chunks Per File")
  flagCSize := flag.Uint("csize", 128, "Chunk Size")
  flagSize := flag.Uint("size", 256, "Map Size")
  flagSeed := flag.Uint64("seed", 209323094, "Seed for noise generation")
  flagThreshold := flag.Float64("threshold", 0.36, "Threshold for empty blocks")
  flagClean := flag.Bool("clean", false, "Clean but do not generate map.")

  flag.Parse()

  dir := flag.Arg(0)

  if dir == "" {
    flag.PrintDefaults()
    return
  }

  errMkdir := os.MkdirAll(dir, 0777)
  if errMkdir != nil {
    fmt.Println(errMkdir)
    return
  }

  var info world.WorldInfo
  info.ChunksPerFile = uint32(*flagCPF)
  info.ChunkSize = uint32(*flagCSize)
  info.Size = uint32(*flagSize)
  info.BlocksPerChunk = info.ChunkSize * info.ChunkSize
  info.BlocksPerFile = info.BlocksPerChunk * info.ChunksPerFile
  info.NumFiles = uint32((int64(info.Size) * int64(info.Size)) / int64(info.ChunksPerFile))
  info.Seed = *flagSeed
  info.Threshold = *flagThreshold

  cleanOnly := *flagClean

  fmt.Printf("%v", info)

  fmt.Printf("\nCleaning out %s...", dir)
  cleanFiles(dir)

  if (cleanOnly) {
    return
  }

  fmt.Printf("\nGenerating noise...")
  noise := opensimplex.New(int64(info.Seed))

  fmt.Printf("\nGenerating map...\r\n")
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
    writeChunksToFile(fileBytesCopy, fmt.Sprintf("%s/%03d.chunks", dir, fileId))
    fmt.Printf("\r%d/%d", fileId, info.NumFiles)
  }

  metaFile, err := os.Create(fmt.Sprintf("%s/meta.chunks", dir))
  if err != nil {
    fmt.Println(err)
    return
  }
  defer metaFile.Close()

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
  f, err := os.Create(fileName)
  if err != nil {
    fmt.Println(err)
    return
  }
  defer f.Close()

  zw := zlib.NewWriter(f)
  _, errG := zw.Write(fileBytes)
  if errG != nil {
    fmt.Println(err)
    return
  }

  if err = zw.Close(); err != nil {
    fmt.Println(err)
    return
  }

  if err = f.Sync(); err != nil {
    fmt.Println(err)
    return
  }
}
