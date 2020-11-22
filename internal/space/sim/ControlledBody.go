package sim

import(
  //"log"
  "github.com/go-gl/mathgl/mgl32"
  "github.com/chewxy/math32"
  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/util"
  //"go-space-serv/internal/space/geom"
  "go-space-serv/internal/space/world"
  //"go-space-serv/internal/space/sim/msg"
)

type ControlledBody struct {
  controllingPlayer *SimPlayer
  owningPlayer      *SimPlayer
  bod               *udp.UDPBody
  stateBuffer       *StateBuffer
  Collider          *SpaceCollider

  inputSeq          int
}

var accLUT = [361]mgl32.Vec3{
  {0.0000, 0.3600, 0},
  {-0.0063, 0.3599, 0},
  {-0.0126, 0.3598, 0},
  {-0.0188, 0.3595, 0},
  {-0.0251, 0.3591, 0},
  {-0.0314, 0.3586, 0},
  {-0.0376, 0.3580, 0},
  {-0.0439, 0.3573, 0},
  {-0.0501, 0.3565, 0},
  {-0.0563, 0.3556, 0},
  {-0.0625, 0.3545, 0},
  {-0.0687, 0.3534, 0},
  {-0.0748, 0.3521, 0},
  {-0.0810, 0.3508, 0},
  {-0.0871, 0.3493, 0},
  {-0.0932, 0.3477, 0},
  {-0.0992, 0.3461, 0},
  {-0.1053, 0.3443, 0},
  {-0.1112, 0.3424, 0},
  {-0.1172, 0.3404, 0},
  {-0.1231, 0.3383, 0},
  {-0.1290, 0.3361, 0},
  {-0.1349, 0.3338, 0},
  {-0.1407, 0.3314, 0},
  {-0.1464, 0.3289, 0},
  {-0.1521, 0.3263, 0},
  {-0.1578, 0.3236, 0},
  {-0.1634, 0.3208, 0},
  {-0.1690, 0.3179, 0},
  {-0.1745, 0.3149, 0},
  {-0.1800, 0.3118, 0},
  {-0.1854, 0.3086, 0},
  {-0.1908, 0.3053, 0},
  {-0.1961, 0.3019, 0},
  {-0.2013, 0.2985, 0},
  {-0.2065, 0.2949, 0},
  {-0.2116, 0.2912, 0},
  {-0.2167, 0.2875, 0},
  {-0.2216, 0.2837, 0},
  {-0.2266, 0.2798, 0},
  {-0.2314, 0.2758, 0},
  {-0.2362, 0.2717, 0},
  {-0.2409, 0.2675, 0},
  {-0.2455, 0.2633, 0},
  {-0.2501, 0.2590, 0},
  {-0.2546, 0.2546, 0},
  {-0.2590, 0.2501, 0},
  {-0.2633, 0.2455, 0},
  {-0.2675, 0.2409, 0},
  {-0.2717, 0.2362, 0},
  {-0.2758, 0.2314, 0},
  {-0.2798, 0.2266, 0},
  {-0.2837, 0.2216, 0},
  {-0.2875, 0.2167, 0},
  {-0.2912, 0.2116, 0},
  {-0.2949, 0.2065, 0},
  {-0.2985, 0.2013, 0},
  {-0.3019, 0.1961, 0},
  {-0.3053, 0.1908, 0},
  {-0.3086, 0.1854, 0},
  {-0.3118, 0.1800, 0},
  {-0.3149, 0.1745, 0},
  {-0.3179, 0.1690, 0},
  {-0.3208, 0.1634, 0},
  {-0.3236, 0.1578, 0},
  {-0.3263, 0.1521, 0},
  {-0.3289, 0.1464, 0},
  {-0.3314, 0.1407, 0},
  {-0.3338, 0.1349, 0},
  {-0.3361, 0.1290, 0},
  {-0.3383, 0.1231, 0},
  {-0.3404, 0.1172, 0},
  {-0.3424, 0.1112, 0},
  {-0.3443, 0.1053, 0},
  {-0.3461, 0.0992, 0},
  {-0.3477, 0.0932, 0},
  {-0.3493, 0.0871, 0},
  {-0.3508, 0.0810, 0},
  {-0.3521, 0.0748, 0},
  {-0.3534, 0.0687, 0},
  {-0.3545, 0.0625, 0},
  {-0.3556, 0.0563, 0},
  {-0.3565, 0.0501, 0},
  {-0.3573, 0.0439, 0},
  {-0.3580, 0.0376, 0},
  {-0.3586, 0.0314, 0},
  {-0.3591, 0.0251, 0},
  {-0.3595, 0.0188, 0},
  {-0.3598, 0.0126, 0},
  {-0.3599, 0.0063, 0},
  {-0.3600, 0.0000, 0},
  {-0.3599, -0.0063, 0},
  {-0.3598, -0.0126, 0},
  {-0.3595, -0.0188, 0},
  {-0.3591, -0.0251, 0},
  {-0.3586, -0.0314, 0},
  {-0.3580, -0.0376, 0},
  {-0.3573, -0.0439, 0},
  {-0.3565, -0.0501, 0},
  {-0.3556, -0.0563, 0},
  {-0.3545, -0.0625, 0},
  {-0.3534, -0.0687, 0},
  {-0.3521, -0.0748, 0},
  {-0.3508, -0.0810, 0},
  {-0.3493, -0.0871, 0},
  {-0.3477, -0.0932, 0},
  {-0.3461, -0.0992, 0},
  {-0.3443, -0.1053, 0},
  {-0.3424, -0.1112, 0},
  {-0.3404, -0.1172, 0},
  {-0.3383, -0.1231, 0},
  {-0.3361, -0.1290, 0},
  {-0.3338, -0.1349, 0},
  {-0.3314, -0.1407, 0},
  {-0.3289, -0.1464, 0},
  {-0.3263, -0.1521, 0},
  {-0.3236, -0.1578, 0},
  {-0.3208, -0.1634, 0},
  {-0.3179, -0.1690, 0},
  {-0.3149, -0.1745, 0},
  {-0.3118, -0.1800, 0},
  {-0.3086, -0.1854, 0},
  {-0.3053, -0.1908, 0},
  {-0.3019, -0.1961, 0},
  {-0.2985, -0.2013, 0},
  {-0.2949, -0.2065, 0},
  {-0.2912, -0.2116, 0},
  {-0.2875, -0.2167, 0},
  {-0.2837, -0.2216, 0},
  {-0.2798, -0.2266, 0},
  {-0.2758, -0.2314, 0},
  {-0.2717, -0.2362, 0},
  {-0.2675, -0.2409, 0},
  {-0.2633, -0.2455, 0},
  {-0.2590, -0.2501, 0},
  {-0.2546, -0.2546, 0},
  {-0.2501, -0.2590, 0},
  {-0.2455, -0.2633, 0},
  {-0.2409, -0.2675, 0},
  {-0.2362, -0.2717, 0},
  {-0.2314, -0.2758, 0},
  {-0.2266, -0.2798, 0},
  {-0.2216, -0.2837, 0},
  {-0.2167, -0.2875, 0},
  {-0.2116, -0.2912, 0},
  {-0.2065, -0.2949, 0},
  {-0.2013, -0.2985, 0},
  {-0.1961, -0.3019, 0},
  {-0.1908, -0.3053, 0},
  {-0.1854, -0.3086, 0},
  {-0.1800, -0.3118, 0},
  {-0.1745, -0.3149, 0},
  {-0.1690, -0.3179, 0},
  {-0.1634, -0.3208, 0},
  {-0.1578, -0.3236, 0},
  {-0.1521, -0.3263, 0},
  {-0.1464, -0.3289, 0},
  {-0.1407, -0.3314, 0},
  {-0.1349, -0.3338, 0},
  {-0.1290, -0.3361, 0},
  {-0.1231, -0.3383, 0},
  {-0.1172, -0.3404, 0},
  {-0.1112, -0.3424, 0},
  {-0.1053, -0.3443, 0},
  {-0.0992, -0.3461, 0},
  {-0.0932, -0.3477, 0},
  {-0.0871, -0.3493, 0},
  {-0.0810, -0.3508, 0},
  {-0.0748, -0.3521, 0},
  {-0.0687, -0.3534, 0},
  {-0.0625, -0.3545, 0},
  {-0.0563, -0.3556, 0},
  {-0.0501, -0.3565, 0},
  {-0.0439, -0.3573, 0},
  {-0.0376, -0.3580, 0},
  {-0.0314, -0.3586, 0},
  {-0.0251, -0.3591, 0},
  {-0.0188, -0.3595, 0},
  {-0.0126, -0.3598, 0},
  {-0.0063, -0.3599, 0},
  {0.0000, -0.3600, 0},
  {0.0063, -0.3599, 0},
  {0.0126, -0.3598, 0},
  {0.0188, -0.3595, 0},
  {0.0251, -0.3591, 0},
  {0.0314, -0.3586, 0},
  {0.0376, -0.3580, 0},
  {0.0439, -0.3573, 0},
  {0.0501, -0.3565, 0},
  {0.0563, -0.3556, 0},
  {0.0625, -0.3545, 0},
  {0.0687, -0.3534, 0},
  {0.0748, -0.3521, 0},
  {0.0810, -0.3508, 0},
  {0.0871, -0.3493, 0},
  {0.0932, -0.3477, 0},
  {0.0992, -0.3461, 0},
  {0.1053, -0.3443, 0},
  {0.1112, -0.3424, 0},
  {0.1172, -0.3404, 0},
  {0.1231, -0.3383, 0},
  {0.1290, -0.3361, 0},
  {0.1349, -0.3338, 0},
  {0.1407, -0.3314, 0},
  {0.1464, -0.3289, 0},
  {0.1521, -0.3263, 0},
  {0.1578, -0.3236, 0},
  {0.1634, -0.3208, 0},
  {0.1690, -0.3179, 0},
  {0.1745, -0.3149, 0},
  {0.1800, -0.3118, 0},
  {0.1854, -0.3086, 0},
  {0.1908, -0.3053, 0},
  {0.1961, -0.3019, 0},
  {0.2013, -0.2985, 0},
  {0.2065, -0.2949, 0},
  {0.2116, -0.2912, 0},
  {0.2167, -0.2875, 0},
  {0.2216, -0.2837, 0},
  {0.2266, -0.2798, 0},
  {0.2314, -0.2758, 0},
  {0.2362, -0.2717, 0},
  {0.2409, -0.2675, 0},
  {0.2455, -0.2633, 0},
  {0.2501, -0.2590, 0},
  {0.2546, -0.2546, 0},
  {0.2590, -0.2501, 0},
  {0.2633, -0.2455, 0},
  {0.2675, -0.2409, 0},
  {0.2717, -0.2362, 0},
  {0.2758, -0.2314, 0},
  {0.2798, -0.2266, 0},
  {0.2837, -0.2216, 0},
  {0.2875, -0.2167, 0},
  {0.2912, -0.2116, 0},
  {0.2949, -0.2065, 0},
  {0.2985, -0.2013, 0},
  {0.3019, -0.1961, 0},
  {0.3053, -0.1908, 0},
  {0.3086, -0.1854, 0},
  {0.3118, -0.1800, 0},
  {0.3149, -0.1745, 0},
  {0.3179, -0.1690, 0},
  {0.3208, -0.1634, 0},
  {0.3236, -0.1578, 0},
  {0.3263, -0.1521, 0},
  {0.3289, -0.1464, 0},
  {0.3314, -0.1407, 0},
  {0.3338, -0.1349, 0},
  {0.3361, -0.1290, 0},
  {0.3383, -0.1231, 0},
  {0.3404, -0.1172, 0},
  {0.3424, -0.1112, 0},
  {0.3443, -0.1053, 0},
  {0.3461, -0.0992, 0},
  {0.3477, -0.0932, 0},
  {0.3493, -0.0871, 0},
  {0.3508, -0.0810, 0},
  {0.3521, -0.0748, 0},
  {0.3534, -0.0687, 0},
  {0.3545, -0.0625, 0},
  {0.3556, -0.0563, 0},
  {0.3565, -0.0501, 0},
  {0.3573, -0.0439, 0},
  {0.3580, -0.0376, 0},
  {0.3586, -0.0314, 0},
  {0.3591, -0.0251, 0},
  {0.3595, -0.0188, 0},
  {0.3598, -0.0126, 0},
  {0.3599, -0.0063, 0},
  {0.3600, 0.0000, 0},
  {0.3599, 0.0063, 0},
  {0.3598, 0.0126, 0},
  {0.3595, 0.0188, 0},
  {0.3591, 0.0251, 0},
  {0.3586, 0.0314, 0},
  {0.3580, 0.0376, 0},
  {0.3573, 0.0439, 0},
  {0.3565, 0.0501, 0},
  {0.3556, 0.0563, 0},
  {0.3545, 0.0625, 0},
  {0.3534, 0.0687, 0},
  {0.3521, 0.0748, 0},
  {0.3508, 0.0810, 0},
  {0.3493, 0.0871, 0},
  {0.3477, 0.0932, 0},
  {0.3461, 0.0992, 0},
  {0.3443, 0.1053, 0},
  {0.3424, 0.1112, 0},
  {0.3404, 0.1172, 0},
  {0.3383, 0.1231, 0},
  {0.3361, 0.1290, 0},
  {0.3338, 0.1349, 0},
  {0.3314, 0.1407, 0},
  {0.3289, 0.1464, 0},
  {0.3263, 0.1521, 0},
  {0.3236, 0.1578, 0},
  {0.3208, 0.1634, 0},
  {0.3179, 0.1690, 0},
  {0.3149, 0.1745, 0},
  {0.3118, 0.1800, 0},
  {0.3086, 0.1854, 0},
  {0.3053, 0.1908, 0},
  {0.3019, 0.1961, 0},
  {0.2985, 0.2013, 0},
  {0.2949, 0.2065, 0},
  {0.2912, 0.2116, 0},
  {0.2875, 0.2167, 0},
  {0.2837, 0.2216, 0},
  {0.2798, 0.2266, 0},
  {0.2758, 0.2314, 0},
  {0.2717, 0.2362, 0},
  {0.2675, 0.2409, 0},
  {0.2633, 0.2455, 0},
  {0.2590, 0.2501, 0},
  {0.2546, 0.2546, 0},
  {0.2501, 0.2590, 0},
  {0.2455, 0.2633, 0},
  {0.2409, 0.2675, 0},
  {0.2362, 0.2717, 0},
  {0.2314, 0.2758, 0},
  {0.2266, 0.2798, 0},
  {0.2216, 0.2837, 0},
  {0.2167, 0.2875, 0},
  {0.2116, 0.2912, 0},
  {0.2065, 0.2949, 0},
  {0.2013, 0.2985, 0},
  {0.1961, 0.3019, 0},
  {0.1908, 0.3053, 0},
  {0.1854, 0.3086, 0},
  {0.1800, 0.3118, 0},
  {0.1745, 0.3149, 0},
  {0.1690, 0.3179, 0},
  {0.1634, 0.3208, 0},
  {0.1578, 0.3236, 0},
  {0.1521, 0.3263, 0},
  {0.1464, 0.3289, 0},
  {0.1407, 0.3314, 0},
  {0.1349, 0.3338, 0},
  {0.1290, 0.3361, 0},
  {0.1231, 0.3383, 0},
  {0.1172, 0.3404, 0},
  {0.1112, 0.3424, 0},
  {0.1053, 0.3443, 0},
  {0.0992, 0.3461, 0},
  {0.0932, 0.3477, 0},
  {0.0871, 0.3493, 0},
  {0.0810, 0.3508, 0},
  {0.0748, 0.3521, 0},
  {0.0687, 0.3534, 0},
  {0.0625, 0.3545, 0},
  {0.0563, 0.3556, 0},
  {0.0501, 0.3565, 0},
  {0.0439, 0.3573, 0},
  {0.0376, 0.3580, 0},
  {0.0314, 0.3586, 0},
  {0.0251, 0.3591, 0},
  {0.0188, 0.3595, 0},
  {0.0126, 0.3598, 0},
  {0.0063, 0.3599, 0},
  {0.0000, 0.3600, 0},
}
// instantiation
///////////////////////

