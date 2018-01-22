package service

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/hoop33/entrevista"

	"github.com/sniperkit/xlogger/pkg"
	// "github.com/sniperkit/xtask/test/service/provider"
)

var (
	logx     logger.Logger
	services          = make(map[string]Service)
	skiplist []string = []string{}
)

// Service represents a service
type Service interface {
	Login(ctx context.Context) error
	Cache(ctx context.Context, provider interface{}) error
}

func registerService(service Service) {
	services[Name(service)] = service
}

// Name returns the name of a service
func Name(service Service) string {
	parts := strings.Split(reflect.TypeOf(service).String(), ".")
	return strings.ToLower(parts[len(parts)-1])
}

// ForName returns the service for a given name, or an error if it doesn't exist
func ForName(name string) (Service, error) {
	if service, ok := services[strings.ToLower(name)]; ok {
		return service, nil
	}
	return &NotFound{}, fmt.Errorf("Service '%s' not found", name)
}

func createInterview() *entrevista.Interview {
	interview := entrevista.NewInterview()
	interview.ShowOutput = func(message string) {
		fmt.Print(color.GreenString(message))
	}
	interview.ShowError = func(message string) {
		color.Red(message)
	}
	return interview
}

/*
func init() {
	registerService(&Github{})
}

// ref. https://github.com/docker/libkv/blob/master/libkv.go

// Initialize creates a new Provider object, initializing the client
type Initialize func(addrs []string, options *provider.Config) (provider.Provider, error)

var (
	// Provider initializers
	initializers = make(map[provider.Provider]Initialize)
	supportedProvider = func() string {
		keys := make([]string, 0, len(initializers))
		for k := range initializers {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)
		return strings.Join(keys, ", ")
	}()
)

// NewService creates an instance of store
func NewService(backend provider.Provider, addrs []string, options *provider.Config) (provider.Provider, error) {
	if init, exists := initializers[backend]; exists {
		return init(addrs, options)
	}
	return nil, fmt.Errorf("%s %s", store.ErrBackendNotSupported.Error(), supportedProvider)
}

// AddService adds a new provider backend to service
func AddService(p provider.Provider, init Initialize) {
	initializers[p] = init
}
*/
