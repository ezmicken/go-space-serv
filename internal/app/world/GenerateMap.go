package world

import (
	//"log"
	//"bytes"

	"github.com/ojrac/opensimplex-go"

	."go-space-serv/internal/app/world/types"
)

func GenerateMap(seed int64, w int, h int) (blocks [][]BlockType) {
	noise := opensimplex.New(seed)

	// initialize multidimensional array
	blocks = make([][]BlockType, h)
	for i := range blocks {
	  blocks[i] = make([]BlockType, w)
	}

	//var b bytes.Buffer

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			floatVal := noise.Eval2(float64(x) * 0.05, float64(y) * 0.05)
			if floatVal > 0.36 {
				blocks[y][x] = SOLID
				//b.WriteString("1")
			} else {
				blocks[y][x] = EMPTY
				//b.WriteString("0")
			}
		}
		//b.WriteString("\n")
	}

	//log.Printf(b.String())
	return;
}
