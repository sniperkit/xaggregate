package stats

import (
	"time"

	"github.com/google/go-github/github"
)

func GetStarDateSparklineData(firstStarDate time.Time, starMap map[string]int) []int {
	data := []int{}
	for d := firstStarDate; d.Unix() < time.Now().Unix(); d = d.AddDate(0, 0, 1) {
		count, exist := starMap[string(d.Format("1/2/06"))]
		if exist {
			data = append(data, count)
		} else {
			data = append(data, 0)
		}
	}
	return data
}

func HistogramStarDates(list []*github.Stargazer) map[string]int {
	dupFrequency := make(map[string]int)
	for _, item := range list {
		// check if the item/element exist in the dupFrequency map
		dateStr := string(item.GetStarredAt().Format("1/2/06"))
		_, exist := dupFrequency[dateStr]

		if exist {
			dupFrequency[dateStr]++
		} else {
			dupFrequency[dateStr] = 1
		}
	}
	return dupFrequency
}
