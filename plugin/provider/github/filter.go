package github

import (
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/k0kubun/pp"

	"github.com/google/go-github/github"

	kf "github.com/miraclesu/keywords-filter"
	goac "github.com/sniperkit/xfilter/backend/goac"

	cuckoo "github.com/seiflotfy/cuckoofilter"
	"github.com/sniperkit/cuckoofilter"
	"github.com/sniperkit/xtask/plugin/counter"
)

/*
	Refs:
	- https://github.com/client9/gosupplychain/blob/master/github.go#L41-L49
*/

var (
	cfMax     *uint32
	cfVisited *cuckoo.CuckooFilter
	cfDone    *cuckoofilter.Filter
	cf404     *cuckoofilter.Filter
	fx        *kf.Filter
	wac       *goac.AhoCorasick      = goac.NewAhoCorasick()
	wai       map[string]*FilterInfo = make(map[string]*FilterInfo, 0)

	Threshold int = 50

	depRegexp    *regexp.Regexp
	DOC_EXTS     = []string{"md", "markdown", "mdown", "mkdn", "mdwn", "mdtxt", "txt", "text", "doc", "htm", "html"}
	DEPS_REGEXPS = []string{`(^|/)cache/`, `^[Dd]ependencies/`, `^deps/`, `^tools/`, `(^|/)configure$`, `(^|/)configure.ac$`, `(^|/)config.guess$`, `(^|/)config.sub$`, `cpplint.py`, `node_modules/`, `bower_components/`, `^rebar$`, `erlang.mk`, `Godeps/_workspace/`, `(\.|-)min\.(js|css)$`, `([^\s]*)import\.(css|less|scss|styl)$`, `(^|/)bootstrap([^.]*)\.(js|css|less|scss|styl)$`, `(^|/)custom\.bootstrap([^\s]*)(js|css|less|scss|styl)$`, `(^|/)font-awesome\.(css|less|scss|styl)$`, `(^|/)foundation\.(css|less|scss|styl)$`, `(^|/)normalize\.(css|less|scss|styl)$`, `(^|/)[Bb]ourbon/.*\.(css|less|scss|styl)$`, `(^|/)animate\.(css|less|scss|styl)$`, `third[-_]?party/`, `3rd[-_]?party/`, `vendors?/`, `extern(al)?/`, `(^|/)[Vv]+endor/`, `^debian/`, `run.n$`, `bootstrap-datepicker/`, `(^|/)jquery([^.]*)\.js$`, `(^|/)jquery\-\d\.\d+(\.\d+)?\.js$`, `(^|/)jquery\-ui(\-\d\.\d+(\.\d+)?)?(\.\w+)?\.(js|css)$`, `(^|/)jquery\.(ui|effects)\.([^.]*)\.(js|css)$`, `jquery.fn.gantt.js`, `jquery.fancybox.(js|css)`, `fuelux.js`, `(^|/)jquery\.fileupload(-\w+)?\.js$`, `(^|/)slick\.\w+.js$`, `(^|/)Leaflet\.Coordinates-\d+\.\d+\.\d+\.src\.js$`, `leaflet.draw-src.js`, `leaflet.draw.css`, `Control.FullScreen.css`, `Control.FullScreen.js`, `leaflet.spin.js`, `wicket-leaflet.js`, `.sublime-project`, `.sublime-workspace`, `(^|/)prototype(.*)\.js$`, `(^|/)effects\.js$`, `(^|/)controls\.js$`, `(^|/)dragdrop\.js$`, `(.*?)\.d\.ts$`, `(^|/)mootools([^.]*)\d+\.\d+.\d+([^.]*)\.js$`, `(^|/)dojo\.js$`, `(^|/)MochiKit\.js$`, `(^|/)yahoo-([^.]*)\.js$`, `(^|/)yui([^.]*)\.js$`, `(^|/)ckeditor\.js$`, `(^|/)tiny_mce([^.]*)\.js$`, `(^|/)tiny_mce/(langs|plugins|themes|utils)`, `(^|/)MathJax/`, `(^|/)Chart\.js$`, `(^|/)[Cc]ode[Mm]irror/(\d+\.\d+/)?(lib|mode|theme|addon|keymap|demo)`, `(^|/)shBrush([^.]*)\.js$`, `(^|/)shCore\.js$`, `(^|/)shLegacy\.js$`, `(^|/)angular([^.]*)\.js$`, `(^|\/)d3(\.v\d+)?([^.]*)\.js$`, `(^|/)react(-[^.]*)?\.js$`, `(^|/)modernizr\-\d\.\d+(\.\d+)?\.js$`, `(^|/)modernizr\.custom\.\d+\.js$`, `(^|/)knockout-(\d+\.){3}(debug\.)?js$`, `(^|/)docs?/_?(build|themes?|templates?|static)/`, `(^|/)admin_media/`, `^fabfile\.py$`, `^waf$`, `^.osx$`, `\.xctemplate/`, `\.imageset/`, `^Carthage/`, `^Pods/`, `(^|/)Sparkle/`, `Crashlytics.framework/`, `Fabric.framework/`, `gitattributes$`, `gitignore$`, `gitmodules$`, `(^|/)gradlew$`, `(^|/)gradlew\.bat$`, `(^|/)gradle/wrapper/`, `-vsdoc\.js$`, `\.intellisense\.js$`, `(^|/)jquery([^.]*)\.validate(\.unobtrusive)?\.js$`, `(^|/)jquery([^.]*)\.unobtrusive\-ajax\.js$`, `(^|/)[Mm]icrosoft([Mm]vc)?([Aa]jax|[Vv]alidation)(\.debug)?\.js$`, `^[Pp]ackages\/.+\.\d+\/`, `(^|/)extjs/.*?\.js$`, `(^|/)extjs/.*?\.xml$`, `(^|/)extjs/.*?\.txt$`, `(^|/)extjs/.*?\.html$`, `(^|/)extjs/.*?\.properties$`, `(^|/)extjs/.sencha/`, `(^|/)extjs/docs/`, `(^|/)extjs/builds/`, `(^|/)extjs/cmd/`, `(^|/)extjs/examples/`, `(^|/)extjs/locale/`, `(^|/)extjs/packages/`, `(^|/)extjs/plugins/`, `(^|/)extjs/resources/`, `(^|/)extjs/src/`, `(^|/)extjs/welcome/`, `(^|/)html5shiv\.js$`, `^[Tt]ests?/fixtures/`, `^[Ss]pecs?/fixtures/`, `(^|/)cordova([^.]*)\.js$`, `(^|/)cordova\-\d\.\d(\.\d)?\.js$`, `foundation(\..*)?\.js$`, `^Vagrantfile$`, `.[Dd][Ss]_[Ss]tore$`, `^vignettes/`, `^inst/extdata/`, `octicons.css`, `sprockets-octicons.scss`, `(^|/)activator$`, `(^|/)activator\.bat$`, `proguard.pro`, `proguard-rules.pro`, `^puphpet/`, `(^|/)\.google_apis/`}

	typeListFiles = map[string][]string{
		"cmake":    []string{"CMakeLists.txt"},
		"docker":   []string{"Dockerfile"},
		"crystal":  []string{"Projectfile"},
		"markdown": []string{".markdown", ".md", ".mdown", ".mkdn"},
		"asciidoc": []string{".adoc", ".asc", ".asciidoc"},
		"groovy":   []string{".groovy", ".gradle"},
		"msbuild":  []string{".csproj", ".fsproj", ".vcxproj", ".proj", ".props", ".targets"},
		"wiki":     []string{".mediawiki", ".wiki"},
		"make":     []string{"gnumakefile", "Gnumakefile", "makefile", "Makefile"},
		"mk":       []string{"mkfile"},
		"ruby":     []string{"Gemfile", ".irbrc", "Rakefile"},
		"toml":     []string{"Cargo.lock"},
		"zsh":      []string{"zshenv", ".zshenv", "zprofile", ".zprofile", "zshrc", ".zshrc", "zlogin", ".zlogin", "zlogout", ".zlogout"},
	}

	lang_path_map = map[string]string{
		`rakefile`: `ruby`,
		`/(Makefile|CMakeLists.txt|Imakefile|makepp|configure)$`: `make`,
		`/config$`:     `conf`,
		`/zsh/_[^/]+$`: `sh`,
		`patch`:        `diff`,
	}
)

