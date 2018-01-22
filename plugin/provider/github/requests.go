package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cnf/structhash"
	"github.com/google/go-github/github"
	"github.com/tiaotiao/mapstruct"

	// "github.com/abourget/llerrgroup"
	// requests
	// "github.com/franela/goreq"

	// "github.com/anacrolix/sync"
	// "github.com/k0kubun/pp"
	// "github.com/anacrolix/sync"
	// "github.com/viant/toolbox"
	// "github.com/thoas/go-funk"
	// "github.com/tuvistavie/structomap"
	// "github.com/src-d/enry/data"

	"github.com/sniperkit/xtask/util/runtime"
	"github.com/sniperkit/xutil/plugin/format/convert/mxj/pkg"
	"github.com/sniperkit/xutil/plugin/struct"
)

// Analyzing trends on Github using topic models and machine learning.
// var wg sync.WaitGroup

func (g *Github) counterTrack(name string, incr int) {
	go func() {
		g.counters.Increment(name, incr)
	}()
}

func (g *Github) GetFunc(entity string, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	if g.Client == nil {
		return nil, nil, errInvalidClient
	}

	/*
		if ExceededRateLimit(g.client) {
			log.Debugln("new client required as exceeded rate limit detected for the current token, token.old=", g.ctoken, "debug", runtime.WhereAmI())
			g = g.Manager.Fetch()
		}
	*/

	switch entity {
	case "getStars":
		return getStars(g, opts)

	case "getRepoList":
		return getRepoList(g, opts)

	case "getUser":
		return getUser(g, opts)

	case "getUserNode":
		return getUserNode(g, opts)

	case "getFollowers":
		return getFollowers(g, opts)

	case "getFollowing":
		return getFollowing(g, opts)

	case "getRepo":
		return getRepo(g, opts)

	case "getReadme":
		return getReadme(g, opts)

	case "getTree":
		return getTree(g, opts)

	case "getLanguages":
		return getLanguages(g, opts)

	case "getTopics":
		return getTopics(g, opts)

	case "getLatestSHA":
		return getLatestSHA(g, opts)

	}

	return nil, nil, nil
}

func Do(g *Github, entity string, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	if g.Client == nil {
		return nil, nil, errInvalidClient
	}

	/*
		if ExceededRateLimit(g.Client) {
			log.Debugln("new client required as exceeded rate limit detected for the current token, token.old=", g.ctoken, "debug", runtime.WhereAmI())
			g = g.Manager.Fetch()
		}
	*/

	switch entity {
	case "getStars":
		return getStars(g, opts)

	case "getRepoList":
		return getRepoList(g, opts)

	case "getUser":
		return getUser(g, opts)

	case "getUserNode":
		return getUserNode(g, opts)

	case "getFollowers":
		return getFollowers(g, opts)

	case "getFollowing":
		return getFollowing(g, opts)

	case "getRepo":
		return getRepo(g, opts)

	case "getReadme":
		return getReadme(g, opts)

	case "getTree":
		return getTree(g, opts)

	case "getLanguages":
		return getLanguages(g, opts)

	case "getTopics":
		return getTopics(g, opts)

	case "getLatestSHA":
		return getLatestSHA(g, opts)

	}

	return nil, nil, nil
}

func nextClient(g *Github, response *github.Response) *Github {
	log.Warnln("new client required, token.old=", g.ctoken, "debug", runtime.WhereAmI())
	go func() {
		g.wg.Add(1)
		defer g.wg.Done()
		g.Reclaim((*response).Reset.Time)
	}()
	return g.Manager.Fetch()
}

func (g *Github) nextClient(response *github.Response) *Github {
	log.Warnln("new client required, token.old=", g.ctoken, "debug", runtime.WhereAmI())
	go func() {
		g.wg.Add(1)
		defer g.wg.Done()
		log.Println("g=", g != nil, "response=", response != nil)
		g.Reclaim((*response).Reset.Time)
	}()

	return g.Manager.Fetch()
}

