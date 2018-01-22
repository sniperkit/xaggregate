package github

import (
	"time"
)

func (g *Github) notifyAttempts(err error, i time.Duration) {
	log.Println(err.Error())
}
