package phys

import(
	"fmt"
)

type HistoricalTransform struct {
	Timestamp int64
	Angle float32
	XPos float32
	YPos float32
	XVel float32
	YVel float32
}

func (ht *HistoricalTransform) String() string {
	return fmt.Sprintf("[%d] %f deg. xPos: %f yPos: %f xVel: %f yVel: %f", ht.Timestamp, ht.Angle, ht.XPos, ht.YPos, ht.XVel, ht.YVel)
}
