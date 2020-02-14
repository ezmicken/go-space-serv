package world

type ServerState byte

const (
	WAIT_PHYS ServerState = iota	// Program was just launched.
	SETUP_PHYS										// Waiting for physics server to connect & receive data
	RUNNING												// Accepting player connections, sending blocks
	SHUTDOWN											// Shutting down the server.
)
