package sim

import(
  "github.com/go-gl/mathgl/mgl32"
  "go-space-serv/internal/space/snet"
  "go-space-serv/internal/space/snet/udp"
  "go-space-serv/internal/space/util"
)

type ControlledBody struct {
  controllingPlayer *udp.UDPPlayer
  owningPlayer      *udp.UDPPlayer
  bod               *udp.UDPBody
  stateBuffer       *StateBuffer
}

// instantiation
///////////////////////

func NewControlledBody(plr *udp.UDPPlayer) (*ControlledBody) {
  var cbod ControlledBody
  cbod.controllingPlayer = plr
  cbod.owningPlayer = plr
  cbod.bod = udp.NewUDPBody(snet.GetNextId())
  cbod.stateBuffer = NewStateBuffer(256)

  return &cbod
}

func (cb *ControlledBody) Initialize(ht HistoricalTransform) {
  cb.bod.Position = ht.Position
  cb.bod.Velocity = ht.Velocity
  cb.stateBuffer.Initialize(ht)
}

func (cb *ControlledBody) InputToState(seq int, moveshoot byte) {
  stats := cb.controllingPlayer.GetStats()

  ht := cb.stateBuffer.Get(seq - 1)

  acceleration := float32(0)
  timestep := helpers.GetConfiguredTimestep()

  ht.Position = ht.Position.Add(ht.Velocity)

  left := helpers.BitOn(moveshoot, 0)
  right := helpers.BitOn(moveshoot, 1)
  forward := helpers.BitOn(moveshoot, 2)
  backward := helpers.BitOn(moveshoot, 3)

  if left && !right {
    ht.AngleDelta = helpers.PerSecondOverTime(stats.Rotation, timestep)
  } else if !left && right {
    ht.AngleDelta = helpers.PerSecondOverTime(stats.Rotation, timestep) * -1
  } else {
    ht.AngleDelta = float32(0)
  }

  if forward && !backward {
    acceleration = helpers.PerSecondOverTime(stats.Thrust, timestep)
  } else if !forward && backward {
    acceleration = helpers.PerSecondOverTime(stats.Thrust, timestep) * -1
  } else {
    acceleration = float32(0)
  }

  ht.Angle = helpers.WrapAngle(ht.Angle + ht.AngleDelta)

  if acceleration != 0 {
    q := mgl32.AnglesToQuat(mgl32.DegToRad(ht.Angle), 0, 0, mgl32.ZYX).Normalize()

    // Apply force along the Y axis
    accVec := mgl32.Vec3{0, acceleration, 0}
    originalVel := ht.Velocity
    ht.Velocity = ht.Velocity.Add(q.Rotate(accVec))

    if ht.Velocity.LenSqr() > (stats.MaxSpeed * stats.MaxSpeed) {
      ht.Velocity = ht.Velocity.Normalize().Mul(stats.MaxSpeed)
    }

    ht.VelocityDelta = ht.Velocity.Sub(originalVel)
  } else {
    ht.VelocityDelta = mgl32.Vec3{0, 0, 0}
  }

  ht.Seq++
  cb.stateBuffer.Insert(ht)
  cb.stateBuffer.Clean()
}

func (cb *ControlledBody) ProcessFrame(frameStart int64, seq int) {
  cb.bod.Position = cb.bod.TargetPosition
  cb.bod.Angle = cb.bod.TargetAngle

  if cb.stateBuffer.GetCurrentSeq() <= (seq - 1) {
    ht := cb.stateBuffer.Advance()

    for ht.Seq < (seq - 1) {
      ht = cb.stateBuffer.Advance()
    }

    cb.bod.TargetPosition = ht.Position
    cb.bod.TargetAngle = ht.Angle
    cb.bod.Velocity = ht.Velocity
  }
}

func (cb *ControlledBody) GetBody() *udp.UDPBody {
  return cb.bod
}

func (cb *ControlledBody) SetControllingPlayer(player *udp.UDPPlayer) {
  cb.controllingPlayer = player
}

func (cb *ControlledBody) GetControllingPlayer() *udp.UDPPlayer {
  return cb.controllingPlayer
}

func (cb *ControlledBody) SetOwningPlayer(player *udp.UDPPlayer) {
  cb.owningPlayer = player
}

func (cb *ControlledBody) GetOwningPlayer() *udp.UDPPlayer {
  return cb.owningPlayer
}
