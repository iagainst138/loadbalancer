package lb

import "testing"

func TestNewManager(t *testing.T) {
	testCases := []struct {
		name        string
		shouldError bool
		configFile  string
	}{
		{"empty config path", true, ""},
		{"valid config", false, "../sample_configs/config.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewManager(tc.configFile)
			if !tc.shouldError && err != nil {
				t.Fatal(err)
			} else if tc.shouldError && err == nil {
				t.Fatal("expected error but got none")
			}
		})
	}
}
