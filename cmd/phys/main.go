package main

import (
  "log"
  "sync"
  "time"
  "net"
  "bufio"
  //"runtime"
  //"runtime/debug"

  "github.com/panjf2000/gnet"
  "github.com/panjf2000/gnet/pool/goroutine"
  "github.com/bgadrian/data-structures/priorityqueue"

  //"github.com/bxcodec/saint" // integer math

  //"go-space-serv/internal/app/phys"
  "go-space-serv/internal/app/snet"
  . "go-space-serv/internal/app/player/types"
  . "go-space-serv/internal/app/phys/types"
  . "go-space-serv/internal/app/snet/types"
  "go-space-serv/internal/app/world"
)

// physicsServer's job is to run the physics simulation
// and propagate input between clients.
// The simulation runs at 30fps and the loop runs
// at or near 120fps.
// server chooses a time to start simulation
// and updates it every 256 frames.
// server sends this time to clients so that they
// can simulate from the same time as well. This
// keeps the numbers from growing too large.

type physicsServer struct {
  *gnet.EventServer
  pool            	*goroutine.Pool

  // Simulation data
  playerCount			  	int    			// count of connectedPlayers
  connectedPlayers 		sync.Map		// player id <-> UdpPlayer
  controlMap        	sync.Map    // player id <-> UdpBody
  inputs            	*priorityqueue.HierarchicalQueue // Inputs that are not part of the simulation
  bodies 							[]*UdpBody

	// Timing
  seq									byte					// incremented each simulation frame, sync when rolls over
  launchTime        	int64					// unix nanos when the program started
  lastSync						int64 				// unix nanos the last time sync was performed
  lastFrame           int64					// unix nanos since the last simulation frame
  framesSinceLastSync int64					// simulation frames since last sync
  tick								time.Duration // loop speed
  TIMESTEP            int64					// simulation speed

  // World map data
  worldConn         	net.Conn
  worldMap						world.WorldMap
  worldMapBytes     	[]byte
  worldMapLen					int
}

// protocol constants
// message structure is as follows:
// [ length, command, content ]
//     4b       1b    <= 4091b
const prefixLen int = 4
const cmdLen int = 1
const maxMsgSize int = 1500
const udpAddr = "udp://:9495";
const udpPort = 9495

// map constants
const spawnX int = 500
const spawnY int = 500

// Helpers
/////////////
func makeTimestamp() uint64 {
	return uint64(time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond)))
}

// Actions
/////////////

func (ps *physicsServer) spawnPlayer(in UdpInput) {
	playerName := in.GetName();
	player, ok := ps.connectedPlayers.Load(in.GetName())
	p := player.(*UdpPlayer)
	if !ok {
		log.Printf("unable to find connected player %s", playerName)
	} else if p.GetState() != SPECTATING {
		log.Printf("player %s spawn when not spectating", playerName)
	}

	// OK, add body to simulation
	pBod := NewControlledUdpBody(p)
	x, y := ps.worldMap.GetCellCenter(spawnX, spawnY)
	pBod.SetPos(x, y)
	pBod.SetRot(0)
	ps.bodies = append(ps.bodies, pBod)

	// Map body <-> player for easier input assignment
	ps.controlMap.Store(playerName, pBod)

	// Tell client where to spawn
  var msg NetworkMsg
  msg.PutByte(byte(SSpawn))
  msg.PutUint16(pBod.GetId())
  msg.PutUint32(spawnX)
  msg.PutUint32(spawnY)

	conn := p.GetConnection()
	conn.SendTo(snet.GetDataFromNetworkMsg(&msg));
	log.Printf("Spawning %s at %d/%d -- %f/%f", p.GetName(), spawnX, spawnY, x, y)
}

func (ps *physicsServer) NewPlayer(id string) {
	player := NewUdpPlayer(id)
	player.SetStats(NewPlayerStats())

	_, exists := ps.connectedPlayers.LoadOrStore(id, player)
	if !exists {
		ps.playerCount += 1
		log.Printf("%s joined the simulation", id)
	}
}

func (ps *physicsServer) RemovePlayer(id string) {
	pBody, ok := ps.controlMap.Load(id)
	if ok {
		pBody.(*UdpBody).Kill()
	}
	ps.controlMap.Delete(id)
	ps.connectedPlayers.Delete(id)
	log.Printf("%s left the simulation", id)
}

func (ps *physicsServer) SyncPlayers(syncTime int64) {
	var syncMsg NetworkMsg
	syncMsg.PutByte(byte(SSync))
	syncMsg.PutUint64(uint64(syncTime  / (int64(time.Millisecond)/int64(time.Nanosecond))))

	syncMsgData := snet.GetDataFromNetworkMsg(&syncMsg)

	ps.lastSync = syncTime

	ps.connectedPlayers.Range(func(key, value interface{}) bool {
		// TODO: customize frame for each player
		p := value.(*UdpPlayer)
		conn := p.GetConnection()

		if p.IsActive() {
			p.Sync(syncTime)
			conn.SendTo(syncMsgData)
			log.Printf("Synced %s", p.GetName())
		}

		return true
	})
}