type FilterInfo struct {
	Ignore  bool
	Extract bool
	Regexp  []string
}

func HasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {

			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}

func leafPathsPatterns(input []string) []string {
	var output []string
	var re = regexp.MustCompile(`.([0-9]+)`)
	for _, value := range input {
		value = re.ReplaceAllString(value, `[*]`)
		if !contains(output, value) {
			output = append(output, value)
		}
	}
	return dedup(output)
}

func GlobalFilters(filters map[string]*FilterInfo) *goac.AhoCorasick {
	defer funcTrack(time.Now())
	if wac == nil {
		wac = goac.NewAhoCorasick()
	}
	wai = filters
	for group, params := range filters {
		wac.AddPatterns(group, params.Regexp...)
	}
	pp.Println("filters=", filters)
	wac.Build()
	return wac
}

func AddWordFilters(filters map[string]*FilterInfo) *goac.AhoCorasick {
	defer funcTrack(time.Now())
	wai = filters
	for group, params := range filters {
		wac.AddPatterns(group, params.Regexp...)
	}
	pp.Println("filters=", filters)
	return wac
}

func Scan(s string, filters []string) []map[string]interface{} {
	sr := make([]map[string]interface{}, 0)
	results := wac.Scan(s)
	for _, result := range results {
		if contains(filters, result.Group) {
			output := map[string]interface{}{
				`input`:   s,
				`match`:   string([]rune(s)[result.Start : result.End+1]),
				`group`:   result.Group,
				`start`:   result.Start,
				`end`:     result.End + 1,
				`ignore`:  wai[result.Group].Ignore,
				`extract`: wai[result.Group].Extract,
			}
			log.Println("sr=", sr)
			if wai[result.Group].Ignore {
				return sr
			}
			sr = append(sr, output)
		}
	}
	return sr
}

