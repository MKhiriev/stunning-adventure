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
	defaultFileStoragePath = "tmp"
	defaultRestoreValue    = true
	//defaultDatabaseDSN     = "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"
	defaultDatabaseDSN = ""
)

type NetAddress struct {
	Host string
	Port int
}

func ParseServerFlags() (netAddress string, storeInterval int64, fileStoragePath string, restore bool, databaseDSN string) {
	serverAddress := NetAddress{}
	_ = flag.Value(&serverAddress)

	flag.Var(&serverAddress, "a", "Net address host:port")
	flag.Int64Var(&storeInterval, "i", defaultStoreInterval, "Store interval in seconds")
	flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "Storage file path string")
	flag.BoolVar(&restore, "r", defaultRestoreValue, "Boolean - restore previous metrics from file")
	flag.StringVar(&databaseDSN, "d", defaultDatabaseDSN, "Postgres database connection string")

	flag.Parse()

	return serverAddress.String(), storeInterval, fileStoragePath, restore, databaseDSN
}

func ParseAgentFlags() (netAddress string, pollInterval int64, reportInterval int64) {
	serverAddress := NetAddress{}
	_ = flag.Value(&serverAddress)

	flag.Var(&serverAddress, "a", "Net address host:port")
	flag.Int64Var(&pollInterval, "p", defaultPollInterval, "Poll interval in seconds")
	flag.Int64Var(&reportInterval, "r", defaultReportInterval, "Report interval in seconds")

	flag.Parse()

	return serverAddress.String(), pollInterval, reportInterval
}

func (a *NetAddress) String() string {
	if a.Host == "" && a.Port == 0 {
		return defaultServerAddress
	}

	return a.Host + ":" + strconv.Itoa(a.Port)
}

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