// Event Loop
/////////////////

func (ps *physicsServer) interpret(i UdpInput, c gnet.Conn) (out []byte) {
	playerName := i.GetName()
	p, ok := ps.connectedPlayers.Load(playerName)
	if !ok {
		log.Printf("Attempting to interpret command from unknown player %s", playerName)
	} else {
		player := p.(*UdpPlayer)
		// TODO: authenticate
		if i.GetType() == HELLO && !player.IsActive(){
			log.Printf("%s connected.", playerName)
			player.Activate()
			player.SetState(SPECTATING)
			player.SetConnection(c)
		}

		// If the player is controlling a body
		// add the input to the body.
		bod, ok := ps.controlMap.Load(playerName)
		if ok && bod != nil {
			bod.(*UdpBody).QueueInput(&i)
		} else {
			ps.inputs.Enqueue(i, uint8(ps.seq))
		}
	}

	return
}

func (ps *physicsServer) React(data []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	//log.Printf("%s", c.RemoteAddr())
	//data := append([]byte{}, c.ReadFromUDP()...)
  _ = ps.pool.Submit(func() {
    if len(data) >= 4 {
      msg := snet.GetNetworkMsgFromData(data)
      if msg != nil {
      	var i UdpInput
      	i.Deserialize(msg)
      	//log.Printf("%s", i.String())

      	out = ps.interpret(i, c);
      }
    }
  })

  return
}

func (ps *physicsServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Printf("UDP server is listening on %s (multi-cores: %t, loops: %d)\n",
		srv.Addr.String(), srv.Multicore, srv.NumLoops)
	return
}

// Process physics for one frame
func (ps *physicsServer) Simulate(frameStart int64) {
	// initialize the frame
	frame := NewUdpFrame(ps.seq)

	// Update all bodies
	// flag dead bodies for removal
	// process live bodies
	// replace ps.bodies with filtered list
	filteredBodies := ps.bodies[:0]
	for i, b := range ps.bodies {
		if !b.IsDead() {
			filteredBodies = append(filteredBodies, b)
		} else {
			ps.bodies[i] = nil
			continue
		}
		player := b.GetControllingPlayer()
		if player != nil && player.IsActive() {
			b.ProcessInput(frameStart / (int64(time.Millisecond)/int64(time.Nanosecond)))
		} else {
			log.Printf("player is nil on controlled body %d", b.GetId());
		}

		b.ApplyTransform()

		frame.AddUdpBody(b)
	}
	ps.bodies = filteredBodies

	if frame.Len() > 0 {
		serializedFrame := frame.Serialize()
		var frameMsg NetworkMsg
		frameMsg.PutByte(byte(SFrame))
		frameMsg.PutByte(ps.seq)
		frameMsg.PutBytes(serializedFrame)
		frameData := snet.GetDataFromNetworkMsg(&frameMsg)

		// Send the frame to relevent players.
		ps.connectedPlayers.Range(func(key, value interface{}) bool {
			// TODO: customize frame for each player
			p := value.(*UdpPlayer)
			if p.IsActive() {
				conn := p.GetConnection()
				conn.SendTo(frameData)
			}
			return true
		})
	}
}

// Simulation loop
func (ps *physicsServer) Simulation() {
	simulationStart := time.Now().UnixNano()
	ps.lastSync = simulationStart
	shouldSync := false

	for {
		frameStartTime := time.Now()
		frameStart := frameStartTime.UnixNano()
		if ps.playerCount > 0 {
			// Determine whether to process a simulation frame
			framesToProcess := ((frameStart - ps.lastSync) / ps.TIMESTEP) - ps.framesSinceLastSync
			if framesToProcess > 0 {
				for i := int64(0); i < framesToProcess; i++ {
					ps.seq++;
					ps.framesSinceLastSync++;
					ps.lastFrame = ps.lastSync + (ps.framesSinceLastSync * ps.TIMESTEP)
					ps.Simulate(ps.lastFrame)

					if ps.seq == 0 {
						shouldSync = true
					}
				}
			} else if shouldSync {
				log.Printf("SYNC")
				ps.SyncPlayers(ps.lastFrame)
				shouldSync = false
				ps.framesSinceLastSync = 0
			} else {
		  	nonBodyInputs := ps.inputs.Len()
		  	for i := 0; i < nonBodyInputs; i++ {
			  	in, err := ps.inputs.Dequeue()
			  	if err != nil {
			  		log.Printf(err.Error())
			  		break
			  	}

			  	inp := in.(UdpInput)

					if err == nil && inp.GetType() == SPAWN {
						ps.spawnPlayer(inp)
			  	}
		  	}
			}

		}

		elapsed := time.Since(frameStartTime);
		if ps.tick > elapsed {
			time.Sleep(ps.tick - elapsed)
		} else {
			time.Sleep(32)
		}
	}
}

