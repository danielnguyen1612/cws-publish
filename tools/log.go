package tools

import (
	"fmt"

	shutdown "github.com/klauspost/shutdown2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var isInitialized bool

// InitLogging initialize a zap Logger that is docker friendly and a function to trap error and send panic trace if
func InitLogging() *zap.Logger {
	if isInitialized {
		panic("Already initialized (InitLogging)")
	}

	viper.SetDefault("log.level", zap.DebugLevel.String())
	viper.SetDefault("log.timestamp", true)

	var (
		level    zapcore.Level
		levelStr = viper.GetString("log.level")
		levelErr error
		timeKey  string
	)

	if viper.GetBool("log.timestamp") {
		timeKey = "ts"
	}

	if levelErr = level.UnmarshalText([]byte(levelStr)); levelErr != nil {
		level = zap.DebugLevel
	}

	conf := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        timeKey,
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, err := conf.Build()
	if err != nil {
		panic(err)
	}

	if levelErr != nil {
		log.Error("Couldn't parse logging level, switch to debug", zap.String("level", levelStr), zap.Error(levelErr))
	}

	sl := log.Sugar()
	shutdown.SetLogPrinter(sl.Infof)

	return log
}

// RecoverLog log the panic as an error
func RecoverLog(log *zap.Logger, f func()) {
	defer func() {
		var field zapcore.Field
		err := recover()
		switch rval := err.(type) {
		case nil:
			return
		case error:
			field = zap.Error(rval)
		default:
			field = zap.String("error_raw", fmt.Sprint(rval))
		}
		log.Error("Recovered panic", field)
	}()

	f()
}
