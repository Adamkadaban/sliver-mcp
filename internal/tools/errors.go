package tools

import (
	"fmt"
)

func NewInvalidArgsError(message string) error {
	return fmt.Errorf("invalid arguments: %s", message)
}
