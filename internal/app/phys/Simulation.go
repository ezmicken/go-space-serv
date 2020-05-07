package phys

import(
	"sync"
	"log"
	"time"

	"go-space-serv/internal/app/helpers"
	"go-space-serv/internal/app/world"

	. "go-space-serv/internal/app/phys/types"
	. "go-space-serv/internal/app/snet/types"
)

type Simulation struct {
	controlledBodies 		sync.Map
	allBodies 					[]*UdpBody
	worldMap            *world.WorldMap

  // Timing
  seq                 uint16        // incremented each simulation frame, sync when rolls over
  lastSync            int64         // unix nanos the last time sync was performed
  lastFrame           int64         // unix nanos since the last simulation frame
  framesSinceLastSync int64         // simulation frames since last sync

  output       chan		*NetworkMsg
}

func (s *Simulation) Start(worldMap *world.WorldMap) {
	s.output = make(chan *NetworkMsg, 100)
	s.worldMap = worldMap
	s.seq = 0
	s.lastSync = 0
	s.framesSinceLastSync = 0
	s.worldMap.SpawnX = 500
	s.worldMap.SpawnY = 500
	go s.loop()
}

func (s *Simulation) processFrame(frameStart int64) {
  // initialize the frame
  frame := NewUdpFrame(s.seq)
  frameStartMillis := helpers.NanosToMillis(frameStart)

  // Update all bodies
  // flag dead bodies for removal
  // process live bodies
  // replace ps.bodies with filtered list
  filteredBodies := s.allBodies[:0]
  for i, b := range s.allBodies {
    if !b.IsDead() {
      filteredBodies = append(filteredBodies, b)
    } else {
      s.allBodies[i] = nil
      continue
    }

    player := b.GetControllingPlayer()
    if player != nil && player.IsActive() {
      b.ProcessInput(s.seq, frameStartMillis)
    } else {
      log.Printf("player is nil on controlled body %d", b.GetId());
    }

    b.ApplyTransform(frameStartMillis)
    frame.AddUdpBody(b)
  }
  s.allBodies = filteredBodies

  // propagate input to clients
  if frame.Len() > 0 {
    serializedFrame := frame.Serialize()
    var frameMsg NetworkMsg
    frameMsg.PutByte(byte(SFrame))
    frameMsg.PutUint16(s.seq)
    frameMsg.PutBytes(serializedFrame)

    // Send the frame to relevent players.
    s.push(&frameMsg)
  }
}

// Simulation loop
// Determines when to process frames.
// Initiates synchronization.
// Processes input not related to controlled bodies
func (s *Simulation) loop() {
  simulationStart := time.Now().UnixNano()
  s.lastSync = simulationStart
  s.seq = 0
  shouldSync := false
  timestepNano := helpers.GetConfiguredTimestepNanos()

  for {
    frameStartTime := time.Now()
    frameStart := frameStartTime.UnixNano()
    framesToProcess := ((frameStart - s.lastSync) / timestepNano) - int64(s.seq)
    if framesToProcess > 0 {
      for i := int64(0); i < framesToProcess; i++ {
        s.seq++;
        s.lastFrame = s.lastSync + (int64(s.seq) * timestepNano)
        s.processFrame(s.lastFrame)

        if s.seq == 0 {
          shouldSync = true
          s.seq = 1
        }
      }
    } else if shouldSync {
      log.Printf("SYNC")
      s.sync()
      shouldSync = false;
    }

    time.Sleep(32)
  }
}

// Actions
//////////////

func (s *Simulation) push(msg *NetworkMsg) {
	select {
		case s.output <- msg:
		default:
			log.Printf("Simulation output full. Discarding...")
	}
}

func (s *Simulation) Pull() *NetworkMsg {
	var msg *NetworkMsg
	select {
		case msg = <- s.output:
		default:
			msg = nil
	}

	return msg
}

func (s *Simulation) sync() {
	syncTime := s.lastFrame  / (int64(time.Millisecond)/int64(time.Nanosecond))

  var syncMsg NetworkMsg
  syncMsg.PutByte(byte(SSync))
  syncMsg.PutUint16(s.seq)
  syncMsg.PutUint64(uint64(syncTime))

  s.lastSync = syncTime

  s.push(&syncMsg)
}

func (s *Simulation) SpawnPlayer(in *UdpInput, player *UdpPlayer) {
  playerName := in.GetName();

  if player != nil && player.GetState() != SPECTATING {
    log.Printf("player %s spawn when not spectating", playerName)
    return
  }

  // OK, add body to simulation
  spawnX := s.worldMap.SpawnX
  spawnY := s.worldMap.SpawnY
  x, y := s.worldMap.GetCellCenter(spawnX, spawnY)
  pBod := NewControlledUdpBody(player)
  pBod.SetPos(x, y)
  s.addControlledBody(playerName, pBod)

  var msg NetworkMsg
  msg.PutByte(byte(SSpawn))
  msg.PutUint16(pBod.GetId())
  msg.PutUint32(spawnX)
  msg.PutUint32(spawnY)
  s.push(&msg)

  log.Printf("Spawning %s at %d/%d -- %f/%f", player.GetName(), spawnX, spawnY, x, y)
}

// Modify
///////////

func (s *Simulation) addControlledBody(id string, bod *UdpBody) {
	s.allBodies = append(s.allBodies, bod)
	s.controlledBodies.Store(id, bod)
}

func (s *Simulation) RemoveControlledBody(id string) {
	bod, ok := s.controlledBodies.Load(id)
	if ok && bod != nil {
		bod.(*UdpBody).Kill()
		s.controlledBodies.Delete(id)
	}
}

func (s *Simulation) AddMove(input *UdpInput) {
	playerName := input.GetName()
	bod, ok := s.controlledBodies.Load(playerName)
	if ok && bod != nil {
		bod.(*UdpBody).QueueInput(input)
	}
}

// Access
/////////////

func (s *Simulation) GetLastSync() int64 {
	return s.lastSync
}

func (s *Simulation) GetSeq() uint16 {
	return s.seq
}
