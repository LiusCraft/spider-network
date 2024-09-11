package xlog

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal

	levelDebugStr = "DEBUG"
	levelInfoStr  = "INFO"
	levelWarnStr  = "WARN"
	levelErrorStr = "ERROR"
	levelFatalStr = "FATAL"
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return levelDebugStr
	case LevelInfo:
		return levelInfoStr
	case LevelWarn:
		return levelWarnStr
	case LevelError:
		return levelErrorStr
	case LevelFatal:
		return levelFatalStr
	default:
		return "unknown"
	}
}
