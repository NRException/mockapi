package settings

import (
	"errors"
	"fmt"
	"os"
	co "src/common"

	"gopkg.in/yaml.v3"
)

type UnmarshalledRootSettingWebListenerHTTPSCertFiles struct {
	CertFile string
	KeyFile  string
}
func (re *UnmarshalledRootSettingWebListenerHTTPSCertFiles) ValidateCritical() error {
	err := validateStringLength(re.CertFile, "CertFile")
	if err != nil {return err}
	err = validateStringLength(re.KeyFile, "KeyFile")
	if err != nil {return err}

	return nil
}

type UnmarshalledRootSettingWebListenerResponseHeaders struct {
	HeaderKey   string
	HeaderValue string
}
func (re UnmarshalledRootSettingWebListenerResponseHeaders) ValidateCritical() error {
	err := validateStringLength(re.HeaderKey, "HeaderKey")
	if err != nil {return err}
	err = validateStringLength(re.HeaderValue, "KeaderValue")
	if err != nil {return err}

	return nil
}

type UnmarshalledRootSettingWebListenerContentBodyType string
func (re UnmarshalledRootSettingWebListenerContentBodyType) String() string {return string(re)}
const (
	CONST_RESPONSEBODYTYPE_FILE 	UnmarshalledRootSettingWebListenerContentBodyType = "file"
	CONST_RESPONSEBODYTYPE_STATIC	UnmarshalledRootSettingWebListenerContentBodyType = "static"
	CONST_RESPONSEBODYTYPE_PROXY	UnmarshalledRootSettingWebListenerContentBodyType = "proxy"
)

type UnmarshalledRootSettingWebListenerContentBinding struct {
	BindingPath     string
	ResponseHeaders []UnmarshalledRootSettingWebListenerResponseHeaders
	ResponseCode    int
	ResponseBody    string
	ResponseBodyType UnmarshalledRootSettingWebListenerContentBodyType
}
func (re UnmarshalledRootSettingWebListenerContentBinding) ValidateCritical() error {
	err := validateStringLength(re.ResponseBody, "ResponseBody")
	if err != nil {return err}
	err = validateIntGtZero(re.ResponseCode, "ResponseCode")
	if err != nil {return err}
	err = validateStringLength(re.BindingPath, "BindingPath")
	if err != nil {return err}
	err = validateStringLength(re.ResponseBodyType.String(), "ResponseBodyType")
	if err != nil {return err}

	for _, responseHeaders := range re.ResponseHeaders {
		err := responseHeaders.ValidateCritical()
		if err != nil {
			return err
		}
	}

	return nil
}

type UnmarshalledRootSettingWebListener struct {
	ListenerName       string
	ListenerPort       int
	OnConnectKeepAlive bool
	EnableTLS          bool
	CertDetails        UnmarshalledRootSettingWebListenerHTTPSCertFiles
	ContentBindings    []UnmarshalledRootSettingWebListenerContentBinding
}
func (re *UnmarshalledRootSettingWebListener) ValidateCritical() error {
	err := validateStringLength(re.ListenerName, "ListenerName")
	if err != nil {return err}
	err = validateIntGtZero(re.ListenerPort, "ListenerPort")
	if err != nil {return err}

	err = re.CertDetails.ValidateCritical()
	if err != nil {return fmt.Errorf("UnmarshalledRootSettingWebListener.ValidateCritical: %w", err)}

	for _, contentBindings := range re.ContentBindings {
		err := contentBindings.ValidateCritical()
		if err != nil {
			return fmt.Errorf("UnmarshalledRootSettingWebListener.ValidateCritical: %w", err)
		}
	}

	return nil
}

type UnmarshalledRootSettings struct {
	Id                   string
	Schema               string
	AdditionalProperties bool
	Description          string
	WebListeners         []UnmarshalledRootSettingWebListener
}
func (re *UnmarshalledRootSettings) ValidateCritical() error {

	if len(re.WebListeners) > 0 {
		return errors.New("UnmarshalledRootSettings.ValidateCritical: webListeners field in settings must be present and must contain atleast one web listener array")
	}

	for _, webListenerStruct := range re.WebListeners {
		err := webListenerStruct.ValidateCritical()
		if err != nil {
			return fmt.Errorf("UnmarshalledRootSettings.ValidateCritical: %w", err)
		}
	}

	return nil
}

// Generic type validation funcs / methods
var errorTemplateStringLength string = "%s field in settings file must be present and greater than 0 characters"
var errorTemplateIntPositive string = "%s field in settings file must be present and greater than 0 characters"

func validateStringLength(str string, varname string) error {
	if len(str) <= 0 {
		return fmt.Errorf("validateStringLength: %w", co.TemplateError(errorTemplateIntPositive, varname))
	}
	return nil
}
func validateIntGtZero(i int, varname string) error {
	if i <= 0 {
		return fmt.Errorf("validateIntGtZero: %w", co.TemplateError(errorTemplateIntPositive, varname))
	}
	return nil
}

// Base funcs / methods
func UnmarshalSettingsFile(path string) (*UnmarshalledRootSettings, error) {
	decodedSettings := UnmarshalledRootSettings{}

	// Read file and validate
	b, err := os.ReadFile(path)

	if err != nil {
		return &decodedSettings, err
	}
	if len(b) == 0 {
		return &decodedSettings, errors.New("settings file is empty")
	}

	// Unmarshal and validate
	err = yaml.Unmarshal(b, &decodedSettings)

	if err != nil {
		return &decodedSettings, err
	}

	// Validate struct critical datatypes...
	decodedSettings.ValidateCritical()

	if err != nil {
		return &decodedSettings, err
	}

	return &decodedSettings, nil
}
