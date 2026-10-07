package smtp

import (
	"cmp"
	"errors"
	"fmt"
	"net"
	netmail "net/mail"
	"os"
	"strconv"
	"strings"

	"gokick/app/core/mail"
)

const defaultPort = 587

const implicitTLSPort = 465

var errHalfAccount = errors.New("SMTP_USERNAME and SMTP_PASSWORD are set together or not at all")

type Config struct {
	Enabled bool

	Host string

	Port uint16

	Username string

	Password string

	SenderEmail string

	SenderName string
}

func Read() (Config, error) {
	enabled, err := parseEnabled(os.Getenv("SMTP_ENABLED"))
	if err != nil {
		return Config{}, err
	}

	if enabled == false {
		return Config{}, nil
	}

	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return Config{}, errors.New("SMTP_HOST is required with SMTP_ENABLED=true")
	}

	if strings.ContainsAny(host, "/ ") || (strings.Contains(host, ":") && net.ParseIP(host) == nil) {
		return Config{}, fmt.Errorf("SMTP_HOST must be a host name without a scheme or port, got %q", host)
	}

	port, err := parsePort(cmp.Or(os.Getenv("SMTP_PORT"), strconv.Itoa(defaultPort)))
	if err != nil {
		return Config{}, err
	}

	if port == implicitTLSPort {
		return Config{}, errors.New("SMTP_PORT 465 starts with TLS, which the app does not speak; use 587 with STARTTLS")
	}

	cfg := Config{
		Enabled: true,

		Host: host,

		Port: port,

		Username: os.Getenv("SMTP_USERNAME"),

		Password: os.Getenv("SMTP_PASSWORD"),

		SenderEmail: os.Getenv("SMTP_SENDER_EMAIL"),

		SenderName: os.Getenv("SMTP_SENDER_NAME"),
	}
	if (cfg.Username == "") != (cfg.Password == "") {
		return Config{}, errHalfAccount
	}

	if sender, err := netmail.ParseAddress(cfg.SenderEmail); err != nil || sender.Address != cfg.SenderEmail {
		return Config{}, fmt.Errorf("SMTP_SENDER_EMAIL must be an email address like noreply@gokick.dev, got %q", cfg.SenderEmail)
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.FormatUint(uint64(c.Port), 10))
}

func (c *Config) Server() mail.SMTPServer {
	if c.Enabled == false {
		return mail.SMTPServer{}
	}

	return mail.SMTPServer{Addr: c.Addr(), Username: c.Username, Password: c.Password}
}

func (c *Config) From() string {
	return (&netmail.Address{Name: c.SenderName, Address: c.SenderEmail}).String()
}

func parseEnabled(value string) (bool, error) {
	switch value {
	case "", "false":
		return false, nil
	case "true":
		return true, nil
	default:
		return false, fmt.Errorf("SMTP_ENABLED must be true or false, got %q", value)
	}
}

func parsePort(value string) (uint16, error) {
	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil || port == 0 {
		return 0, fmt.Errorf("SMTP_PORT must be a number from 1 to 65535, got %q", value)
	}

	return uint16(port), nil
}
