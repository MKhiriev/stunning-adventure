package config

import (
	"errors"
	"flag"
	"net"
	"strconv"
	"strings"
)

// NetAddress holds structured network address data for host and port.
// It implements the flag.Value interface.
type NetAddress struct {
	Host string
	Port int
}

// ParseServerFlags parses all server-related configuration flags.
// It returns the resolved network address, store interval, file storage path,
// restore flag, database DSN, hash key, and audit endpoints.
//
// Flags:
//
//	-a server address in format [host]:[port]
//	-i store interval in seconds
//	-f file storage path
//	-r restore from file
//	-d database DSN
//	-k hash key
//	-audit-file path to audit log file
//	-audit-url audit server endpoint
//	-crypto-key public key path
//	-c/-config json file path with configs
func ParseServerFlags() *ServerConfig {
	var serverAddress NetAddress
	var fileStoragePath, databaseDSN, hashKey, auditFilePath, auditURL, privateKeyFilePath string
	var storeInterval int64
	var restore bool
	var jsonConfigFilePath string

	flag.Var(&serverAddress, "a", "Net address host:port")
	flag.Int64Var(&storeInterval, "i", 0, "Store interval in seconds")
	flag.StringVar(&fileStoragePath, "f", "", "Storage file path string")
	flag.BoolVar(&restore, "r", false, "Boolean - restore previous metrics from file")
	flag.StringVar(&databaseDSN, "d", "", "Postgres database connection string")
	flag.StringVar(&hashKey, "k", "", "Hash key for hashing")
	flag.StringVar(&auditFilePath, "audit-file", "", "Audit file path string")
	flag.StringVar(&auditURL, "audit-url", "", "Full Audit URL string")
	flag.StringVar(&privateKeyFilePath, "crypto-key", "", "Private key file path")

	flag.StringVar(&jsonConfigFilePath, "config", "", "JSON config file path")
	flag.StringVar(&jsonConfigFilePath, "c", "", "JSON config file path")

	flag.Parse()

	return &ServerConfig{
		ServerAddress:          serverAddress.String(),
		StoreInterval:          storeInterval,
		FileStoragePath:        fileStoragePath,
		RestoreMetricsFromFile: restore,
		DatabaseDSN:            databaseDSN,
		HashKey:                hashKey,
		AuditFile:              auditFilePath,
		AuditURL:               auditURL,
		PrivateCryptoKeyPath:   privateKeyFilePath,
		JSONConfigFile:         jsonConfigFilePath,
	}
}

// parseAgentFlags parses all agent-related configuration flags.
// It returns the resolved server address, poll interval, report interval,
// hashing key, and request rate limit.
//
// Flags:
//
//	-a host:port
//	-p poll interval in seconds
//	-r report interval in seconds
//	-k hash key
//	-l concurrency limit
//	-crypto-key public key path
//	-c/-config json file path with configs
func parseAgentFlags() *AgentConfig {
	var hashKey, publicKeyFilePath string
	var pollInterval, reportInterval, rateLimit int64
	var serverAddress NetAddress
	var jsonConfigFilePath string

	flag.Var(&serverAddress, "a", "Net address \"host:port\" format")
	flag.Int64Var(&pollInterval, "p", 0, "Poll interval in seconds")
	flag.Int64Var(&reportInterval, "r", 0, "Report interval in seconds")
	flag.StringVar(&hashKey, "k", "", "Hash key for hashing")
	flag.Int64Var(&rateLimit, "l", 0, "Concurrent request limit to the server")
	flag.StringVar(&publicKeyFilePath, "crypto-key", "", "Public key file path")

	flag.StringVar(&jsonConfigFilePath, "config", "", "JSON config file path")
	flag.StringVar(&jsonConfigFilePath, "c", "", "JSON config file path")

	flag.Parse()

	return &AgentConfig{
		ServerAddress:       serverAddress.String(),
		ReportInterval:      reportInterval,
		PollInterval:        pollInterval,
		HashKey:             hashKey,
		RateLimit:           rateLimit,
		PublicCryptoKeyPath: publicKeyFilePath,
		JSONConfigFile:      jsonConfigFilePath,
	}
}

// String returns a canonical host:port string for a NetAddress.
// If neither Host nor Port are set, it returns the default server address.
func (a *NetAddress) String() string {
	if a.Host == "" && a.Port == 0 {
		return ""
	}

	return a.Host + ":" + strconv.Itoa(a.Port)
}

// Set parses the input string of form host:port and populates the NetAddress.
// It validates the port range, checks IP correctness unless host is "localhost",
// and returns an error if the format or values are invalid.
func (a *NetAddress) Set(s string) error {
	hostAndPort := strings.Split(s, ":")
	if len(hostAndPort) != 2 {
		return errors.New("need address in a form `host:port`")
	}

	host := hostAndPort[0]
	port, err := strconv.Atoi(hostAndPort[1])
	if err != nil {
		return err
	}

	if port < 1 {
		return errors.New("port number is a positive integer")
	}

	if host != "localhost" {
		ip := net.ParseIP(hostAndPort[0])
		if ip == nil {
			return errors.New("incorrect IP-address provided")
		}
	}

	a.Host = host
	a.Port = port
	return nil
}
