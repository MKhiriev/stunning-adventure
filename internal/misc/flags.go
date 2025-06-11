package misc

import (
	"errors"
	"flag"
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"net"
	"strconv"
	"strings"
	"time"
)

var ServerAddress NetAddress
var PollInterval time.Duration
var ReportInterval time.Duration

type NetAddress struct {
	Host string
	Port int
}

func ParseServerFlags() {
	ServerAddress = NetAddress{}
	// compile-time check: does NetAddress implement flag.Value interface?
	_ = flag.Value(&ServerAddress)

	// extract value
	flag.Var(&ServerAddress, "a", "Net address host:port")

	flag.Parse()
}

func ParseAgentFlags() {
	ServerAddress = NetAddress{}
	// compile-time check: does NetAddress implement flag.Value interface?
	_ = flag.Value(&ServerAddress)
	flag.Var(&ServerAddress, "a", "Net address host:port")

	flag.DurationVar(&PollInterval, "p", agent.DefaultPollInterval, "Poll interval in seconds")
	flag.DurationVar(&ReportInterval, "r", agent.DefaultReportInterval, "Report interval in seconds")

	flag.Parse()
}

func (a *NetAddress) String() string {
	// by default: localhost:8080
	if a.Host == "" && a.Port == 0 {
		return server.DefaultServerAddress
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

type TimeInterval struct {
	Seconds int64
}

func (i *TimeInterval) Duration() time.Duration {
	return time.Duration(i.Seconds)
}

func (i *TimeInterval) String() string {
	return strconv.FormatInt(i.Seconds, 10)
}

func (i *TimeInterval) Set(s string) error {
	if s == "" {
		i.Seconds = 2
		return nil
	}

	seconds, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	if seconds < 1 {
		return errors.New("time interval should be a positive integer")
	}

	i.Seconds = seconds
	return nil
}