func main() {
  p := goroutine.Default()
  defer p.Release()

  // connect via TCP to the world server
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 9499,
		},
	}
  conn, err := dialer.Dial("tcp", "127.0.0.1:9494")
  if err != nil {
  	log.Printf("%s", err)
  }
  defer conn.Close()

  // Initialize UDP server
  ps := &physicsServer{
  	pool: p,
  	bodies: []*UdpBody{},
  	tick: 8333333,
  	TIMESTEP: 34000000,
  	seq: 0,
  	launchTime: time.Now().UnixNano(),
  	lastSync: 0,
  	framesSinceLastSync: 0,
  	lastFrame: 0,
  	inputs: priorityqueue.NewHierarchicalQueue(255, true),
  	worldConn: conn,
  }

  // react to events from world server
  go ps.ReactWorld(conn)
  go ps.Simulation()

  log.Fatal(gnet.Serve(ps, udpAddr, gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithReusePort(true)))
}

// world server <-> physics server interaction
func (ps *physicsServer) ReactWorld(c net.Conn) {

	const(
		readingSize byte = iota
		readingWidth
		readingHeight
		readingResolution
		readingMap
		readingEvent
	)

	var state byte = readingSize
	var size int = 0
	var mapBuf []byte = nil
	var mapBufRead int = 0
	reader := bufio.NewReader(c)

	for {
		switch state {
			case readingSize:
				sizeBytes, err := reader.Peek(4)
				if err == nil {
					size = snet.Read_int32(sizeBytes)
					mapBuf = make([]byte, size - 4 - 4)

					reader.Discard(4)

					state = readingWidth
					log.Printf("Read worldMap data length=%d", size)
				} else {
					log.Printf(err.Error())
				}
			case readingWidth:
				widthBytes, err := reader.Peek(4)
				if err == nil {
					ps.worldMap.W = snet.Read_int32(widthBytes)

					reader.Discard(4)

					state = readingHeight
					log.Printf("Read worldMap width=%d", ps.worldMap.W)
				} else {
					log.Printf(err.Error())
				}
			case readingHeight:
				heightBytes, err := reader.Peek(4)
				if err == nil {
					ps.worldMap.H = snet.Read_int32(heightBytes)

					reader.Discard(4)

					state = readingResolution
					log.Printf("Read worldMap height=%d", ps.worldMap.H)
				} else {
					log.Printf(err.Error())
				}
			case readingResolution:
				resBytes, err := reader.Peek(1)
				if err == nil {
					ps.worldMap.Resolution = int(resBytes[0])

					reader.Discard(1)

					state = readingMap
					log.Printf("Read worldMap resolution=%d", ps.worldMap.Resolution)
				} else {
					log.Printf(err.Error())
				}
			case readingMap:
				for b, err := reader.Peek(1); err == nil && mapBufRead < size - 4 - 4 - 1; {
					mapBuf[mapBufRead] = b[0]
					mapBufRead++
					reader.Discard(1)
				}

				if mapBufRead == size - 4 - 4 - 1 {
					ps.worldMap.Deserialize(mapBuf, false)
					reader.Reset(c)

					state = readingEvent
					log.Printf("Read worldMap bytes")

					var msg NetworkMsg
					msg.PutByte(byte(IReady))
					msg.PutUint32(udpPort)
					c.Write(snet.GetDataFromNetworkMsg(&msg))
				}
			case readingEvent:
				eventSizeBytes, err := reader.Peek(4)
				if err == nil {
					log.Printf("Received event from world")
					eventSize := snet.Read_int32(eventSizeBytes)
					reader.Discard(4)

					eventContent, err2 := reader.Peek(eventSize)
					if err2 == nil {
						if eventContent[0] == byte(IJoin) {
							playerIdLenBytes := eventContent[1:3]
							playerIdLen := snet.Read_uint16(playerIdLenBytes)
							playerId := snet.Read_utf8(eventContent[3:3+playerIdLen])
							ps.NewPlayer(playerId)
						} else if eventContent[0] == byte(ILeave) {
							playerIdLenBytes := eventContent[1:3]
							playerIdLen := snet.Read_uint16(playerIdLenBytes)
							playerId := snet.Read_utf8(eventContent[3:3+playerIdLen])
							ps.RemovePlayer(playerId)
						}
					}
					reader.Discard(eventSize)
				} else {
					log.Printf("%s", err.Error())
				}
		}
		time.Sleep(1 * time.Second)
	}
}