func getRepoList(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		repos    []*github.Repository
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			repos, response, err = g.Client.Repositories.List(
				context.Background(),
				opts.Target.Owner,
				&github.RepositoryListOptions{
					Sort:      "updated",
					Direction: "desc",
					ListOptions: github.ListOptions{
						Page:    opts.Page,
						PerPage: opts.PerPage,
					},
				},
			)
			if response == nil {
				return err
			}

			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}
		if response == nil {
			err = errorResponseIsNull
			goto finish
		}
		if repos == nil {
			err = errorMarshallingResponse
			goto finish
		}

		for _, repo := range repos {
			key := fmt.Sprintf("%s/%d/%d", repo.GetFullName(), repo.GetID(), repo.GetStargazersCount())
			mv := mxj.Map(structs.Map(repo))
			if opts.Filter != nil {
				if opts.Filter.Maps != nil {
					res[key] = extractWithMaps(mv, opts.Filter.Maps)
				}
			}
		}
		goto finish
	}

finish:
	return res, response, nil

}

func getStars(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		svc      *Github = g
		stars    []*github.StarredRepository
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)

	// client = clientManager.Fetch()

	goto request

request:
	{
		get := func() error {
			var err error
			stars, response, err = svc.Client.Activity.ListStarred(
				context.Background(),
				opts.Runner,
				&github.ActivityListStarredOptions{
					Sort:      "updated",
					Direction: "desc",
					ListOptions: github.ListOptions{
						Page:    opts.Page,
						PerPage: opts.PerPage,
					},
				},
			)
			if response == nil {
				return err
			}

			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}

		if response == nil {
			err = errorResponseIsNull
			goto finish
		}
		if stars == nil {
			err = errorMarshallingResponse
			goto finish
		}

		for _, star := range stars {
			key := fmt.Sprintf("%s/%d/%d", star.Repository.GetFullName(), star.Repository.GetID(), star.Repository.GetStargazersCount())
			mv := mxj.Map(structs.Map(star.Repository))
			if opts.Filter != nil {
				if opts.Filter.Maps != nil {
					res[key] = extractWithMaps(mv, opts.Filter.Maps)
				}
			}
		}
		goto finish
	}

finish:
	return res, response, nil

}

func getUser(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		svc      *Github = g
		user     *github.User
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)

	opts.Runner = "me"
	if opts.Target == nil {
		opts.Target = &Target{}
		opts.Target.Owner = "me"
	}

	goto request

request:
	{

		get := func() error {
			var err error
			user, response, err = svc.Client.Users.Get(context.Background(), opts.Target.Owner)
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}
		if response == nil {
			err = errorResponseIsNull
			goto finish
		}
		if user == nil {
			err = errorMarshallingResponse
			goto finish
		}

		mv := mxj.Map(structs.Map(user))
		if opts.Filter != nil {
			if opts.Filter.Maps != nil {
				res = extractWithMaps(mv, opts.Filter.Maps)
			}
		}
		res["request_url"] = response.Request.URL.String()
		res["object_hash"] = fmt.Sprintf("%x", structhash.Sha1(res, 1))
		goto finish
	}

	/*
	   changeClient:
	   	{
	   		g = g.nextClient(response)
	   		// g = nextClient(g, response)
	   		goto request
	   	}
	*/

finish:
	return res, response, nil
}

func getFollowers(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		users    []*github.User
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			users, response, err = g.Client.Users.ListFollowers(context.Background(), opts.Target.Owner, &github.ListOptions{Page: opts.Page, PerPage: opts.PerPage})
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}

		if response == nil {
			err = errorResponseIsNull
			goto finish
		}

		if users == nil {
			err = errorMarshallingResponse
			goto finish
		}

		for _, user := range users {
			key := fmt.Sprintf("%s/%d", user.GetLogin(), user.GetID())
			mv := mxj.Map(structs.Map(user))
			if mv == nil {
				continue
			}
			// pp.Println("mv=", mv, "opts=", opts, "key=", key)
			if opts.Filter != nil {
				// key := fmt.Sprintf("followers-%s", opts.Target.Owner)
				// row := make(map[string]interface{}, 0)
				if opts.Filter.Maps != nil {
					// row = extractWithMaps(mv, opts.Filter.Maps)
					// row["parent"] = opts.Target.Owner
					// res[key] = row //extractWithMaps(mv, opts.Filter.Maps)
					res[key] = extractWithMaps(mv, opts.Filter.Maps)
				}
			} else {
				res[key] = mv
			}
		}
		// res["request_url"] = response.Request.URL.String()
		// res["object_hash"] = fmt.Sprintf("%x", structhash.Sha1(res, 1))

		goto finish
	}

