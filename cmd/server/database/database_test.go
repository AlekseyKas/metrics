package database

import (
	"testing"

	"github.com/AlekseyKas/metrics/internal/helpers"
	"github.com/stretchr/testify/require"
)

func TestDBConnect(t *testing.T) {
	DBURL, id, _ := helpers.StartDB()
	t.Run("Test DB connect", func(t *testing.T) {
		err := DBConnect(DBURL)
		require.NoError(t, err)
	})
	helpers.StopDB(id)
}

func TestDBClose(t *testing.T) {
	err := DBClose()
	require.NoError(t, err)
}