func NewControlledBody(plr *SimPlayer) (*ControlledBody) {
  var cbod ControlledBody
  cbod.controllingPlayer = plr
  cbod.owningPlayer = plr
  cbod.bod = udp.NewUDPBody(snet.GetNextId())
  cbod.stateBuffer = NewStateBuffer(256)
  cbod.Collider = NewSpaceCollider(96, 48)

  return &cbod
}

func (cb *ControlledBody) Initialize(ht HistoricalTransform) {
  cb.bod.Position = ht.Position
  cb.bod.Velocity = ht.Velocity
  cb.stateBuffer.Initialize(ht)
  cb.inputSeq = ht.Seq
}

func (cb *ControlledBody) InputToState(seq int, moveshoot byte) {
  if seq < cb.inputSeq {
    return
  } else {
    cb.inputSeq = seq
  }
  stats := cb.controllingPlayer.Stats

  ht := cb.stateBuffer.Get(seq - 1)

  acceleration := float32(0)
//  timestep := helpers.GetConfiguredTimestep()

  ht.Position = ht.Position.Add(ht.Velocity)

  left := helpers.BitOn(moveshoot, 0)
  right := helpers.BitOn(moveshoot, 1)
  forward := helpers.BitOn(moveshoot, 2)
  backward := helpers.BitOn(moveshoot, 3)

  if left && !right {
    //ht.AngleDelta = helpers.PerSecondOverTime(stats.Rotation, timestep)
    ht.AngleDelta = float32(9.0);
  } else if !left && right {
    //ht.AngleDelta = helpers.PerSecondOverTime(stats.Rotation, timestep) * -1
    ht.AngleDelta = float32(-9.0)
  } else {
    ht.AngleDelta = float32(0)
  }

  if forward && !backward {
    //acceleration = helpers.PerSecondOverTime(stats.Thrust, timestep)
    //acceleration = float32(0.36)
    acceleration = float32(1)
  } else if !forward && backward {
    //acceleration = helpers.PerSecondOverTime(stats.Thrust, timestep) * -1
    //acceleration = float32(-0.36)
    acceleration = float32(-1)
  }

  ht.Angle = helpers.WrapAngle(ht.Angle + ht.AngleDelta)

  if acceleration != 0 {
    //q := angleToQuat(ht.Angle)
    //q := mgl32.AnglesToQuat(mgl32.DegToRad(ht.Angle), 0, 0, mgl32.ZYX).Normalize()

    // Apply force along the Y axis
    // accVec := mgl32.Vec3{0, acceleration, 0}
    // accVec = q.Rotate(accVec)
    originalVel := ht.Velocity
    //ht.Velocity = ht.Velocity.Add(accVec)
    ht.Velocity = ht.Velocity.Add(accLUT[int(ht.Angle)].Mul(acceleration))

    if ht.Velocity.LenSqr() > (stats.MaxSpeed * stats.MaxSpeed) {
      ht.Velocity = ht.Velocity.Normalize().Mul(stats.MaxSpeed)
    }
    ht.VelocityDelta = ht.Velocity.Sub(originalVel)
  } else {
    ht.VelocityDelta = mgl32.Vec3{0, 0, 0}
  }

  ht.Seq++
  cb.stateBuffer.Insert(ht, 0)
  cb.stateBuffer.Clean()
}