finish:
	return res, response, err
}

func getFollowing(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		users    []*github.User
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			users, response, err = g.Client.Users.ListFollowing(context.Background(), opts.Target.Owner, &github.ListOptions{Page: opts.Page, PerPage: opts.PerPage})
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}

		if response == nil {
			err = errorResponseIsNull
			goto finish
		}

		if users == nil {
			err = errorMarshallingResponse
			goto finish
		}

		for _, user := range users {
			key := fmt.Sprintf("%s/%d", user.GetLogin(), user.GetID())
			mv := mxj.Map(structs.Map(user))
			if mv == nil {
				continue
			}

			if opts.Filter != nil {
				// key := fmt.Sprintf("following-%s", opts.Target.Owner)
				// row := make(map[string]interface{}, 0)
				if opts.Filter.Maps != nil {
					//row = extractWithMaps(mv, opts.Filter.Maps)
					//row["parent"] = opts.Target.Owner
					//res[key] = row //extractWithMaps(mv, opts.Filter.Maps)
					res[key] = extractWithMaps(mv, opts.Filter.Maps)
				}
			} else {
				res[key] = mv
			}
			/*
				if opts.Filter != nil {
					if opts.Filter.Maps != nil {
						res[key] = extractWithMaps(mv, opts.Filter.Maps)
					}
				}
			*/
		}

		// res["request_url"] = response.Request.URL.String()
		// res["object_hash"] = fmt.Sprintf("%x", structhash.Sha1(res, 1))
		goto finish
	}

finish:
	return res, response, err
}

func getRepo(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		repo     *github.Repository
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			repo, response, err = g.Client.Repositories.Get(context.Background(), opts.Target.Owner, opts.Target.Name)
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}
		if response == nil {
			err = errorResponseIsNull
			goto finish
		}
		if repo == nil {
			err = errorMarshallingResponse
			goto finish
		}

		mv := mxj.Map(structs.Map(repo))
		if opts.Filter != nil {
			if opts.Filter.Maps != nil {
				res = extractWithMaps(mv, opts.Filter.Maps)
			}
		}
		res["request_url"] = response.Request.URL.String()
		res["object_hash"] = fmt.Sprintf("%x", structhash.Sha1(res, 1))
		goto finish
	}

finish:
	return res, response, err
}

func getTopics(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		topics   []string
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			topics, response, err = g.Client.Repositories.ListAllTopics(context.Background(), opts.Target.Owner, opts.Target.Name)
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}
		if response == nil {
			err = errorResponseIsNull
			goto finish
		}
		if topics == nil {
			err = errorMarshallingResponse
			goto finish
		}

		for _, topic := range topics {
			key := fmt.Sprintf("%s", topic)
			row := make(map[string]interface{}, 0)
			row["label"] = topic
			row["owner"] = opts.Target.Owner
			row["name"] = opts.Target.Name
			row["remote_id"] = strconv.Itoa(opts.Target.RepoId)
			row["request_url"] = response.Request.URL.String()
			res[key] = row
		}
		goto finish
	}

finish:
	return res, response, nil
}

func getLanguages(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		langs    map[string]int
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			langs, response, err = g.Client.Repositories.ListLanguages(context.Background(), opts.Target.Owner, opts.Target.Name)
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}
		if response == nil {
			err = errorResponseIsNull
			goto finish
		}
		if langs == nil {
			err = errorMarshallingResponse
			goto finish
		}

		for lang, _ := range langs {
			key := fmt.Sprintf("%s", lang)
			row := make(map[string]interface{}, 0)
			row["lang"] = lang
			row["owner"] = opts.Target.Owner
			row["name"] = opts.Target.Name
			row["remote_id"] = strconv.Itoa(opts.Target.RepoId)
			row["request_url"] = response.Request.URL.String()
			res[key] = row
		}
		goto finish
	}

finish:
	return res, response, nil
}

