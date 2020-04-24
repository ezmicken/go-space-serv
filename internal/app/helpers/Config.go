package helpers

type Config struct {
  TIMESTEP int64
  TIMESTEP_NANO int64
}

var configInstance *Config

func SetConfig(inst *Config)            { configInstance = inst }
func GetConfig() *Config                { return configInstance }
func GetConfiguredTimestep()      int64 { return configInstance.TIMESTEP }
func GetConfiguredTimestepNanos() int64 { return configInstance.TIMESTEP_NANO }
