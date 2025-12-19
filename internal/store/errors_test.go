package store

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresErrorClassifier_Classify(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	tests := []struct {
		name string
		err  error
		want ErrorClassification
	}{
		{
			name: "nil error is non-retryable",
			err:  nil,
			want: NonRetryable,
		},
		{
			name: "non postgres error is non-retryable",
			err:  errors.New("some random error"),
			want: NonRetryable,
		},
		{
			name: "postgres retryable error is classified correctly",
			err: &pgconn.PgError{
				Code: pgerrcode.ConnectionFailure,
			},
			want: Retryable,
		},
		{
			name: "postgres non-retryable error is classified correctly",
			err: &pgconn.PgError{
				Code: pgerrcode.UniqueViolation,
			},
			want: NonRetryable,
		},
	}

	for _, tt := range tests {
		tt := tt // защита от capture
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := classifier.Classify(tt.err)
			if got != tt.want {
				t.Fatalf("Classify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyPgError(t *testing.T) {
	tests := []struct {
		name string
		code string
		want ErrorClassification
	}{
		// --- Retryable ---
		{
			name: "connection exception",
			code: pgerrcode.ConnectionException,
			want: Retryable,
		},
		{
			name: "connection failure",
			code: pgerrcode.ConnectionFailure,
			want: Retryable,
		},
		{
			name: "serialization failure",
			code: pgerrcode.SerializationFailure,
			want: Retryable,
		},
		{
			name: "deadlock detected",
			code: pgerrcode.DeadlockDetected,
			want: Retryable,
		},
		{
			name: "cannot connect now",
			code: pgerrcode.CannotConnectNow,
			want: Retryable,
		},

		// --- NonRetryable ---
		{
			name: "data exception",
			code: pgerrcode.DataException,
			want: NonRetryable,
		},
		{
			name: "not null violation",
			code: pgerrcode.NotNullViolation,
			want: NonRetryable,
		},
		{
			name: "foreign key violation",
			code: pgerrcode.ForeignKeyViolation,
			want: NonRetryable,
		},
		{
			name: "syntax error",
			code: pgerrcode.SyntaxError,
			want: NonRetryable,
		},
		{
			name: "undefined table",
			code: pgerrcode.UndefinedTable,
			want: NonRetryable,
		},
		{
			name: "unknown error code is non-retryable",
			code: "XXXXX",
			want: NonRetryable,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pgErr := &pgconn.PgError{Code: tt.code}
			got := ClassifyPgError(pgErr)

			if got != tt.want {
				t.Fatalf("ClassifyPgError(%s) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