func getLatestSHA(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		ref      *github.Reference
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			if opts.Target.Branch == "" {
				opts.Target.Branch = "master"
			}
			ref, response, err = g.Client.Git.GetRef(context.Background(), opts.Target.Owner, opts.Target.Name, "refs/heads/"+opts.Target.Branch)
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}
		if response == nil {
			err = errorResponseIsNull
			goto finish
		}
		if ref == nil {
			err = errorMarshallingResponse
			goto finish
		}

		res["sha"] = *ref.Object.SHA
		res["owner"] = opts.Target.Owner
		res["name"] = opts.Target.Name
		res["branch"] = opts.Target.Branch
		res["remote_repo_id"] = opts.Target.RepoId
		res["request_url"] = response.Request.URL.String()
		res["object_hash"] = fmt.Sprintf("%x", structhash.Sha1(res, 1))
		goto finish
	}

finish:
	return res, response, nil
}

func getTree(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		tree     *github.Tree
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			tree, response, err = g.Client.Git.GetTree(context.Background(), opts.Target.Owner, opts.Target.Name, opts.Target.Ref, true)
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}
		if response == nil {
			err = errorResponseIsNull
			goto finish
		}
		if tree == nil {
			err = errorMarshallingResponse
			goto finish
		}

		for k, entry := range tree.Entries {
			row := make(map[string]interface{}, 5)

			entry_path := entry.GetPath()
			/*
				if len(opts.Target.Filters) > 0 {
					if Ignore(entry_path, opts.Target.Filters) {
						log.Debugln("[FILTER] entry.path=", entry_path)
						continue
					}
				}
			*/

			row["path"] = entry_path
			row["owner"] = opts.Target.Owner
			row["name"] = opts.Target.Name
			row["remote_id"] = strconv.Itoa(opts.Target.RepoId)
			row["request_url"] = response.Request.URL.String()
			// row["sha"] = entry.GetSHA()
			// row["size"] = entry.GetSize()
			// row["url"] = entry.GetURL()
			key := fmt.Sprintf("entry-%d", k)
			res[key] = row
		}
		goto finish
	}

finish:
	return res, response, nil
}

func getReadme(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	var (
		readme   *github.RepositoryContent
		response *github.Response
		res      = make(map[string]interface{}, 0)
		err      error
	)
	goto request

request:
	{
		get := func() error {
			var err error
			readme, response, err = g.Client.Repositories.GetReadme(context.Background(), opts.Target.Owner, opts.Target.Name, nil)
			if response == nil {
				return err
			}
			return err
		}

		if err = retryRegistrationFunc(get); err != nil {
			log.Errorln("exceeded?", strings.Contains(err.Error(), "exceeded"), "error: ", err, "debug=", runtime.WhereAmI())
			goto finish
		}

		if response == nil {
			err = errorResponseIsNull
			goto finish
		}

		if readme == nil {
			err = errorMarshallingResponse
			goto finish
		}

		content, _ := readme.GetContent()
		readme.Content = &content

		mv := mxj.Map(structs.Map(readme))
		if opts.Filter != nil {
			if opts.Filter.Maps != nil {
				res = extractWithMaps(mv, opts.Filter.Maps)
			}
		}

		res["owner"] = opts.Target.Owner
		res["name"] = opts.Target.Name
		res["remote_repo_id"] = opts.Target.RepoId
		res["request_url"] = response.Request.URL.String()
		res["object_hash"] = fmt.Sprintf("%x", structhash.Sha1(res, 1))
		goto finish
	}

finish:
	return res, response, nil
}

// Run starts the dispatcher and pushes a new request for the root user onto
// the queue. Returns the *UserNode that is received on the done channel.
func getUserNode(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) { // *UserNode { // start, end string, opts.Workers int, g *Github) *UserNode {
	// defaultCLI, ctx = g.client, context.Background()

	if opts.Workers <= 0 {
		opts.Workers = 6
	}

	startDispatcher(opts.Workers)
	origin, target = opts.Start, opts.End
	jobQueue <- jobRequest{User: newUserNode(origin, nil)}

	for {
		select {
		case user := <-done:
			// pp.Println(user)
			return mapstruct.Struct2Map(user), nil, nil
		}
	}

	//finish:
	//	return res, response, nil
}

