package github

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

func FormatDate(date time.Time) string {
	return date.Format(DateFormat)
}

func ParseDate(date string) (time.Time, error) {
	return time.Parse(DateFormat, date)
}

func escapeSearch(s string) string {
	return strings.Replace(s, " ", "+", -1)
}

// Parse the incoming query string to get the last request time
func SplitQuery(query string) string {
	if query == "" {
		return ""
	}

	dateSlice := strings.Split(query, ":")[2]
	dateSeg := strings.Split(dateSlice, " ..")[0]
	dateStr := strings.Split(dateSeg, "\"")[1]

	return dateStr
}

// SplitDate parses the returned stopAt string to get the date
// stopAt:
// "2006-01-02"
func SplitDate(stopAt string) ([]int, error) {
	dateSlice := strings.Split(stopAt, "-")

	dateInt, err := StrToInt(dateSlice)
	if err != nil {
		return nil, err
	}

	return dateInt, nil
}

// StrToInt will string converted to int
// dates:
// ["2006", "01", "02"]
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

// Add the date to the corresponding month, then convert the result to a string for use with the SearchReposByCreated function
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

// formatTimestamp formats a github.Timestamp to a string suitable to use
// as a timestamp with timezone PostgreSQL data type
func formatTimestamp(timeStamp *github.Timestamp) string {
	timeFormat := time.RFC3339
	if timeStamp == nil {
		log.Error("'timeStamp' arg given is nil")
		t := time.Time{}
		return t.Format(timeFormat)
	}
	return timeStamp.Format(timeFormat)
}

func isStatus2XX(status int) bool {
	return status > 199 && status < 300
}

func randIntMapKey(m map[int]bool) int {
	defer funcTrack(time.Now())

	i := rand.Intn(len(m))
	for k, v := range m {
		if !v {
			if i == 0 {
				return k
			}
			i--
		}
	}
	return randIntMapKey(m)
}

func random(min, max int) int {
	defer funcTrack(time.Now())

	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func Uint64ToBytes(value uint64, buffer []byte) {
	mask := uint64(0xff)

	var b byte
	v := value
	for i := 0; i < 8; i++ {
		b = byte(v & mask)
		buffer[i] = b
		v = v >> 8
	}
}

func BytesToUint64(buffer []byte) uint64 {
	var v uint64

	v = uint64(buffer[7])
	for i := 6; i >= 0; i-- {
		v = v<<8 + uint64(buffer[i])
	}
	return v
}

func Int64ToBytes(value int64, buffer []byte) {
	mask := int64(0xff)

	var b byte
	v := value
	for i := 0; i < 8; i++ {
		b = byte(v & mask)
		buffer[i] = b
		v = v >> 8
	}
}

func BytesToInt64(buffer []byte) int64 {
	var v int64

	v = int64(buffer[7])
	for i := 6; i >= 0; i-- {
		v = v<<8 + int64(buffer[i])
	}
	return v
}

func int64ToBytes(value int64) []byte {
	buffer := make([]byte, 8)
	mask := int64(0xff)
	var b byte
	v := value
	for i := 0; i < 8; i++ {
		b = byte(v & mask)
		buffer[i] = b
		v = v >> 8
	}
	// Int64ToBytes(*val, buffer)
	return buffer
}

func toBytes(input string) []byte {
	return []byte(input)
}

func mapToString(input map[string]interface{}) string {
	return toString(input)
}

func toString(obj interface{}) string {
	return fmt.Sprintf("%v", obj)
}
