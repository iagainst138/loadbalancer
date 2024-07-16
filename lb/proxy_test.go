package lb

import "testing"

func TestNewProxy(t *testing.T) {
	testCases := []struct {
		name        string
		shouldError bool
		entry       *Entry
	}{
		{"nil entry", true, nil},
		{"bad backend", true, &Entry{Backend: "__bad__"}},
		{"good backend", false, &Entry{Backend: "RoundRobin"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewProxy(tc.entry)
			if !tc.shouldError && err != nil {
				t.Fatal(err)
			} else if tc.shouldError && err == nil {
				t.Fatal("expected error but got none")
			}
		})
	}

}
