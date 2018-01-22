package github

import (
	"time"

	"github.com/google/go-github/github"

	"github.com/sniperkit/xtask/plugin/aggregate/service"
	"github.com/sniperkit/xtask/plugin/counter"
	"github.com/sniperkit/xtask/plugin/rate"
)

var (
	Service             *Github
	rateLimiters        = map[string]*rate.RateLimiter{}
	isBackoff           bool
	defaultOpts         *Options
	defaultRetryDelay   time.Duration = 150 * time.Millisecond
	defaultAbuseDelay   time.Duration = 5 * time.Second
	defaultRetryAttempt uint64        = 1
	defaultPrefixApi                  = "https://api.github.com/"
)

func New(tokens []*service.Token, opts *Options) *Github {
	defer funcTrack(time.Now())
	g := &Github{
		ctoken:       tokens[0].Key,
		ctokens:      tokens,
		coptions:     opts,
		rateLimiters: make(map[string]*rate.RateLimiter, len(tokens)),
		counters:     counter.NewOc(),
	}
	g.getClient(tokens[0].Key)
	return g

}

func Init() {
	defaultOpts = &Options{}
	defaultOpts.Page = 1
	defaultOpts.PerPage = 100
	Service = New(nil, defaultOpts)
}

func (Github) ProviderName() string {
	return serviceName
}

func (Github) PrefixApi() string {
	return defaultPrefixApi
}

type Context struct {
	Runner string
	Target *Target
}

type Profile struct {
	Owner    string
	Contribs bool
	Followed bool
	Starred  bool
}

type Target struct {
	Owner   string
	Name    string
	Branch  string
	Ref     string
	Filters []string
	OwnerId int
	RepoId  int
	Workers int
	Start   string
	End     string
}

type Output struct {
	Disabled bool
	Timeout  time.Duration
}

type Leafs struct {
	Paths  bool
	Nodes  bool
	Values bool
}

type Schema struct {
	Disabled bool
	Leafs    Leafs
}

type Filter struct {
	Disabled   bool
	UsePrefix  bool
	Validate   bool
	MaxSize    uint
	ExtFile    string
	PrefixFile string
	SuffixFile string
	PrefixPath string
	ChunkSize  int
	Root       string
	Leafs      Leafs
	Multi      map[string]map[string]string
	Blocks     map[string][]string
	Paths      []string
	Maps       map[string]string
	Timeout    time.Duration
}

func NewFilter() *Filter {
	export := &Filter{}
	export.Leafs = Leafs{}
	export.Blocks = make(map[string][]string, 0)
	export.Maps = make(map[string]string, 0)
	export.Multi = make(map[string]map[string]string, 0)
	return export
}

type Options struct {
	Runner               string
	Accounts             []string
	Page                 int
	PerPage              int
	Workers              int
	Start                string
	End                  string
	Target               *Target
	Filter               *Filter
	ActivityListStarred  *github.ActivityListStarredOptions
	RepositoryContentGet *github.RepositoryContentGetOptions
	Project              *github.ProjectOptions
	List                 *github.ListOptions
	Raw                  *github.RawOptions
	Search               *github.SearchOptions
}
