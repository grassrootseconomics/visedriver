// +build configreport

package config

import (
	"fmt"
	"os"
	"testing"

	"git.defalsify.org/vise.git/logging"
)

// go test -tags configreport ./config/...   ---> run with tag
func TestConfig(t *testing.T) {
	logger := logging.NewVanilla().WithDomain("test")
	cfg := NewConfig(logger)

	t.Run("Default Values", func(t *testing.T) {
		cfg.AddKey("TEST_KEY", "default", false, nil)
		value, err := cfg.GetValue("TEST_KEY")
		t.Logf("Got value: %q, error: %v", value, err)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if value != "default" {
			t.Errorf("expected 'default', got '%s'", value)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("TEST_ENV_KEY", "override")
		defer os.Unsetenv("TEST_ENV_KEY")

		cfg.AddKey("TEST_ENV_KEY", "default", false, nil)
		value, err := cfg.GetValue("TEST_ENV_KEY")
		t.Logf("Got value: %q, error: %v", value, err)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if value != "override" {
			t.Errorf("expected 'override', got '%s'", value)
		}
	})

	t.Run("Validation", func(t *testing.T) {
		validator := func(v string) error {
			if v != "valid" {
				return fmt.Errorf("invalid value")
			}
			return nil
		}

		cfg.AddKey("VALIDATED_KEY", "valid", false, validator)
		os.Setenv("VALIDATED_KEY", "invalid")
		defer os.Unsetenv("VALIDATED_KEY")

		value, err := cfg.GetValue("VALIDATED_KEY")
		t.Logf("Got value: %q, error: %v", value, err)
		if err == nil {
			t.Error("expected validation error, got nil")
		}
	})
}
