package service

import (
	"log"
	"sync"
	"time"
)

type Token struct {
	lock sync.RWMutex

	Disabled bool          `yaml:"disabled,omitempty" toml:"disabled,omitempty" json:"disabled,omitempty" mapstructure:"disabled,omitempty" default:"false"`
	Key      string        `yaml:"key" json:"key" toml:"key" mapstructure:"key"`
	sleep    time.Duration `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
	reset    time.Time     `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
	status   TokenStatus   `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
}

func (t *Token) Ready() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	wait := time.Until(t.reset).Nanoseconds()
	log.Println("t.Key", t.Key, "sleep: ", t.sleep.Seconds(), "s, reset: ", t.reset.Format(time.UnixDate), ", wait: ", wait, ", ready? ", wait <= 0)
	return wait <= 0
}

func (t *Token) GetSleep() time.Duration {
	// t.lock.RLock()
	return t.sleep
	// t.lock.RUnlock()
}

func (t *Token) SetSleep(sleep time.Duration) {
	// t.lock.Lock()
	t.sleep = sleep
	// t.lock.Unlock()
}

func (t *Token) GetReset() time.Time {
	// t.lock.RLock()
	return t.reset
	// t.lock.RUnlock()
}

func (t *Token) SetReset(reset time.Time) {
	// t.lock.Lock()
	t.reset = reset
	// t.lock.Unlock()
}

type TokenStatus struct {
	Paused    bool      `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
	Rate      int       `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
	Remaining int       `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
	CreatedAt time.Time `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
	ResetAt   time.Time `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
	Err       error     `yaml:"-" json:"-" toml:"-" mapstructure:"-"`
}

type tokenRegistry map[string]*TokenProfile

// A rateLimitError is returned when the requestor's rate limit has been exceeded.
type TokenProfile struct {
	Enabled  bool   // enable ratelimit defined rules
	Debug    bool   // debug ratelimit processing
	Verbose  bool   // verbose ratelimit details
	Provider string // service name
	Next     bool   // Use next token if available
	Wait     bool   // sleep the tasks until the next rateLimit reset

	// Config *TokenLimitConfig  // behaviour settings
	Status *TokenLimitContext // rate limit status
	// Stats  *tokenLimitStats   // stats about remaing jobs/flows

	resetAtUnix *int64 // Unix seconds at which rate limit is reset
}

type TokenConfig struct {
	Token struct {
		Next bool // Use next token if available
		Wait bool // sleep the tasks until the next rateLimit reset
	}
	Notify struct {
		Runner bool // notify runner about rate limit block
	}
	Request struct {
		ForceConditional bool // force conditional requests (enabled httpcache automatically)
	}
	Dump struct {
		Enabled bool
	}
}

type TokenLimitContext struct {
	Locked    bool
	CreatedAt *time.Time // ratelimit block triggered at
	Caller    struct {
		Task        *string // ratelimit block triggered by function/method
		Job         *string // ratelimit block triggered by function/method
		FuncName    *string // ratelimit block triggered by function/method
		RequestInfo struct {
			Url *string // requested API endpoint URL
			Uri *string // requested API endpoint URI
		}
		ResponseInfo struct {
			Code   *int
			Cached *bool
		}
	}
}

func registerToken(provider string, token string) (tokenRegistry map[string]*TokenProfile) {
	if provider == "" {
		provider = "unkown"
	}
	tokenRegistry = make(map[string]*TokenProfile, 0)
	tokenRegistry[token] = &TokenProfile{} // TokenConfig
	tokenRegistry[token].Enabled = true
	tokenRegistry[token].Provider = provider
	tokenRegistry[token].Next = true
	tokenRegistry[token].Wait = true
	return tokenRegistry
}

type tokenLimitStats struct {
	enabled     bool
	performance struct {
		qps float32 // query per seconds
		eps float32
	}
	cache struct {
		requests int
		hit      int
		miss     int
	}
	expected struct {
		rateLimits int // expected rate limit triggers
		tasks      int // remaining concurrent group of jobs to finish
		jobs       int // remnaining jobs to complete
	}
	enqueued struct {
		tasks int // remaining concurrent group of jobs to finish
		jobs  int // remnaining jobs to complete
	}
	done struct {
		tasks int // count for groups of jobs done
		jobs  int //  count for jobs done
	}
}
