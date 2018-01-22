package github

const (
	DateFormat  = "2006-01-02"
	serviceName = "github"

	coreCategory rateLimitCategory = iota
	searchCategory
	categories

	QueryCondLanguage = "language"
	QueryCondCreated  = "created"

	LangC       = "C"
	LangCSharp  = "C#"
	LangCPlus   = "C++"
	LangCSS     = "CSS"
	LangGo      = "Go"
	LangHTML    = "HTML"
	LangJava    = "Java"
	LangJS      = "JavaScript"
	LangLua     = "Lua"
	LangObjC    = "Objective-C"
	LangPHP     = "PHP"
	LangPython  = "Python"
	LangR       = "R"
	LangRuby    = "Ruby"
	LangScala   = "Scala"
	LangShell   = "Shell"
	LangSwift   = "Swift"
	LangAS      = "ActionScript"
	LangClojure = "Clojure"
	LangCS      = "CoffeeScript"
	LangHaskell = "Haskell"
	LangMatlab  = "Matlab"
	LangPerl    = "Perl"
	LangTeX     = "TeX"
	LangVS      = "VimScript"
	LangErlang  = "Erlang"
	LangKotlin  = "Kotlin"
	LangSQL     = "SQL"
	LangTS      = "TypeScript"
	LangVue     = "Vue"

	SortByStars   = "stars"
	SortByForks   = "forks"
	SortByUpdated = "updated"

	OrderByAsc  = "asc"
	OrderByDesc = "desc"

	Quarter = "quarter"
	Month   = "month"
	Week    = "week"
	Day     = "day"
)
