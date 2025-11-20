package logger

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
)

type Logger struct {
	*zerolog.Logger
}

func NewLogger(role string) *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("role", role).
		Caller().
		Logger()

	return &logger
}
