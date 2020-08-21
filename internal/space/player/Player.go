package player

import(
  "github.com/google/uuid"
)

type Player struct {
  Stats   PlayerStats
  BodyId  uint16
  Id      uuid.UUID
}
