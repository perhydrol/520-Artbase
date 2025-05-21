package log

type LogConfig struct {
	DisableCaller     bool
	DisableStacktrace bool
	Level             string
	Encoding          string
	OutputPaths       []string
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		DisableCaller:     false,
		DisableStacktrace: false,
		Level:             "debug",
		Encoding:          "console",
		OutputPaths:       []string{"stdout"},
	}
}
