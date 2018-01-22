package librariesio

import (
	"errors"
)

var (
	errEmptyToken = errors.New("Empty libraries.io token provided, cannot instantiate api client...")
)
