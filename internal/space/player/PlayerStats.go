package player

type PlayerStats struct {
  Thrust        float32
  MaxSpeed      float32
  Rotation      float32
}

func DefaultPlayerStats() PlayerStats{
  var ps PlayerStats

  ps.Thrust = 12
  ps.MaxSpeed = 20
  ps.Rotation = 260

  return ps
}
