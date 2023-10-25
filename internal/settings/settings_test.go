package settings_test

import (
	"net/http"
	"testing"

	"github.com/nrexception/mockapi/internal/settings"
)

func TestResponseHeaders_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		key           string
		value         string
		expectedError bool
	}{
		{
			name:          "no key or value",
			key:           "",
			value:         "",
			expectedError: true,
		},
		{
			name:          "no key",
			key:           "",
			value:         "value",
			expectedError: true,
		},
		{
			name:          "no value",
			key:           "key",
			value:         "",
			expectedError: true,
		},
		{
			name:          "key and value",
			key:           "key",
			value:         "value",
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header := &settings.ResponseHeader{
				Key:   tc.key,
				Value: tc.value,
			}

			err := header.Validate()
			if (err != nil) != tc.expectedError {
				t.Errorf("unexpected error response: %v", err)
			}
		})
	}
}

func TestBodyType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		bodyType settings.BodyType
	}{
		{
			name:     "file",
			bodyType: settings.File,
		},
		{
			name:     "inline",
			bodyType: settings.Inline,
		},
		{
			name:     "proxy",
			bodyType: settings.Proxy,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.bodyType.String()
			want := string(tc.bodyType)

			if got != want {
				t.Errorf("unexpected string representation:\ngot: %s\nwant:%s", got, want)
			}
		})
	}
}

func TestResponseBinding_Validate(t *testing.T) {
	t.Parallel()

	// TODO: Add more test cases
	testCases := []struct {
		name             string
		path             string
		responseHeaders  []settings.ResponseHeader
		responseCode     int
		responseBody     string
		responseBodyType settings.BodyType
		expectedError    bool
	}{
		{
			name:             "file",
			path:             "/",
			responseHeaders:  []settings.ResponseHeader{{Key: "key", Value: "value"}},
			responseCode:     http.StatusOK,
			responseBody:     "somefile.json",
			responseBodyType: settings.File,
			expectedError:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			binding := &settings.ResponseBinding{
				Path:             tc.path,
				ResponseHeaders:  tc.responseHeaders,
				ResponseCode:     tc.responseCode,
				ResponseBody:     tc.responseBody,
				ResponseBodyType: tc.responseBodyType,
			}

			err := binding.Validate()
			if (err != nil) != tc.expectedError {
				t.Errorf("unexpected error response: %v", err)
			}
		})
	}
}
