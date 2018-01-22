package librariesio

import (
	"context"
	"time"
)

var (
	requestTimeout = time.Duration(time.Second * 10)
	requestContext = context.Context
)
