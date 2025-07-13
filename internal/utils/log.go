package utils

import (
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"strconv"
)

func NewLogger(role string) *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("role", role).
		Logger()

	return &logger
}
