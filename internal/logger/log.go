// Package logger provides a thin wrapper around zerolog for structured logging.
// It configures global logging behavior, caller formatting, and per-role metadata.
package logger

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger to allow future extensions without modifying callers.
// Embedding permits access to the full zerolog API.
type Logger struct {
	*zerolog.Logger
}

// NewLogger constructs and returns a configured zerolog.Logger.
// It applies global debug-level logging, custom caller formatting (file:line),
// timestamp injection, and a constant "role" field for component identification.
//
// Caller output is reduced to the base filename for compact log lines.
//
// Example:
//
//	log := logger.NewLogger("agent")
//	log.Info().Msg("started")
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
