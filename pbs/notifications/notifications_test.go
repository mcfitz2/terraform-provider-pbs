package notifications

import (
	"errors"
	"testing"
)

func TestIsAPINotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		apiPath  string
		expected bool
	}{
		{
			name:     "status 404",
			err:      errors.New("API request failed with status 404: Path '/api2/json/config/notifications/endpoints' not found."),
			apiPath:  "/config/notifications/endpoints",
			expected: true,
		},
		{
			name:     "not found without status",
			err:      errors.New("Path '/api2/json/config/notifications/endpoints' not found"),
			apiPath:  "/config/notifications/endpoints",
			expected: true,
		},
		{
			name:     "different path not matched",
			err:      errors.New("Path '/api2/json/other' not found"),
			apiPath:  "/config/notifications/endpoints",
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			apiPath:  "/config/notifications/endpoints",
			expected: false,
		},
		{
			name:     "unrelated error",
			err:      errors.New("API request failed with status 500: internal error"),
			apiPath:  "/config/notifications/endpoints",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := isAPINotFoundError(tt.err, tt.apiPath); result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
