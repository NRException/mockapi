// Package settings provides functionality for parsing and validating settings.
package settings

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/nrexception/mockapi/internal/logging"
)

// ResponseHeader is a key/value struct for a response header.
type ResponseHeader struct {
	Key   string `yaml:"headerkey"`
	Value string `yaml:"headervalue"`
}

// Validate checks that the ResponseHeader is defined correctly.
func (header *ResponseHeader) Validate() error {
	if header.Key == "" {
		return fmt.Errorf("header key must be defined")
	}

	if header.Value == "" {
		return fmt.Errorf("header value must be defined")
	}

	return nil
}

// Response body types.
const (
	File   BodyType = "file"
	Inline BodyType = "inline"
	Proxy  BodyType = "proxy"
)

// BodyType is a type alias for type safety.
type BodyType string

// String returns the string representation of the BodyType.
func (bodyType BodyType) String() string {
	return string(bodyType)
}

// ResponseBinding is the binding of a response to a URL path.
type ResponseBinding struct {
	Path             string           `yaml:"bindingpath"`
	ResponseHeaders  []ResponseHeader `yaml:"responseheaders"`
	ResponseCode     int              `yaml:"responsecode"`
	ResponseBody     string           `yaml:"responsebody"`
	ResponseBodyType BodyType         `yaml:"responsebodytype"`
}

// Validate checks that the ResponseBinding is defined correctly.
func (binding *ResponseBinding) Validate() error {
	allowedResponseBodyTypes := []BodyType{File, Inline, Proxy}
	allowedFileTypes := []string{".json", ".txt", ".csv", ".html", ".xml"}

	if binding.Path == "" {
		return fmt.Errorf("binding path must be defined")
	}

	// We might not want any headers...
	if len(binding.ResponseHeaders) > 0 {
		for _, i := range binding.ResponseHeaders {
			err := i.Validate()
			if err != nil {
				return err
			}
		}
	}

	if binding.ResponseCode < http.StatusContinue || binding.ResponseCode > http.StatusNetworkAuthenticationRequired {
		return fmt.Errorf("invalid response code: %d", binding.ResponseCode)
	}

	// TODO: response body could be empty
	if binding.ResponseBody == "" {
		return fmt.Errorf("invalid response body: %s", binding.ResponseBody)
	}

	if !slices.Contains(allowedResponseBodyTypes, binding.ResponseBodyType) {
		return fmt.Errorf("invalid response body type: %s", binding.ResponseBodyType)
	}

	if binding.ResponseBodyType != File {
		return nil
	}

	isValidFileType := slices.ContainsFunc(allowedFileTypes, func(fileExtension string) bool {
		return strings.HasSuffix(binding.ResponseBody, fileExtension)
	})

	if !isValidFileType {
		return fmt.Errorf("invalid response body file type: %s", binding.ResponseBody)
	}

	return nil
}

// UnmarshalledRootSettingWebListenerHTTPSCertFiles is a certificate
// public/private key pair.
type UnmarshalledRootSettingWebListenerHTTPSCertFiles struct {
	CertFile string
	KeyFile  string
}

// Validate checks that the UnmarshalledRootSettingWebListenerHTTPSCertFiles is
// defined correctly.
func (s *UnmarshalledRootSettingWebListenerHTTPSCertFiles) Validate() error {
	if s.CertFile == "" {
		return fmt.Errorf("cert file is not defined")
	}

	if s.KeyFile == "" {
		return fmt.Errorf("key file is not defined")
	}

	_, err := os.Stat(s.CertFile)
	if err != nil {
		return fmt.Errorf("error checking cert file: %w", err)
	}

	_, err = os.Stat(s.KeyFile)
	if err != nil {
		return fmt.Errorf("error checking key file: %w", err)
	}

	return nil
}

// UnmarshalledRootSettingWebListener is an HTTP listener with response
// bindings.
type UnmarshalledRootSettingWebListener struct {
	ListenerName       string
	ListenerPort       int
	OnConnectKeepAlive bool
	EnableTLS          bool
	CertDetails        *UnmarshalledRootSettingWebListenerHTTPSCertFiles
	ContentBindings    []*ResponseBinding
}

// Validate checks that the UnmarshalledRootSettingWebListener is defined
// correctly.
func (s *UnmarshalledRootSettingWebListener) Validate() error {
	if s.ListenerName == "" {
		return fmt.Errorf("listener name must be defined")
	}

	logging.LogVerbose(logging.Info, fmt.Sprintf("UnmarshalledRootSettingWebListener.Validate() Evaluating %s...", s.ListenerName))

	if s.ListenerPort <= 0 {
		return fmt.Errorf("listener port must be greater than 0")
	}

	// Object is "nillable" as it's a ptr reference...
	if s.CertDetails != nil {
		err := s.CertDetails.Validate()
		if err != nil {
			return fmt.Errorf("error validating certificate files: %w", err)
		}
	}

	for _, i := range s.ContentBindings {
		err := i.Validate()
		if err != nil {
			return fmt.Errorf("error validating content bindings: %w", err)
		}
	}

	return nil
}

// UnmarshalledRootSettings is a collection of
// UnmarshalledRootSettingWebListener objects and metadata.
type UnmarshalledRootSettings struct {
	ID           string                               `yaml:"id"`
	Schema       string                               `yaml:"schema"`
	Description  string                               `yaml:"description"`
	WebListeners []UnmarshalledRootSettingWebListener `yaml:"weblisteners"`
}

// Validate checks that the UnmarshalledRootSettings is defined correctly.
func (s *UnmarshalledRootSettings) Validate() error {
	switch {
	case s.ID == "":
		return fmt.Errorf("field id in settings file must be defined")
	case s.Schema == "":
		return fmt.Errorf("field schema in settings file must be defined")
	case s.Description == "":
		return fmt.Errorf("field description in settings file must be defined")
	case len(s.WebListeners) == 0:
		return fmt.Errorf("field weblisteners in settings file must have at least one valid entry")
	}

	for _, i := range s.WebListeners {
		err := i.Validate()
		if err != nil {
			return fmt.Errorf("error validating weblistener definition: %w", err)
		}
	}

	return nil
}

// UnmarshalSettingsFile unmarshals the settings file and returns a
// UnmarshalledRootSettings.
func UnmarshalSettingsFile(path string) (*UnmarshalledRootSettings, error) {
	logging.LogVerbose(logging.Info, fmt.Sprintf("UnmarshalSettingsFile() Unmarshalling settings file %s", path))

	var decodedSettings UnmarshalledRootSettings

	// Read file and validate
	logging.LogVerbose(logging.Info, "UnmarshalSettingsFile() Reading file data...")

	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if len(b) == 0 {
		return nil, fmt.Errorf("config file %s is empty", path)
	}

	logging.LogVerbose(logging.Info, fmt.Sprintf("UnmarshalSettingsFile() file is %d bytes", len(b)))

	// Unmarshal and validate

	logging.LogVerbose(logging.Info, "UnmarshalSettingsFile() Unmarshalling bytes...")

	err = yaml.Unmarshal(b, &decodedSettings)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling file contents: %w", err)
	}

	// Validate struct critical datatypes...
	logging.LogVerbose(logging.Info, "UnmarshalSettingsFile() Validating data structures...")

	err = decodedSettings.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating yaml file: %w", err)
	}

	logging.LogVerbose(logging.Info, "UnmarshalSettingsFile() All data structures valid!")

	return &decodedSettings, nil
}
