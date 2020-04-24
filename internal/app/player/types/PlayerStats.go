package player

type PlayerStats struct {
  Thrust        float32
  MaxSpeed      float32
  Rotation      float32
}

func NewPlayerStats() *PlayerStats{
  var ps PlayerStats

  ps.Thrust = 12
  ps.MaxSpeed = 20
  ps.Rotation = 172

  return &ps
}
