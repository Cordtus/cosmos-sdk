package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
)

func TestCheckConfigFiles(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Test 1: No changes
	t.Run("No changes", func(t *testing.T) {
		setupTestConfig(t, configDir, map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		})

		err := CheckAndUpdateFile(configDir, "test.toml", map[string]interface{}{
			"param1": "different_value",
			"param2": 100,
		}, true)
		require.NoError(t, err)

		// Ensure no update is performed for value changes
		content, err := os.ReadFile(filepath.Join(configDir, "test.toml"))
		require.NoError(t, err)
		require.Contains(t, string(content), "value1")
		require.Contains(t, string(content), "42")
	})

	// Test 2: Added parameter
	t.Run("Added parameter", func(t *testing.T) {
		setupTestConfig(t, configDir, map[string]interface{}{
			"param1": "value1",
		})

		err := CheckAndUpdateFile(configDir, "test.toml", map[string]interface{}{
			"param1": "value1",
			"param2": "new_value",
		}, true)
		require.NoError(t, err)

		// Check if update is performed for added parameter
		content, err := os.ReadFile(filepath.Join(configDir, "test.toml"))
		require.NoError(t, err)
		require.Contains(t, string(content), "param2")
		require.Contains(t, string(content), "new_value")
	})

	// Test 3: Removed parameter
	t.Run("Removed parameter", func(t *testing.T) {
		setupTestConfig(t, configDir, map[string]interface{}{
			"param1": "value1",
			"param2": "value2",
		})

		err := CheckAndUpdateFile(configDir, "test.toml", map[string]interface{}{
			"param1": "value1",
		}, true)
		require.NoError(t, err)

		// Check if update is performed for removed parameter
		content, err := os.ReadFile(filepath.Join(configDir, "test.toml"))
		require.NoError(t, err)
		require.NotContains(t, string(content), "param2")
		require.NotContains(t, string(content), "value2")
	})

	// Test 4: Nested structure changes
	t.Run("Nested structure changes", func(t *testing.T) {
		setupTestConfig(t, configDir, map[string]interface{}{
			"parent": map[string]interface{}{
				"child1": "value1",
			},
		})

		err := CheckAndUpdateFile(configDir, "test.toml", map[string]interface{}{
			"parent": map[string]interface{}{
				"child1": "value1",
				"child2": "new_value",
			},
		}, true)
		require.NoError(t, err)

		// Check if update is performed for nested structure changes
		content, err := os.ReadFile(filepath.Join(configDir, "test.toml"))
		require.NoError(t, err)
		require.Contains(t, string(content), "child2")
		require.Contains(t, string(content), "new_value")
	})
}

func setupTestConfig(t *testing.T, configDir string, config map[string]interface{}) {
	content, err := toml.Marshal(config)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(configDir, "test.toml"), content, 0644)
	require.NoError(t, err)
}
