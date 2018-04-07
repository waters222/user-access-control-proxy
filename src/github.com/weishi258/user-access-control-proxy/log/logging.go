package log
import (
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap"
	"os"
)

var bProduction bool
var logger *zap.Logger
var loggerSugar *zap.SugaredLogger

func InitZapLogger(logLevel string, isProduction bool) *zap.Logger{
	bProduction = isProduction
	var levelEnabler zap.LevelEnablerFunc
	switch(logLevel) {
	case "debug":
		levelEnabler = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {return lvl >= zapcore.DebugLevel})
	case "info":
		levelEnabler = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {return lvl >= zapcore.InfoLevel})
	case "warn":
		levelEnabler = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {return lvl >= zapcore.WarnLevel})
	case "error":
		levelEnabler = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {return lvl >= zapcore.ErrorLevel})
	case "fatal":
		levelEnabler = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {return lvl >= zapcore.FatalLevel})
	case "panic":
		levelEnabler = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {return lvl >= zapcore.PanicLevel})
	default:
		levelEnabler = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {return lvl >= zapcore.InfoLevel})
	}

	consoleOut := zapcore.Lock(os.Stdout)
	var consoleEncoder zapcore.Encoder
	if bProduction {
		consoleEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}else {
		consoleEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleOut, levelEnabler),
	)
	logger = zap.New(core)
	loggerSugar = logger.Sugar()
	return logger
}

func GetLogger() *zap.Logger{
	return logger
	//pc := make([]uintptr, 1)  // at least 1 entry needed
	//runtime.Callers(2, pc)
	//f := runtime.FuncForPC(pc[0])
	//if bProduction{
	//	return logger.With(zap.String("caller", f.Name()))
	//}else{
	//	file, line := f.FileLine(pc[0])
	//	return logger.With(zap.String("caller", f.Name()),
	//	zap.String("file", file),
	//	zap.Int("line", line))
	//}
}
