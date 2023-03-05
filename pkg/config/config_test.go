package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {

	type cfg struct {
		DB struct {
			User string `toml:"user"`
			Pass string `toml:"pass"`
		} `toml:"db"`
	}

	raw := `
		[db]
		user = "ing_user"
		pass = "{% ENV:ING_E2E_TEST_PASS %}"
	`

	t.Setenv("ING_E2E_TEST_PASS", "ing_pass")

	filename, tearDown := setUp(t, raw)
	defer tearDown()

	var got cfg

	err := ParseFile(filename, &got)
	assert.NoError(t, err)

	want := cfg{}
	want.DB.User = "ing_user"
	want.DB.Pass = "ing_pass"

	assert.Equal(t, want, got)
}

func setUp(t *testing.T, raw string) (string, func()) {
	t.Helper()

	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	if _, err := f.WriteString(raw); err != nil {
		t.Fatalf("could not write test config to temp file %v: %v", f.Name(), err)
	}
	if err := f.Sync(); err != nil {
		t.Fatalf("could not flush test config file %v: %v", f.Name(), err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("could not close test config file %v: %v", f.Name(), err)
	}

	return f.Name(), func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Fatalf("could not cleanup test config file %v: %v", f.Name(), err)
		}
	}
}
