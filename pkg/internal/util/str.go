package util

import (
	"strconv"
	"strings"
	"time"
)

// 解析传入的 query 字符串，得到最后请求时间
func SplitQuery(query string) string {
	if query == "" {
		return ""
	}

	dateSlice := strings.Split(query, ":")[2]
	dateSeg := strings.Split(dateSlice, " ..")[0]
	dateStr := strings.Split(dateSeg, "\"")[1]

	return dateStr
}

// SplitDate 解析返回的 stopAt 字符串，得到具体年月日
// stopAt:
//     "2006-01-02"
func SplitDate(stopAt string) ([]int, error) {
	dateSlice := strings.Split(stopAt, "-")

	dateInt, err := StrToInt(dateSlice)
	if err != nil {
		return nil, err
	}

	return dateInt, nil
}

// StrToInt 将字符串转换为 int
// dates:
//     ["2006", "01", "02"]
func StrToInt(dates []string) ([]int, error) {
	var datesInt []int

	for _, s := range dates {
		d, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}

		datesInt = append(datesInt, d)
	}

	return datesInt, nil
}

// 将日期增加相应月份，再将结果转换为字符串，以便给 SearchReposByCreated 函数使用
// date: "2006-01-02"
func DateStrInc(date string, month int) (string, error) {
	startAt, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}
	startStr := startAt.Format("2006-01-02")
	stopAt := startAt.AddDate(0, month, 0)
	stopStr := stopAt.Format("2006-01-02")
	dateRange := startStr + " .. " + stopStr

	return dateRange, nil
}
