package common

import (
	"errors"
	"fmt"
)

func TemplateError(str string, str2 string) error {
	return errors.New(fmt.Sprintf(str, str2))
}