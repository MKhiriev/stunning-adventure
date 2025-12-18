package config

import "errors"

// ServerConfig
//
//	Configuration for the server component.
//	Controls server address, storage intervals, file persistence, database connection, hashing, and audit settings.
type ServerConfig struct {
	ServerAddress          string `env:"ADDRESS"`           // network address to bind the server
	StoreInterval          int64  `env:"STORE_INTERVAL"`    // interval in seconds to persist metrics to storage
	FileStoragePath        string `env:"FILE_STORAGE_PATH"` // file path for metric storage
	RestoreMetricsFromFile bool   `env:"RESTORE"`           // flag to restore metrics from file on startup
	DatabaseDSN            string `env:"DATABASE_DSN"`      // database connection string
	HashKey                string `env:"KEY"`               // key used for hashing metric payloads

	AuditFile string `env:"AUDIT_FILE"` // local file path for audit logs
	AuditURL  string `env:"AUDIT_URL"`  // external URL to send audit logs

	PrivateCryptoKeyPath string `env:"CRYPTO_KEY"` // private key path for decryption of data from agent
}

// validate
//
// Description:
//
//	Validates essential fields of ServerConfig.
//	Ensures ServerAddress and StoreInterval are set.
//
// Returns:
//
//	error - descriptive error if validation fails, nil if valid
func (s *ServerConfig) validate() error {
	switch {
	case s.ServerAddress == "":
		return errors.New("invalid Server Address")
	case s.StoreInterval == 0:
		return errors.New("invalid Store Interval")
	}

	return nil
}
