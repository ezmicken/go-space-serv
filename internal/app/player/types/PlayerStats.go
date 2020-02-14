package player

type PlayerStats struct {
	Thrust 				float32
	MaxSpeed 			float32
	Acceleration 	float32
	Rotation 			uint16
}

func NewPlayerStats() *PlayerStats{
	var ps PlayerStats

	ps.Thrust = 8
	ps.MaxSpeed = 16
	ps.Acceleration = 0.3
	ps.Rotation = 5

	return &ps
}
