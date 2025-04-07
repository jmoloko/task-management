package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTask(t *testing.T) {
	env, cleanup := SetupTestEnv(t)
	defer cleanup()

	_, token := createTestUser(t, env)
	tasks := createTestTasks(t, env, token, 3)

	t.Run("Get Tasks", func(t *testing.T) {
		resp, err := makeRequest(env, "GET", "/api/tasks", nil, token)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Get Task", func(t *testing.T) {
		resp, err := makeRequest(env, "GET", "/api/tasks/"+tasks[0].ID, nil, token)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
