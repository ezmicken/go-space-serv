package player

type PlayerStats struct {
  Thrust        float32   // Used to calculate acceleration.
  MaxSpeed      float32   // Maximum speed the player can travel at.
  Rotation      float32   // Speed at which the player rotates.
  BombSpeed     float32   // How much velocity to add to bombs.
  BombRadius    float32   // Radius of bomb explosions.
  BombProximity float32   // Size of bomb explosion colliders.
  BombLife      int32     // How long bombs stay alive. (ms)
  BombBounces   int32     // How many times the bomb can bounce before exploding.
  BombRate      int32     // rate of fire in milliseconds between shots.
}

func DefaultPlayerStats() PlayerStats{
  var ps PlayerStats

  ps.Thrust = 0.36
  ps.MaxSpeed = 20
  ps.Rotation = 9
  ps.BombRate = 750
  ps.BombSpeed = 32
  ps.BombRadius = 64
  ps.BombProximity = 32
  ps.BombLife = 8000
  ps.BombBounces = 3

  return ps
}
