package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	c := Config{
		Addr:   "localhost:5432",
		User:   "user",
		Pass:   "pass",
		DBName: "dbname",
	}

	require.Equal(t, "postgres://user:pass@localhost:5432/dbname", c.DSN())
}
