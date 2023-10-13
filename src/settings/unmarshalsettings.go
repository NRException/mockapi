package settings

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type UnmarshalledRootSettings struct {
	Id                   string
	Schema               string
	AdditionalProperties bool
	Description          string
	WebListeners         []UnmarshalledRootSettingWebListener
}

type UnmarshalledRootSettingWebListenerContentHeaders struct {
	HeaderKey   string
	HeaderValue string
}

type UnmarshalledRootSettingWebListener struct {
	ListenerName       string
	ListenerPort       int
	OnConnectKeepAlive bool
	ContentHeaders     []UnmarshalledRootSettingWebListenerContentHeaders
	ContentResponse    int
	ContentBody        string
}

func UnmarshalSettingsFile(path string) (*UnmarshalledRootSettings, error) {
	decodedSettings := UnmarshalledRootSettings{}

	b, err := os.ReadFile(path)

	if err != nil {
		return &decodedSettings, err
	}
	if len(b) <= 0 {
		return &decodedSettings, errors.New("settings file is empty")
	}

	err = yaml.Unmarshal(b, &decodedSettings)

	if err != nil {
		return &decodedSettings, err
	}

	return &decodedSettings, nil
}
