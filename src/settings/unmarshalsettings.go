package settings

import (
	"errors"
	"fmt"
	"log"
	co "mockapi/src/common"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type UnmarshalledRootSettingWebListenerResponseHeaders struct {
	HeaderKey   string
	HeaderValue string
}
func (s *UnmarshalledRootSettingWebListenerResponseHeaders) Validate() error {
	if len(s.HeaderKey) == 0 {return errors.New("UnmarshalledRootSettingWebListenerResponseHeaders.Validate(): HeaderKey in settings file must be present!")}
	co.LogVerbose(fmt.Sprintf("UnmarshalledRootSettingWebListenerResponseHeaders.Validate() Evaluating \"%s\"...", s.HeaderKey), co.MSGTYPE_INFO)
	if len(s.HeaderValue) == 0 {return errors.New("UnmarshalledRootSettingWebListenerResponseHeaders.Validate(): HeaderValue in settings file must be present!")}
	return nil
}

type UnmarshalledRootSettingWebListenerContentBodyType string
func (re UnmarshalledRootSettingWebListenerContentBodyType) String() string {return string(re)}
const (
	CONST_RESPONSEBODYTYPE_FILE 	UnmarshalledRootSettingWebListenerContentBodyType = "file"
	CONST_RESPONSEBODYTYPE_INLINE	UnmarshalledRootSettingWebListenerContentBodyType = "inline"
	CONST_RESPONSEBODYTYPE_PROXY	UnmarshalledRootSettingWebListenerContentBodyType = "proxy"
)


type UnmarshalledRootSettingWebListenerContentBinding struct {
	BindingPath     string
	ResponseHeaders []UnmarshalledRootSettingWebListenerResponseHeaders
	ResponseCode    int
	ResponseBody    string
	ResponseBodyType UnmarshalledRootSettingWebListenerContentBodyType
}
func (s *UnmarshalledRootSettingWebListenerContentBinding) Validate() error {
	allowedResponseBodyTypes := []string{"inline", "file", "proxy"}
	allowedFileTypes := []string{".json", ".txt", ".csv", ".html"}

	if len(s.BindingPath) == 0 {return errors.New("UnmarshalledRootSettingWebListenerContentBinding.Validate(): BindingPath in settings file must be present!")}
	co.LogVerbose(fmt.Sprintf("UnmarshalledRootSettingWebListenerContentBinding.Validate() Evaluating \"%s\"...", s.BindingPath), co.MSGTYPE_INFO)

	// We might not want any headers...
	if len(s.ResponseHeaders) > 0 {
		for _, i:= range s.ResponseHeaders {
			err := i.Validate()
			if err != nil {return err}
		}
	}

	if s.ResponseCode <= 0 {return errors.New("UnmarshalledRootSettingWebListenerContentBinding.Validate(): Response code in settings file must be greater than 0")}
	if len(s.ResponseBody) == 0 {return errors.New("UnmarshalledRootSettingWebListenerContentBinding.Validate(): ResponseBody in settings file must be present!")}

	// Check response body type matches allowed types defined above...
	finds := 0
	for _, i := range allowedResponseBodyTypes {
		if s.ResponseBodyType.String() == i {finds++}
		if finds > 0 {break}
	}
	if finds == 0 {
		return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListenerContentBinding.Validate(): ResponseBodyType in settings file must be of the following supported values %s", allowedResponseBodyTypes))
	}

	if finds > 0 && s.ResponseBodyType.String() != "file" {return nil}

	// And check the response body
	finds = 0
	for _, i := range allowedFileTypes {
		if strings.HasSuffix(s.ResponseBody, i) {finds++}
		if finds > 0 {break}
	}
	if finds == 0 {
		return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListenerContentBinding.Validate(): ResponseBody in settings file must be of the following supported values %s OR a static value when conjoined with responsebodytype: inline", allowedFileTypes))
	}
	
	return nil
}

type UnmarshalledRootSettingWebListenerHTTPSCertFiles struct {
	CertFile string
	KeyFile  string
}
func (s *UnmarshalledRootSettingWebListenerHTTPSCertFiles) Validate() error {
	if len(s.CertFile) == 0 {return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate(): CertFile in settings file must be present!"))}
	co.LogVerbose(fmt.Sprintf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate() Evaluating \"%s\"...", s.CertFile), co.MSGTYPE_INFO)
	if len(s.KeyFile) == 0 {return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate(): KeyFile in settings file must be present!"))}
	
	// Check to see if files exist and are readable...
	_, err := os.Stat(s.CertFile)
	if err != nil {return fmt.Errorf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate(): Cert File does not exist or is not readable: %w", err)}
	_, err = os.Stat(s.KeyFile)
	if err != nil {return fmt.Errorf("UnmarshalledRootSettingWebListenerHTTPSCertFiles.Validate(): Key File does not exist or is not readable: %w", err)}

	return nil
}


type UnmarshalledRootSettingWebListener struct {
	ListenerName       string
	ListenerPort       int
	OnConnectKeepAlive bool
	EnableTLS          bool
	CertDetails        *UnmarshalledRootSettingWebListenerHTTPSCertFiles
	ContentBindings    []UnmarshalledRootSettingWebListenerContentBinding
}
func (s *UnmarshalledRootSettingWebListener) Validate() error {
	if len(s.ListenerName) == 0 {return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListener.Validate(): ListenerName in settings file must be present!"))}
	co.LogVerbose(fmt.Sprintf("UnmarshalledRootSettingWebListener.Validate() Evaluating \"%s\"...", s.ListenerName), co.MSGTYPE_INFO)

	if s.ListenerPort <= 0 {return errors.New(fmt.Sprintf("UnmarshalledRootSettingWebListener.Validate(): ListenerPort in settings file must be greater than 0"))}

	// Object is "nillable" as it's a ptr reference...
	if s.CertDetails != nil {
		err := s.CertDetails.Validate()
		if err != nil {return fmt.Errorf("UnmarshalledRootSettingWebListener.Validate(): %w", err)}
	}

	for _, i := range s.ContentBindings {
		err := i.Validate()
		if err != nil {return fmt.Errorf("UnmarshalledRootSettingWebListener.Validate(): %w", err)}
	}

	return nil
}


type UnmarshalledRootSettings struct {
	Id                   string
	Schema               string
	Description          string
	WebListeners         []UnmarshalledRootSettingWebListener
}
func (s *UnmarshalledRootSettings) Validate() error {
	if len(s.Id) == 0 {return errors.New("UnmarshalledRootSettings.Validate(): Id field in settings file must be present!")}
	co.LogVerbose(fmt.Sprintf("UnmarshalledRootSettings.Validate() Evaluating \"%s\"...", s.Id), co.MSGTYPE_INFO)

	if len(s.Schema) == 0 {return errors.New("UnmarshalledRootSettings.Validate(): Schema field in settings file must be present!")}
	if len(s.Description) == 0 {return errors.New("UnmarshalledRootSettings.Validate(): Schema field in settings file must be present!")}
	if len(s.WebListeners) < 1 {return errors.New("UnmarshalledRootSettings.Validate(): WebListeners definition must be present and must have at least one valid entry!")}
	
	co.LogVerbose("UnmarshalSettingsFile() Validating web listeners...", co.MSGTYPE_INFO)
	for _, i:=range s.WebListeners {
		err := i.Validate()
		if err != nil {return fmt.Errorf("UnmarshalledRootSettings.Validate(): %w", err)}
	}

	return nil
}


// Base funcs / methods
func UnmarshalSettingsFile(path string) (*UnmarshalledRootSettings, error) {
	co.LogVerbose(fmt.Sprintf("UnmarshalSettingsFile() Unmarshalling settings file \"%s\"", path), co.MSGTYPE_INFO)
	
	var decodedSettings UnmarshalledRootSettings

	// Read file and validate
	b, err := os.ReadFile(path)
	if err != nil {return nil, err}
	if len(b) == 0 {log.Fatalf("UnmarshalSettingsFile: %s file is read 0 bytes, it is likely empty.", err)}

	// Unmarshal and validate
	err = yaml.Unmarshal(b, &decodedSettings)
	if err != nil {return nil, err}

	// Validate struct critical datatypes...
	co.LogVerbose("UnmarshalSettingsFile() Validating data structures...", co.MSGTYPE_INFO)
	err = decodedSettings.Validate()
	if err != nil {return nil, err}
	
	if err != nil {
		return nil, err
	} else {
		co.LogVerbose("UnmarshalSettingsFile() All data structures valid!", co.MSGTYPE_INFO)
	}


	return &decodedSettings, nil
}
