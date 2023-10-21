package settings

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	co "github.com/nrexception/mockapi/pkg/common"
)

type ResponseHeader struct {
	Key   string `yaml:"headerkey"`
	Value string `yaml:"headervalue"`
}

func (header *ResponseHeader) Validate() error {
	if header.Key == "" {
		return fmt.Errorf("header key must be defined")
	}

	if header.Value == "" {
		return fmt.Errorf("header value must be defined")
	}

	return nil
}

const (
	File   BodyType = "file"
	Inline BodyType = "inline"
	Proxy  BodyType = "proxy"
)

type BodyType string

func (bodyType BodyType) String() string { return string(bodyType) }

type ResponseBinding struct {
	Path             string           `yaml:"bindingpath"`
	ResponseHeaders  []ResponseHeader `yaml:"responseheaders"`
	ResponseCode     int              `yaml:"responsecode"`
	ResponseBody     string           `yaml:"responsebody"`
	ResponseBodyType BodyType         `yaml:"responsebodytype"`
}

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

	if binding.ResponseCode <= 100 {
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
		if strings.HasSuffix(binding.ResponseBody, fileExtension) {
			return true
		}

		return false
	})

	if !isValidFileType {
		return fmt.Errorf("invalid response body file type: %s", binding.ResponseBody)
	}

	return nil
}

type UnmarshalledRootSettingWebListenerHTTPSCertFiles struct {
	CertFile string
	KeyFile  string
}

func (s *UnmarshalledRootSettingWebListenerHTTPSCertFiles) Validate() error {
	if len(s.CertFile) == 0 {
		return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate(): CertFile in settings file must be present!"))
	}
	co.LogVerbose(fmt.Sprintf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate() Evaluating \"%s\"...", s.CertFile), co.MSGTYPE_INFO)
	if len(s.KeyFile) == 0 {
		return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate(): KeyFile in settings file must be present!"))
	}

	// Check to see if files exist and are readable...
	_, err := os.Stat(s.CertFile)
	if err != nil {
		return fmt.Errorf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate(): Cert File does not exist or is not readable: %w", err)
	}
	_, err = os.Stat(s.KeyFile)
	if err != nil {
		return fmt.Errorf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate(): Key File does not exist or is not readable: %w", err)
	}

	return nil
}

type UnmarshalledRootSettingWebListener struct {
	ListenerName       string
	ListenerPort       int
	OnConnectKeepAlive bool
	EnableTLS          bool
	CertDetails        *UnmarshalledRootSettingWebListenerHTTPSCertFiles
	ContentBindings    []ResponseBinding
}

func (s *UnmarshalledRootSettingWebListener) Validate() error {
	if len(s.ListenerName) == 0 {
		return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListener.Validate(): ListenerName in settings file must be present!"))
	}
	co.LogVerbose(fmt.Sprintf("UnmarshalledRootSettingWebListener.Validate() Evaluating \"%s\"...", s.ListenerName), co.MSGTYPE_INFO)

	if s.ListenerPort <= 0 {
		return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListener.Validate(): ListenerPort in settings file must be greater than 0"))
	}

	// Object is "nillable" as it's a ptr reference...
	if s.CertDetails != nil {
		err := s.CertDetails.Validate()
		if err != nil {
			return fmt.Errorf("UnmarshalledRootSettingWebListener.Validate(): %w", err)
		}
	}

	for _, i := range s.ContentBindings {
		err := i.Validate()
		if err != nil {
			return fmt.Errorf("UnmarshalledRootSettingWebListener.Validate(): %w", err)
		}
	}

	return nil
}

type UnmarshalledRootSettings struct {
	Id           string
	Schema       string
	Description  string
	WebListeners []UnmarshalledRootSettingWebListener
}

func (s *UnmarshalledRootSettings) Validate() error {
	if len(s.Id) == 0 {
		return errors.New("UnmarshalledRootSettings.Validate(): Id field in settings file must be present!")
	}
	co.LogVerbose(fmt.Sprintf("UnmarshalledRootSettings.Validate() Evaluating \"%s\"...", s.Id), co.MSGTYPE_INFO)

	if len(s.Schema) == 0 {
		return errors.New("UnmarshalledRootSettings.Validate(): Schema field in settings file must be present!")
	}
	if len(s.Description) == 0 {
		return errors.New("UnmarshalledRootSettings.Validate(): Schema field in settings file must be present!")
	}
	if len(s.WebListeners) < 1 {
		return errors.New("UnmarshalledRootSettings.Validate(): WebListeners definition must be present and must have at least one valid entry!")
	}

	co.LogVerbose("UnmarshalSettingsFile() Validating web listeners...", co.MSGTYPE_INFO)
	for _, i := range s.WebListeners {
		err := i.Validate()
		if err != nil {
			return fmt.Errorf("UnmarshalledRootSettings.Validate(): %w", err)
		}
	}

	return nil
}

// Base funcs / methods
func UnmarshalSettingsFile(path string) (umrs *UnmarshalledRootSettings, err error) {
	co.LogVerbose(fmt.Sprintf("UnmarshalSettingsFile() Unmarshalling settings file \"%s\"", path), co.MSGTYPE_INFO)

	var decodedSettings UnmarshalledRootSettings

	// Read file and validate
	co.LogVerbose("UnmarshalSettingsFile() Reading file data...", co.MSGTYPE_INFO)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("UnmarshalSettingsFile: %s file is read 0 bytes, it is likely empty.", err)
	}

	co.LogVerbose(fmt.Sprintf("UnmarshalSettingsFile() file is %d bytes", len(b)), co.MSGTYPE_INFO)

	// Unmarshal and validate

	co.LogVerbose("UnmarshalSettingsFile() Unmarshalling bytes...", co.MSGTYPE_INFO)
	err = yaml.Unmarshal(b, &decodedSettings)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling file contents: %w", err)
	}

	// Validate struct critical datatypes...
	co.LogVerbose("UnmarshalSettingsFile() Validating data structures...", co.MSGTYPE_INFO)
	err = decodedSettings.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating yaml file: %w", err)
	}

	co.LogVerbose("UnmarshalSettingsFile() All data structures valid!", co.MSGTYPE_INFO)

	return &decodedSettings, nil
}
