package snet

type ServerState byte

const (
  DEAD ServerState = iota       // Server is dead.
  WAIT_PHYS                     // Waiting on a physics connection.
  WAIT_WORLD                    // Waiting on a world connection
  SETUP                         // Exchanging data
  ALIVE                         // Accepting player connections
  SHUTDOWN                      // Shutting down the server.

)
