package lb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetResource(t *testing.T) {
	testCases := []struct {
		name        string
		shouldExist bool
		devMode     bool
		path        string
	}{
		{"empty path", false, false, ""},
		{"valid path", true, false, "resources/index.html"},
		{"valid path devmode", true, true, "resources/index.html"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				DevMode = false
			})

			DevMode = tc.devMode

			_, err := GetResource(tc.path)
			if tc.shouldExist && err != nil {
				t.Fatal(err)
			} else if !tc.shouldExist && err == nil {
				t.Fatal("expected error but got none")
			}
		})
	}
}

func TestServeResource(t *testing.T) {
	testCases := []struct {
		name           string
		expectedStatus int
		devMode        bool
		path           string
	}{
		{"valid path", http.StatusOK, false, "resources/index.html"},
		{"valid path devmode", http.StatusOK, true, "resources/index.html"},
		{"invalid path", http.StatusNotFound, false, "resources/does_not_exist.html"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				DevMode = false
			})

			DevMode = tc.devMode

			req := httptest.NewRequest("GET", fmt.Sprintf("http://127.0.0.1/%s", tc.path), nil)
			w := httptest.NewRecorder()
			ServeResource(w, req)

			resp := w.Result()

			if resp.StatusCode != tc.expectedStatus {
				t.Fatalf("expected status %d but got status %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}
