package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDBConnect(t *testing.T) {
	DBURL := "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"

	t.Run("Test DB connect", func(t *testing.T) {
		err := DBConnect(DBURL)
		require.NoError(t, err)
	})
}

func TestDBClose(t *testing.T) {
	err := DBClose()
	require.NoError(t, err)
}
