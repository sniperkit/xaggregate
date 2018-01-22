package github

/*
import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/anacrolix/sync"
	"github.com/google/go-github/github"

	"github.com/sniperkit/xtask/plugin/aggregate/service/github/constant"
)

func searchRepos(g *Github, query string, opts *Options) ([]github.Repository, *github.Response, string, error) {

	var (
		result []github.Repository
		repos  *github.RepositoriesSearchResult
		resp   *github.Response
		stopAt string
		err    error
	)

	page := 1
	maxPage := math.MaxInt32

	ghOptions := &github.SearchOptions{
		// Page: opts.Page,
		// Sort:      "updated",
		// Direction: "desc", // desc
		ListOptions: github.ListOptions{
			Page:    opts.Page,
			PerPage: opts.PerPage,
		},
	}

	for page <= maxPage {
		ghOptions.Page = page
		ghOptions.ListOptions.PerPage = page

		repos, resp, err = g.client.Search.Repositories(context.Background(), query, ghOptions)
		if err != nil {
			goto finish
		}

		maxPage = resp.LastPage
		result = append(result, repos.Repositories...)

		page++
	}

finish:
	stopAt = SplitQuery(query)

	return result, resp, stopAt, err
}

// SearchReposByStartTime Search the library for the specified time, interval, and other conditions
// year, month, day: Start searching from this time
// For example:
// year = 2016 month = time.January day = 1
// time format can only be used "2006-01-02 15:04:05", you can use the year, month, day and hour, minute, second
//
// incremental: Incremental search in this time, such as the first search for the library in January, the second search for the library in February
// For example:
// interval = "month"
//
// querySeg: specify conditions other than the creation time
// For example:
// queryPart: = constants.QueryLanguage + ":" + constants.LangLua + "" + constants.QueryCreated + ":"
//
// opt: Specifies optional parameters for the search method
// For example:
// opt: = & github.SearchOptions {
// Sort: constant.SortByStars,
// Order: constant.OrderByDesc,
// ListOptions: github.ListOptions {PerPage: 100},
//}
// GitHub API docs: https://developer.github.com/v3/search/#search-repositories
func SearchReposByStartTime(g *Github, year int, month time.Month, day int, incremental, querySeg string, opts *Options) ([]github.Repository, *github.Response, string, error) {
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
		case constant.Quarter:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 3, 0).Format("2006-01-02")
		case constant.Month:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 1, 0).Format("2006-01-02")
		case constant.Week:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 0, 6).Format("2006-01-02")
		case constant.Day:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 0, 0).Format("2006-01-02")
		default:
			dateFormat = date.Format("2006-01-02") + " .. " + date.AddDate(0, 1, 0).Format("2006-01-02")
		}

		query := querySeg + "\"" + dateFormat + "\""

		repos, resp, stopAt, err = searchRepos(g, query, opts)
		if err != nil {
			goto finish
		}

		result = append(result, repos...)

		// Prevent misuse detection mechanism that triggers GitHub and wait for one second
		time.Sleep(1 * time.Second)

		switch incremental {
		case constant.Quarter:
			date = date.AddDate(0, 3, 1)
		case constant.Month:
			date = date.AddDate(0, 1, 1)
		case constant.Week:
			date = date.AddDate(0, 0, 7)
		case constant.Day:
			date = date.AddDate(0, 0, 1)
		default:
			date = date.AddDate(0, 1, 1)
		}
	}

finish:
	return result, resp, stopAt, err
}

func SearchRepos(g *Github, opts *Options, year int, month time.Month, day int, incremental, querySeg string) { // , opt *github.SearchOptions) {
	var (
		client  *Github
		ok      bool
		wg      sync.WaitGroup
		e       *github.AbuseRateLimitError
		newDate []int
		result  []github.Repository
	)

	defer g.manager.Shutdown()
	client = g.manager.Fetch()

search:

	repos, resp, stopAt, err := SearchReposByStartTime(client, year, month, day, incremental, querySeg, opts)
	result = append(result, repos...)

	if err != nil {
		if _, ok = err.(*github.RateLimitError); ok {
			log.Errorln("SearchReposByStartTime hit limit error, it's time to change client.", err.Error())

			goto changeClient
		} else if e, ok = err.(*github.AbuseRateLimitError); ok {
			log.Errorln("SearchReposByStartTime have triggered an abuse detection mechanism.", err.Error())

			time.Sleep(*e.RetryAfter)
			goto search
		} else if strings.Contains(err.Error(), "timeout") {
			log.Info("SearchReposByStartTime has encountered a timeout error. Sleep for five minutes.")
			time.Sleep(5 * time.Minute)

			goto search
		} else {
			log.Errorln("SearchRepos terminated because of this error.", err.Error())

			return
		}
	} //else {
	//goto store
	//}

changeClient:
	{
		go func() {
			wg.Add(1)
			defer wg.Done()

			// client =
			Reclaim(g, resp)
			// g.Reclaim(resp)
			// g.Reclaim(g, resp)
		}()

		client = g.manager.Fetch()

		if stopAt != "" {
			newDate, err = SplitDate(stopAt)
			if err != nil {
				log.Errorln("SplitDate returned error.", err.Error())

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

		log.Info("stopAt is empty string, stop searching.")
	}



	wg.Wait()
	log.Info("All search and storage tasks have been successful.")

	return
}

*/

/*
   store:
   	log.Info("Start storing repositories now.")
   	for _, repo := range result {
   	repeatStore:
   		err = StoreRepo(&repo, client)
   		if err != nil {
   			if _, ok = err.(*github.RateLimitError); ok {
   				log.Errorln("StoreRepo hit limit error, it's time to change client.", err.Error())

   				go func() {
   					wg.Add(1)
   					defer wg.Done()

   					Reclaim(g, resp)
   					// g.Reclaim(resp)
   					// g.Reclaim(g, resp)
   				}()

   				g = g.manager.Fetch()

   				goto repeatStore
   			} else if e, ok = err.(*github.AbuseRateLimitError); ok {
   				log.Errorln("SearchReposByStartTime have triggered an abuse detection mechanism.", err.Error())

   				time.Sleep(*e.RetryAfter)
   				goto repeatStore
   			} else {
   				log.Errorln("StoreRepo encounter this error, proceed to the next loop.", err.Error())

   				continue
   			}
   		}
   	}
*/
