package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Memory uint32

	Time uint32

	Threads uint8
}

var OWASP = Params{Memory: 19 * 1024, Time: 2, Threads: 1}

const saltLength = 16

const keyLength = 32

const minSaltLength = 8

const maxMemory = 256 * 1024

const maxTime = 16

var errMalformed = errors.New("the password hash is not an argon2id hash in the PHC format")

var errParams = errors.New("the argon2id params are out of range")

var encoding = base64.RawStdEncoding

var version = "v=" + strconv.Itoa(argon2.Version)

type hash struct {
	params Params

	salt []byte

	key []byte
}

func (p Params) Hash(password string) (string, error) {
	if p.valid() == false {
		return "", errParams
	}

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, keyLength)

	return "$argon2id$" + version + "$" + p.phc() + "$" + encoding.EncodeToString(salt) + "$" + encoding.EncodeToString(key), nil
}

func (p Params) Verify(password, encoded string) (ok, rehash bool, err error) {
	h, err := parse(encoded)
	if err != nil {
		return false, false, err
	}

	key := argon2.IDKey([]byte(password), h.salt, h.params.Time, h.params.Memory, h.params.Threads, keyLength)
	if subtle.ConstantTimeCompare(key, h.key) != 1 {
		return false, false, nil
	}

	return true, h.params != p || len(h.salt) != saltLength, nil
}

func parse(encoded string) (hash, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != version {
		return hash{}, errMalformed
	}

	params, err := parseParams(parts[3])
	if err != nil {
		return hash{}, err
	}

	salt, err := encoding.DecodeString(parts[4])
	if err != nil || len(salt) < minSaltLength {
		return hash{}, errMalformed
	}

	key, err := encoding.DecodeString(parts[5])
	if err != nil || len(key) != keyLength {
		return hash{}, errMalformed
	}

	return hash{params: params, salt: salt, key: key}, nil
}

func parseParams(text string) (Params, error) {
	var p Params
	if _, err := fmt.Sscanf(text, "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Threads); err != nil || p.phc() != text || p.valid() == false {
		return Params{}, errMalformed
	}

	return p, nil
}

func (p Params) valid() bool {
	return p.Memory <= maxMemory && p.Time >= 1 && p.Time <= maxTime && p.Threads >= 1
}

func (p Params) phc() string {
	return fmt.Sprintf("m=%d,t=%d,p=%d", p.Memory, p.Time, p.Threads)
}
