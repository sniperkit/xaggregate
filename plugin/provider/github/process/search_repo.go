package process

import (
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"

	"github.com/sniperkit/xtask/plugin/aggregate/log"
	gitClient "github.com/sniperkit/xtask/plugin/aggregate/service/github/client"
	"github.com/sniperkit/xtask/plugin/aggregate/service/github/model"
	"github.com/sniperkit/xtask/plugin/aggregate/util"
)

// StoreRepo 将库信息存储到数据库
func StoreRepo(repo *github.Repository, client *gitClient.GHClient) error {
	// 判断数据库中是否有此作者信息
	oldUserID, err := models.GitUserService.GetUserID(repo.Owner.Login)
	if err != nil {
		if err != mgo.ErrNotFound {
			return err
		}

		// MDUser 数据库中无此作者信息
		newOwner, _, err := GetOwnerByID(*repo.Owner.ID, client)
		if err != nil {
			return err
		}

		newUserID, err := models.GitUserService.Create(newOwner)
		if err != nil {
			return err
		}

		err = models.GitReposService.Create(repo, &newUserID)
		if err != nil {
			return err
		}
	} else {
		// MDUser 数据库中有此作者信息
		err = models.GitReposService.Create(repo, &oldUserID)
		if err != nil {
			return err
		}
	}

	return nil
}

// SearchRepos 从指定时间（库的创建时间）开始搜索，并将结果保存到数据库
func SearchRepos(year int, month time.Month, day int, incremental, querySeg string, opt *github.SearchOptions) {
	var (
		client  *gitClient.GHClient
		ok      bool
		wg      sync.WaitGroup
		e       *github.AbuseRateLimitError
		newDate []int
		result  []github.Repository
	)

	defer clientManager.Shutdown()

	client = clientManager.Fetch()

search:
	repos, resp, stopAt, err := SearchReposByStartTime(client, year, month, day, incremental, querySeg, opt)
	result = append(result, repos...)

	if err != nil {
		if _, ok = err.(*github.RateLimitError); ok {
			log.Logger.Error("SearchReposByStartTime hit limit error, it's time to change client.", zap.Error(err))

			goto changeClient
		} else if e, ok = err.(*github.AbuseRateLimitError); ok {
			log.Logger.Error("SearchReposByStartTime have triggered an abuse detection mechanism.", zap.Error(err))

			time.Sleep(*e.RetryAfter)
			goto search
		} else if strings.Contains(err.Error(), "timeout") {
			log.Logger.Info("SearchReposByStartTime has encountered a timeout error. Sleep for five minutes.")
			time.Sleep(5 * time.Minute)

			goto search
		} else {
			log.Logger.Error("SearchRepos terminated because of this error.", zap.Error(err))

			return
		}
	} else {

		goto store
	}

changeClient:
	{
		go func() {
			wg.Add(1)
			defer wg.Done()

			gitClient.Reclaim(client, resp)
		}()

		client = clientManager.Fetch()

		if stopAt != "" {
			newDate, err = utility.SplitDate(stopAt)
			if err != nil {
				log.Logger.Error("SplitDate returned error.", zap.Error(err))

				return
			}

			year = newDate[0]
			monthInt := newDate[1]
			switch monthInt {
			case 1:
				month = time.January
			case 2:
				month = time.February
			case 3:
				month = time.March
			case 4:
				month = time.April
			case 5:
				month = time.May
			case 6:
				month = time.June
			case 7:
				month = time.July
			case 8:
				month = time.August
			case 9:
				month = time.September
			case 10:
				month = time.October
			case 11:
				month = time.November
			case 12:
				month = time.December
			}
			day = newDate[2]

			goto search
		}

		log.Logger.Info("stopAt is empty string, stop searching.")
	}

store:
	log.Logger.Info("Start storing repositories now.")
	for _, repo := range result {
	repeatStore:
		err = StoreRepo(&repo, client)
		if err != nil {
			if _, ok = err.(*github.RateLimitError); ok {
				log.Logger.Error("StoreRepo hit limit error, it's time to change client.", zap.Error(err))

				go func() {
					wg.Add(1)
					defer wg.Done()

					gitClient.Reclaim(client, resp)
				}()

				client = clientManager.Fetch()

				goto repeatStore
			} else if e, ok = err.(*github.AbuseRateLimitError); ok {
				log.Logger.Error("SearchReposByStartTime have triggered an abuse detection mechanism.", zap.Error(err))

				time.Sleep(*e.RetryAfter)
				goto repeatStore
			} else {
				log.Logger.Error("StoreRepo encounter this error, proceed to the next loop.", zap.Error(err))

				continue
			}
		}
	}

	wg.Wait()
	log.Logger.Info("All search and storage tasks have been successful.")

	return
}