func Ignore(s string, filters []string) bool {
	results := wac.Scan(s)
	for _, result := range results {
		if contains(filters, result.Group) {
			if wai[result.Group].Ignore {
				log.Println("[IGNORE] input=", s, "group=", result.Group)
				return true
			}
		}
	}
	return false
}

func contains(input []string, match string) bool {
	for _, value := range input {
		if value == match {
			return true
		}
	}
	return false
}

type ScanResult struct {
	Input string
	match string
	Tags  string
	Group string
	start int
	end   int
}

func filterKeywords(input string) []map[string]interface{} {

	sr := make([]map[string]interface{}, 0)
	results := wac.Scan(input)
	for _, result := range results {
		sr = append(sr, map[string]interface{}{
			`tags`:  string([]rune(input)[result.Start : result.End+1]),
			`group`: result.Group,
			`start`: result.Start,
			`end`:   result.End + 1,
		})
		// log.Println("match=", string([]rune(output)[result.Start:result.End+1]), ", group=", result.Group, ", start=", result.Start, ", end=", result.End+1)
	}
	return sr
}

func dedup(input []string) []string {
	var output []string
	for _, value := range input {
		if !contains(output, value) {
			output = append(output, value)
		}
	}
	return output
}

func filterTree(filepath string) bool {
	// data.DocumentationMatchers
	// data.LanguagesByFilename
	// data.LanguagesByInterpreter
	return false
}

func strPtr(s string) *string { return &s }

func interfacePtr(i interface{}) interface{} {
	return &i
}

func findByPath(entries []github.TreeEntry, path string) *github.TreeEntry {
	for _, entry := range entries {
		if *entry.Path == path {
			return &entry
		}
	}
	return nil
}

func (g *Github) CacheCount() int {
	defer funcTrack(time.Now())

	if g.cfVisited == nil {
		return 0
	}
	return int(g.cfVisited.Count())
}

