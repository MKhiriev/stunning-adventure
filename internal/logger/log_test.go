package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLogger_ReturnsNonNilLogger(t *testing.T) {
	log := NewLogger("test")
	require.NotNil(t, log)
}
