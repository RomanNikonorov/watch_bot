package lib

import (
	"os"
	"testing"
)

func TestGetEnvVariableValueWithDefault(t *testing.T) {
	tests := []struct {
		envVariableName string
		envValue        string
		defaultValue    string
		expected        int
	}{
		{"TEST_ENV_VAR", "42", "10", 42},
		{"TEST_ENV_VAR", "", "10", 10},
		//{"TEST_ENV_VAR", "invalid", "10", 10}, // This will cause a log.Fatal, so handle it accordingly
	}

	for _, tt := range tests {
		if tt.envValue != "" {
			os.Setenv(tt.envVariableName, tt.envValue)
		} else {
			os.Unsetenv(tt.envVariableName)
		}

		result := GetEnvVariableValueWithDefault(tt.envVariableName, tt.defaultValue)
		if result != tt.expected {
			t.Errorf("GetEnvVariableValueWithDefault(%s, %s) = %d; want %d", tt.envVariableName, tt.defaultValue, result, tt.expected)
		}
	}
}