func (g *Github) LoadCache(max int, prefix string, remove string, stopPatterns []string) bool {
	defer funcTrack(time.Now())

	if g.Client == nil {
		g.Client = getClient(g.ctoken)
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.cfMax == nil && g.xcache != nil {
		g.counters.Increment("cache.load", 1)
		g.counters.Increment("cache.maxKeys", max)
		maxKeys := uint32(max)
		g.cfMax = &maxKeys
		existingKeys, _ := g.xcache.Action("getKeys", nil)
		loaded := 0
		g.cfVisited, loaded = getCached(g.counters, maxKeys, &prefix, &remove, &stopPatterns, &existingKeys)
		log.Println("[g.LoadCache] loaded=", loaded, " / total.cache=", len(existingKeys))
		g.counters.Increment("cache.keys.loaded", loaded)
		g.counters.Increment("cache.keys.existing", len(existingKeys))
		return true
	}
	return false
}

func (g *Github) CacheSlugExists(slug string) bool {
	defer funcTrack(time.Now())

	return g.cfVisited.Lookup([]byte(slug))
}

func getCached(cnt *counter.Oc, maxKeys uint32, prefix *string, remove *string, stopPatterns *[]string, existingKeys *map[string]*interface{}) (*cuckoo.CuckooFilter, int) {
	defer funcTrack(time.Now())

	registry := cuckoo.NewDefaultCuckooFilter()
	for key, _ := range *existingKeys {
		slug := key
		if prefix != nil {
			if !strings.HasPrefix(slug, *prefix) {
				if cnt != nil {
					cnt.Increment("github.cache.skipped", 1)
				}
				continue
			}
		}
		if stopPatterns != nil {
			var skip bool
			for _, pattern := range *stopPatterns {
				if strings.Contains(slug, pattern) {
					if cnt != nil {
						cnt.Increment("github.cache.skipped", 1)
					}
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		if remove != nil {
			slug = strings.Replace(slug, *remove, "", -1)
			if cnt != nil {
				cnt.Increment("github.cache.remove.substring", 1)
			}
		}
		registry.InsertUnique([]byte(slug))
		if cnt != nil {
			cnt.Increment("github.cache.entry", 1)
		}
	}
	return registry, int(registry.Count())
}

// isLanguageWanted checks if language(s) is in the list of wanted languages.
func isLanguageWanted(suppLangs []string, prjLangs interface{}) (bool, error) {
	if prjLangs == nil {
		return false, nil
	}

	switch prjLangs.(type) {
	case map[string]int:
		langs := prjLangs.(map[string]int)
		for k := range langs {
			for _, v := range suppLangs {
				if strings.EqualFold(k, v) {
					return true, nil
				}
			}
		}
	case *string:
		lang := prjLangs.(*string)
		if lang == nil {
			return false, nil
		}

		for _, sl := range suppLangs {
			if sl == *lang {
				return true, nil
			}
		}
	default:
		return false, errorLanguageNotSupported
	}

	return false, nil
}

/*
func (g *Github) WithLocalFilters(filter *FilterInfo) (g *Github) {
	defer funcTrack(time.Now())
	g.mu.Lock()
	defer g.mu.Unlock()

	var err error
	filter, err = kf.New(*Threshold, nil)
	if err != nil {
		log.Println(err.Error())
		return g
	}
	g.filter = filter
	return
}

func (g *Github) AddSymFilters(filters map[string][]string) (g *Github) {
	defer funcTrack(time.Now())
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, words := range filters {
		for _, word := range words {
			g.filter.AddSymb("Makefile")
		}
	}
	return
}

func (g *Github) AddWordFilters(filters map[string][]string) (g *Github) {
	defer funcTrack(time.Now())
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, words := range filters {
		g.filter.AddWords(words)
	}
	return
}

func contentFilter(g *Github, s string, ignore bool) bool {
	results := g.filter.Scan(s)
	ok := len(results) > 0
	if ignore {
		return !ok
	}
	return ok
}

func (g *Github) ContentFilter(s string, ignore bool) bool {
	keywordMap := make(map[string]struct{}) // use the keyword map to deduplicate keywords for same line
	for key, info := range g.keywordMapping {
		for _, regex := range info.Regexp { // compare the passed string against every regexp possibility
			if _, ok := keywordMap[key]; !ok { // don't check again if the keyword has already been added
				r := regexp.MustCompile(`(?i)` + regex) // add i flag to make case insenstive
				if r.Match([]byte(s)) {
					keywordMap[key] = struct{}{}
					return ignore
				}
			}
		}
	}
	return false
}

func (g *Github) Match(s string) (matches []string) {
	// use the keyword map to deduplicate keywords for same line
	keywordMap := make(map[string]struct{})

	for key, info := range KeywordMapping {
		// compare the passed string against every regexp possibility
		for _, regex := range info.Regexp {
			// don't check again if the keyword has already been added
			if _, ok := keywordMap[key]; !ok {
				// add i flag to make case insenstive
				r := regexp.MustCompile(`(?i)` + regex)
				if r.Match([]byte(s)) {
					keywordMap[key] = struct{}{}
					// log.Printf("SubexpNames: %s\n", r.SubexpNames())
				}
			}
		}
	}

	// return the slice of keys for the keyword map
	for keyword := range keywordMap {
		matches = append(matches, keyword)
	}
	return matches
}
*/
