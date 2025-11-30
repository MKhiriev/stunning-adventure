package config

import (
	"errors"
	"flag"
	"net"
	"strconv"
	"strings"
)

const (
	defaultPollInterval    = int64(2)
	defaultReportInterval  = int64(5)
	defaultServerAddress   = "localhost:8080"
	defaultStoreInterval   = int64(300)
	defaultFileStoragePath = ""
	defaultRestoreValue    = false
	defaultDatabaseDSN     = ""
	defaultHashKey         = ""
	defaultRateLimit       = int64(1)
	defaultAuditFilePath   = ""
	defaultAuditURL        = ""
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
func ParseServerFlags() (netAddress string, storeInterval int64, fileStoragePath string, restore bool, databaseDSN string, hashKey string, auditFilePath string, auditURL string) {
	serverAddress := NetAddress{}

	flag.Var(&serverAddress, "a", "Net address host:port")
	flag.Int64Var(&storeInterval, "i", defaultStoreInterval, "Store interval in seconds")
	flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "Storage file path string")
	flag.BoolVar(&restore, "r", defaultRestoreValue, "Boolean - restore previous metrics from file")
	flag.StringVar(&databaseDSN, "d", defaultDatabaseDSN, "Postgres database connection string")
	flag.StringVar(&hashKey, "k", defaultHashKey, "Hash key for hashing")

	flag.StringVar(&auditFilePath, "audit-file", defaultAuditFilePath, "Audit file path string")
	flag.StringVar(&auditURL, "audit-url", defaultAuditURL, "Full Audit URL string")

	flag.Parse()

	return serverAddress.String(), storeInterval, fileStoragePath, restore, databaseDSN, hashKey, auditFilePath, auditURL
}

// ParseAgentFlags parses all agent-related configuration flags.
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
func ParseAgentFlags() (netAddress string, pollInterval int64, reportInterval int64, hashKey string, rateLimit int64) {
	serverAddress := NetAddress{}

	flag.Var(&serverAddress, "a", "Net address host:port")
	flag.Int64Var(&pollInterval, "p", defaultPollInterval, "Poll interval in seconds")
	flag.Int64Var(&reportInterval, "r", defaultReportInterval, "Report interval in seconds")
	flag.StringVar(&hashKey, "k", defaultHashKey, "Hash key for hashing")
	flag.Int64Var(&rateLimit, "l", defaultRateLimit, "Concurrent request limit to the server")

	flag.Parse()

	return serverAddress.String(), pollInterval, reportInterval, hashKey, rateLimit
}

// String returns a canonical host:port string for a NetAddress.
// If neither Host nor Port are set, it returns the default server address.
func (a *NetAddress) String() string {
	if a.Host == "" && a.Port == 0 {
		return defaultServerAddress
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
