package password_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"

	"gokick/app/core/password"
)

const secret = "correct horse"

var cheap = password.Params{Memory: 64, Time: 1, Threads: 1}

func TestHashWritesThePHCFormat(t *testing.T) {
	encoded := hash(t, cheap)

	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=64,t=1,p=1" || len(parts[4]) != 22 || len(parts[5]) != 43 {
		t.Errorf("hash %s", encoded)
	}

	if again := hash(t, cheap); again == encoded {
		t.Error("two hashes share the salt")
	}
}

func TestVerifyChecksThePassword(t *testing.T) {
	encoded := hash(t, cheap)

	if ok, rehash, err := cheap.Verify(secret, encoded); ok == false || rehash || err != nil {
		t.Errorf("right password: ok %t, rehash %t, error %v", ok, rehash, err)
	}

	if ok, rehash, err := cheap.Verify("correct horsE", encoded); ok || rehash || err != nil {
		t.Errorf("wrong password: ok %t, rehash %t, error %v", ok, rehash, err)
	}
}

func TestVerifyAsksForARehashWhenTheParamsChange(t *testing.T) {
	encoded := hash(t, cheap)

	for _, params := range []password.Params{{
		Memory: 128,

		Time: 1,

		Threads: 1,
	}, {Memory: 64, Time: 2, Threads: 1}, {Memory: 64, Time: 1, Threads: 2}} {
		if ok, rehash, err := params.Verify(secret, encoded); ok == false || rehash == false || err != nil {
			t.Errorf("%+v: ok %t, rehash %t, error %v", params, ok, rehash, err)
		}
	}

	salt := []byte("8 bytes!")
	key := argon2.IDKey([]byte(secret), salt, 1, 64, 1, 32)

	short := "$argon2id$v=19$m=64,t=1,p=1$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key)
	if ok, rehash, err := cheap.Verify(secret, short); ok == false || rehash == false || err != nil {
		t.Errorf("a shorter salt: ok %t, rehash %t, error %v", ok, rehash, err)
	}
}

func TestVerifyRefusesAMalformedHash(t *testing.T) {
	parts := strings.Split(hash(t, cheap), "$")
	salt, key := parts[4], parts[5]

	for name, encoded := range map[string]string{
		"empty": "",

		"argon2i": "$argon2i$v=19$m=64,t=1,p=1$" + salt + "$" + key,

		"another version": "$argon2id$v=16$m=64,t=1,p=1$" + salt + "$" + key,

		"missing a part": "$argon2id$v=19$m=64,t=1,p=1$" + salt,

		"an extra part": "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$" + key + "$",

		"text before": "x$argon2id$v=19$m=64,t=1,p=1$" + salt + "$" + key,

		"params in another order": "$argon2id$v=19$t=1,m=64,p=1$" + salt + "$" + key,

		"missing a param": "$argon2id$v=19$m=64,t=1$" + salt + "$" + key,

		"memory not a number": "$argon2id$v=19$m=x,t=1,p=1$" + salt + "$" + key,

		"no rounds": "$argon2id$v=19$m=64,t=0,p=1$" + salt + "$" + key,

		"no threads": "$argon2id$v=19$m=64,t=1,p=0$" + salt + "$" + key,

		"too many threads": "$argon2id$v=19$m=64,t=1,p=256$" + salt + "$" + key,

		"salt not base64": "$argon2id$v=19$m=64,t=1,p=1$" + salt + "!$" + key,

		"salt too short": "$argon2id$v=19$m=64,t=1,p=1$" + salt[:10] + "$" + key,

		"no key": "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$",

		"key with padding": "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$" + key + "=",

		"key too short": "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$" + key[:40],

		"key too long": "$argon2id$v=19$m=64,t=1,p=1$" + salt + "$" + key + "AAAA",

		"params with a space": "$argon2id$v=19$m= 64,t=1,p=1$" + salt + "$" + key,

		"params with a leading zero": "$argon2id$v=19$m=064,t=1,p=1$" + salt + "$" + key,

		"params with text after": "$argon2id$v=19$m=64,t=1,p=1,x=1$" + salt + "$" + key,

		"memory too large": "$argon2id$v=19$m=4294967296,t=1,p=1$" + salt + "$" + key,

		"memory over the limit": "$argon2id$v=19$m=262145,t=1,p=1$" + salt + "$" + key,

		"rounds over the limit": "$argon2id$v=19$m=64,t=17,p=1$" + salt + "$" + key,
	} {
		if ok, rehash, err := cheap.Verify(secret, encoded); ok || rehash || err == nil {
			t.Errorf("%s: ok %t, rehash %t, error %v", name, ok, rehash, err)
		}
	}
}

func TestHashRefusesParamsOutOfRange(t *testing.T) {
	for name, params := range map[string]password.Params{
		"no threads": {Memory: 64, Time: 1},

		"no rounds": {Memory: 64, Threads: 1},

		"rounds over the limit": {Memory: 64, Time: 17, Threads: 1},

		"memory over the limit": {Memory: 256*1024 + 1, Time: 1, Threads: 1},
	} {
		if encoded, err := params.Hash(secret); err == nil {
			t.Errorf("%s: hash %s", name, encoded)
		}
	}
}

func hash(t *testing.T, params password.Params) string {
	t.Helper()

	encoded, err := params.Hash(secret)
	if err != nil {
		t.Fatal(err)
	}

	return encoded
}
