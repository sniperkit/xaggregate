package librariesio

import (
	"context"
	"strings"

	"github.com/hackebrot/go-librariesio/librariesio"
)

type Service struct {
	client *librariesio.Client
}

func NewClient(token string) (*librariesio.Client, error) {
	if token == "" {
		return nil, errEmptyToken
	}
	client := librariesio.NewClient(strings.TrimSpace(token))
	return client, nil
}

func getProject(owner, name string) (map[string]interface{}, error) {
	project, _, err := client.Project(ctx, owner, name)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	if err != nil {
		log.Errorln("error: ", err.Error())
		return nil, err
	}

	return project, nil
}