// SearchRepos 按条件从 github 搜索库，受 github API 限制，一次请求只能获取 1000 条记录
// GitHub API docs: https://developer.github.com/v3/search/#search-repositories
func searchRepos(client *gitClient.GHClient, query string, opt *github.SearchOptions) ([]github.Repository, *github.Response, string, error) {
	var (
		result []github.Repository
		repos  *github.RepositoriesSearchResult
		resp   *github.Response
		stopAt string
		err    error
	)

	page := 1
	maxPage := math.MaxInt32

	for page <= maxPage {
		opt.Page = page

		repos, resp, err = client.Client.Search.Repositories(context.Background(), query, opt)
		if err != nil {
			goto finish
		}

		maxPage = resp.LastPage
		result = append(result, repos.Repositories...)

		page++
	}

finish:
	stopAt = utility.SplitQuery(query)

	return result, resp, stopAt, err
}

// SearchReposByCreated 按创建时间及其它指定条件搜索库
// queries: 指定库的创建时间
// For example:
//     queries := []string{"\"2008-06-01 .. 2012-09-01\"", "\"2012-09-02 .. 2013-03-01\"", "\"2013-03-02 .. 2013-09-03\"", "\"2013-09-04 .. 2014-03-05\"", "\"2014-03-06 .. 2014-09-07\"", "\"2014-09-08 .. 2015-03-09\"", "\"2015-03-10 .. 2015-09-11\"", "\"2015-09-12 .. 2016-03-13\"", "\"2016-03-14 .. 2016-09-15\"", "\"2016-09-16 .. 2017-03-17\""}
//
// querySeg: 指定除创建时间之外的其它条件
// For example:
//     queryPart := constants.QueryLanguage + ":" + constants.LangLua + " " + constants.QueryCreated + ":"
//
// opt: 为搜索方法指定可选参数
// For example:
//     opt := &github.SearchOptions{
//         Sort:        constants.SortByStars,
//         Order:       constants.OrderByDesc,
//         ListOptions: github.ListOptions{PerPage: 100},
//     }
// GitHub API docs: https://developer.github.com/v3/search/#search-repositories
func SearchReposByCreated(client *gitClient.GHClient, queries []string, querySeg string, opt *github.SearchOptions) ([]github.Repository, *github.Response, string, error) {
	var (
		result, repos []github.Repository
		resp          *github.Response
		stopAt        string
		err           error
	)

	for _, q := range queries {
		query := querySeg + q

		repos, resp, stopAt, err = searchRepos(client, query, opt)
		if err != nil {
			goto finish
		}

		result = append(result, repos...)
	}

finish:
	return result, resp, stopAt, err
}

// SearchReposByStartTime 按指定创建时间、时间间隔及其它条件搜索库
// year、month、day: 从此创建时间开始搜索
// For example：
//     year = 2016 month = time.January day = 1
//     时间格式化只能使用 "2006-01-02 15:04:05" 进行，可将年月日和 时分秒拆开使用
//
// incremental: 以此时间增量搜索，如第一次搜索 1 月份的库，第二次搜索 2 月份的库
// For example:
//     interval = "month"
//
// querySeg: 指定除创建时间之外的其它条件
// For example:
//     queryPart := constants.QueryLanguage + ":" + constants.LangLua + " " + constants.QueryCreated + ":"
//
// opt: 为搜索方法指定可选参数
// For example:
//     opt := &github.SearchOptions{
//         Sort:        constants.SortByStars,
//         Order:       constants.OrderByDesc,
//         ListOptions: github.ListOptions{PerPage: 100},
//     }
// GitHub API docs: https://developer.github.com/v3/search/#search-repositories
func SearchReposByStartTime(client *gitClient.GHClient, year int, month time.Month, day int, incremental, querySeg string, opt *github.SearchOptions) ([]github.Repository, *github.Response, string, error) {
	var (
		result, repos []github.Repository
		resp          *github.Response
		stopAt        string
		err           error
	)

	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	for date.Unix() < time.Now().Unix() {
		var dateFormat string

		switch incremental {
		case constants.Quarter:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 3, 0).Format("2006-01-02")
		case constants.Month:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 1, 0).Format("2006-01-02")
		case constants.Week:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 0, 6).Format("2006-01-02")
		case constants.Day:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 0, 0).Format("2006-01-02")
		default:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 1, 0).Format("2006-01-02")
		}

		query := querySeg + "\"" + dateFormat + "\""

		repos, resp, stopAt, err = searchRepos(client, query, opt)
		if err != nil {
			goto finish
		}

		result = append(result, repos...)

		// 防止触发 GitHub 的滥用检测机制，等待一秒
		time.Sleep(1 * time.Second)

		switch incremental {
		case constants.Quarter:
			date = date.AddDate(0, 3, 1)
		case constants.Month:
			date = date.AddDate(0, 1, 1)
		case constants.Week:
			date = date.AddDate(0, 0, 7)
		case constants.Day:
			date = date.AddDate(0, 0, 1)
		default:
			date = date.AddDate(0, 1, 1)
		}
	}

finish:
	return result, resp, stopAt, err
}
