package config

import (
	"os"
	"os/exec"
	"testing"
)

// setupEnv sets an environment variable and returns a function to restore it
func setupEnv(t *testing.T, key, value string) func() {
	t.Helper()
	prevValue, exists := os.LookupEnv(key)
	os.Setenv(key, value)
	return func() {
		if exists {
			os.Setenv(key, prevValue)
		} else {
			os.Unsetenv(key)
		}
	}
}

func TestMustLoad(t *testing.T) {
	t.Run("Load valid config", func(t *testing.T) {
		teardown := setupEnv(t, "CONFIG_PATH", "testdata/testconfig.yaml")
		defer teardown()

		got := MustLoad()
		if got == nil {
			t.Fatal("MustLoad() returned nil")
		}
		if got.Env != "dev" {
			t.Errorf("Expected Env 'development', got '%s'", got.Env)
		}
	})

	t.Run("Missing config file should exit", func(t *testing.T) {
		teardown := setupEnv(t, "CONFIG_PATH", "nonexistent.yaml")
		defer teardown()

		// Capture os.Exit to prevent test from terminating
		if os.Getenv("TEST_EXIT") == "1" {
			MustLoad()
			return
		}

		cmd := exec.Command(os.Args[0], "-test.run=TestMustLoad")
		cmd.Env = append(os.Environ(), "TEST_EXIT=1")
		err := cmd.Run()

		if exitError, ok := err.(*exec.ExitError); ok && !exitError.Success() {
			// Test passes because MustLoad should exit with an error
			return
		}
		t.Fatalf("MustLoad() did not exit as expected")
	})
}