/*
func getRepoBranchSHA(g *Github, opts *Options) (map[string]interface{}, *github.Response, error) {
	defer funcTrack(time.Now())

	if err := s.waitForRate(); err != nil {
		return "", err
	}
	b, _, err := s.client.Repositories.GetBranch(owner, repo, branch)
	if err != nil {
		if isNotFound(err) {
			return "", errorsp.WithStacksAndMessage(ErrInvalidRepository, "GetBranch %v %v %v failed", owner, repo, branch)
		}
		return "", errorsp.WithStacksAndMessage(err, "GetBranch %v %v %v failed", owner, repo, branch)
	}
	if b.Commit == nil {
		return "", nil
	}
	return stringsp.Get(b.Commit.SHA), nil
}
*/

// verifyRepo checks all essential fields of a Repository structure for nil
// values. An error is returned if one of the essential field is nil.
func verifyRepo(repo *github.Repository) error {
	if repo == nil {
		return newInvalidStructError("verifyRepo: repo is nil")
	}

	var err *invalidStructError
	if repo.ID == nil {
		err = newInvalidStructError("verifyRepo: contains nil fields:").AddField("ID")
	} else {
		err = newInvalidStructError(fmt.Sprintf("verifyRepo: repo #%d contains nil fields: ", *repo.ID))
	}

	if repo.Name == nil {
		err.AddField("Name")
	}

	if repo.Language == nil {
		err.AddField("Language")
	}

	if repo.CloneURL == nil {
		err.AddField("CloneURL")
	}

	if repo.Owner == nil {
		err.AddField("Owner")
	} else {
		if repo.Owner.Login == nil {
			err.AddField("Owner.Login")
		}
	}

	if repo.Fork == nil {
		err.AddField("Fork")
	}

	if err.FieldsLen() > 0 {
		return err
	}

	return nil
}

func extractWithMaps(mv mxj.Map, fields map[string]string) map[string]interface{} {
	l := make(map[string]interface{}, len(fields))
	for key, path := range fields {
		var node []interface{}
		var merr error
		node, merr = mv.ValuesForPath(path)
		if merr != nil {
			log.Fatalln("Error: ", merr, "key=", key, "path=", path)
			continue
		}

		/*
			switch length := len(node); {
			case length == 1:
				l[key] = node[0]
			case length > 1:
				l[key] = node
			default:
				continue
			}
		*/
		if len(node) > 1 {
			l[key] = node
		} else if len(node) == 1 {
			l[key] = node[0]
		}
	}
	return l
}

func extractFlatten(mv mxj.Map, fields []string) map[string]interface{} {
	l := make(map[string]interface{}, len(fields))
	for _, path := range fields {
		// var node []interface{}
		// var merr error
		node, _ := mv.ValuesForPath(path)
		// if merr != nil {
		//	log.Fatalln("Error: ", merr)
		// }
		if node != nil {
			log.Warnln("node len=", len(node))
			if len(node) > 2 {
				l[path] = node
			} else {
				l[path] = node[0]
			}
		}
	}
	return l
}

func extractBlocks(mv mxj.Map, items string, fields map[string][]string) map[string]interface{} {
	l := make(map[string]interface{}, len(fields))
	for attr, field := range fields {
		var keyPath string
		var node []interface{}
		if len(field) == 1 {
			keyPath = fmt.Sprintf("%#s", field[0])
			node, _ = mv.ValuesForPath(keyPath)
			// log.Debugln("attr=", attr, "keyPath=", keyPath, "node=", node)
		} else {
			w := make(map[string]interface{}, len(field))
			var merr error
			for _, whl := range field {
				keyParts := strings.Split(whl, ".")
				keyName := keyParts[len(keyParts)-1]
				keyPath = fmt.Sprintf("%#s", whl)
				// log.Debugln("attr=", attr, "keyPath=", keyPath, "keyName=", keyName)
				node, merr = mv.ValuesForPath(keyPath)
				if merr != nil {
					log.Fatalln("Error: ", merr)
				}
				if node != nil {
					if len(node) == 1 {
						w[keyName] = node[0]
					} else if len(node) > 1 {
						w[keyName] = node
					}
				}
			}
			l[attr] = w
			continue
		}
		if len(node) == 1 {
			l[attr] = node[0]
			// log.Debugln("attr=", attr, "node[0]=", node[0])
		} else if len(node) > 1 {
			// log.Debugln("attr=", attr, "node=", node)
			l[attr] = node
		}
	}
	// log.Println(l)
	return l
}
