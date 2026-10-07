package database

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const defaultPort = 5432

var ErrInvalidPort = errors.New("must be a number from 1 to 65535")

var ErrInvalidParams = errors.New("must be empty or a query after ?, like ?sslmode=disable")

type Address struct {
	Host string

	Port uint16

	Name string

	Username string

	Password string

	Params url.Values
}

func ParsePort(value string) (uint16, error) {
	if value == "" {
		return defaultPort, nil
	}

	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil || port == 0 {
		return 0, ErrInvalidPort
	}

	return uint16(port), nil
}

func ParseParams(value string) (url.Values, error) {
	if value == "" {
		return url.Values{}, nil
	}

	query, found := strings.CutPrefix(value, "?")
	if found == false {
		return nil, ErrInvalidParams
	}

	params, err := url.ParseQuery(query)
	if err != nil {
		return nil, ErrInvalidParams
	}

	return params, nil
}

func (a *Address) URL() string {
	address := url.URL{
		Scheme: "postgres",

		User: url.UserPassword(a.Username, a.Password),

		Host: net.JoinHostPort(a.Host, strconv.FormatUint(uint64(a.Port), 10)),

		Path: "/" + a.Name,

		RawQuery: a.Params.Encode(),
	}

	return address.String()
}