func (cb *ControlledBody) ProcessFrame(frameStart int64, seq int, wm *world.WorldMap) (x, y float32) {
  cb.bod.Position = cb.bod.TargetPosition
  cb.bod.Angle = cb.bod.TargetAngle

  if cb.stateBuffer.GetCurrentSeq() <= (seq - 1) {
    ht := cb.stateBuffer.Advance()

    for ht.Seq < (seq - 1) {
      ht = cb.stateBuffer.Advance()
    }

    cb.Collider.Update(ht.Position, ht.Velocity)
    potentialCollisions := wm.GetBlockRects(cb.Collider.Broad)
    if potentialCollisions != nil {
      cc := 1
      check2 := ht
      check, _ := cb.Collider.Check(ht, potentialCollisions)
      for check != check2 && cc <= 4 {
        check2 = check
        check, _ = cb.Collider.Check(check, potentialCollisions)
        cc++
      }

      if ht != check {
        stats := cb.controllingPlayer.Stats
        if check.Velocity.LenSqr() > (stats.MaxSpeed * stats.MaxSpeed) {
          check.Velocity = check.Velocity.Normalize().Mul(stats.MaxSpeed)
        }
        cb.stateBuffer.Insert(check, 1)
        cb.stateBuffer.Clean()
        ht = check
      }
    }

    cb.bod.TargetPosition = ht.Position
    cb.bod.TargetAngle = ht.Angle
    cb.bod.Velocity = ht.Velocity
    x = ht.Position.X()
    y = ht.Position.Y()
  } else {
    x = -1
    y = -1
  }

  return
}

func (cb *ControlledBody) GetBody() *udp.UDPBody {
  return cb.bod
}

func (cb *ControlledBody) SetControllingPlayer(player *SimPlayer) {
  cb.controllingPlayer = player
}

func (cb *ControlledBody) GetControllingPlayer() *SimPlayer {
  return cb.controllingPlayer
}

func (cb *ControlledBody) SetOwningPlayer(player *SimPlayer) {
  cb.owningPlayer = player
}

func (cb *ControlledBody) GetOwningPlayer() *SimPlayer {
  return cb.owningPlayer
}
