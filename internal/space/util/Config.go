package helpers

type Config struct {
  TIMESTEP int64
  TIMESTEP_NANO int64
  WORLD_RATE int
  NAME string
  VERSION string
  PROTOCOL_ID uint32
  MAX_MSG_SIZE int
}

var configInstance *Config

func SetConfig(inst *Config)              { configInstance = inst }
func GetConfig() *Config                  { return configInstance }
func GetConfiguredTimestep()      int64   { return configInstance.TIMESTEP }
func GetConfiguredTimestepNanos() int64   { return configInstance.TIMESTEP_NANO }
func GetConfiguredWorldRate()     int     { return configInstance.WORLD_RATE }
func GetProtocolId()              uint32  { return configInstance.PROTOCOL_ID }
