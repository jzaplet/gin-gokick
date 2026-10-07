package dotenv

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const key = "TESTKIT_URL"

func TestTheFirstKeyInTheEnvironmentTakesEveryKeyFromThere(t *testing.T) {
	t.Setenv(key, "postgres://env")
	t.Setenv("TESTKIT_PASSWORD", "")
	root := module(t, "TESTKIT_URL=postgres://dotenv\nTESTKIT_PASSWORD=dotenv\n")

	got, err := Lookup(root, key, "TESTKIT_PASSWORD")
	if err != nil || reflect.DeepEqual(got, map[string]string{key: "postgres://env", "TESTKIT_PASSWORD": ""}) == false {
		t.Errorf("got %v %v", got, err)
	}
}

func TestWithoutTheFirstKeyEveryKeyComesFromDotenv(t *testing.T) {
	t.Setenv(key, "")
	t.Setenv("TESTKIT_PASSWORD", "env")
	root := module(t, "TESTKIT_DB=app\nTESTKIT_URL=postgres://${TESTKIT_DB}@db/${TESTKIT_DB}_test\n")

	dir := filepath.Join(root, "app", "core")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	got, err := Lookup(dir, key, "TESTKIT_PASSWORD")
	if err != nil || reflect.DeepEqual(got, map[string]string{
		key: "postgres://app@db/app_test",

		"TESTKIT_PASSWORD": "",
	}) == false {
		t.Errorf("got %v %v", got, err)
	}
}

func TestAMissingFirstKeyIsAnError(t *testing.T) {
	t.Setenv(key, "")

	if _, err := Lookup(module(t, "DB_NAME=app\n"), key, "DB_NAME"); errors.Is(err, errUnset) == false || strings.Contains(err.Error(), key) == false {
		t.Errorf("without the key: %v", err)
	}

	if _, err := Lookup(module(t, ""), key); errors.Is(err, errUnset) == false {
		t.Errorf("empty .env: %v", err)
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module app\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Lookup(root, key); errors.Is(err, errNoDotenv) == false {
		t.Errorf("without .env: %v", err)
	}
}

func TestABrokenDotenvKeepsItsContentOutOfTheError(t *testing.T) {
	t.Setenv(key, "")

	_, err := Lookup(module(t, "TESTKIT_URL=\"postgres://secret\n"), key)
	if errors.Is(err, errBadDotenv) == false || err.Error() != errBadDotenv.Error() {
		t.Errorf("error %v", err)
	}
}

func module(t *testing.T, dotenv string) string {
	t.Helper()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module app\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(dotenv), 0o600); err != nil {
		t.Fatal(err)
	}

	return root
}
