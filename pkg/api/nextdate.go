package api

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	dateLayout = "20060102"
)

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	now := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if now == "" || date == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(""))
		return
	}

	n, err := NextDate(now, date, repeat)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(""))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(n))
}

func NextDate(now, base, repeat string) (string, error) {
	nowTime, err := time.Parse(dateLayout, now)
	if err != nil {
		return "", errors.New("invalid now")
	}
	start, err := time.Parse(dateLayout, base)
	if err != nil {
		return "", errors.New("invalid date")
	}
	t := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.UTC)
	b := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	threshold := t
	if b.After(threshold) {
		threshold = b
	}

	repeat = strings.TrimSpace(repeat)
	if repeat == "" {
		return "", errors.New("empty repeat")
	}

	if strings.HasPrefix(repeat, "d") {
		parts := strings.Fields(repeat)
		if len(parts) != 2 {
			return "", errors.New("invalid daily")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n <= 0 || n > 400 {
			return "", errors.New("invalid daily")
		}
		deltaDays := int(threshold.Sub(b).Hours() / 24)
		k := deltaDays/n + 1
		cand := b.AddDate(0, 0, k*n)
		return cand.Format(dateLayout), nil
	}

	if repeat == "y" {
		year := threshold.Year()
		cand := adjustYearly(b, year)
		if !cand.After(threshold) {
			year++
			cand = adjustYearly(b, year)
		}
		return cand.Format(dateLayout), nil
	}

	if strings.HasPrefix(repeat, "w") {
		parts := strings.Fields(repeat)
		if len(parts) < 2 {
			return "", errors.New("invalid weekly")
		}
		list := strings.Join(parts[1:], ",")
		dows, err := parseWeekDays(list)
		if err != nil {
			return "", err
		}
		var allow [8]bool
		for _, d := range dows {
			allow[d] = true
		}
		for i := 1; i <= 7; i++ {
			c := threshold.AddDate(0, 0, i)
			if allow[weekday1to7(c.Weekday())] {
				return c.Format(dateLayout), nil
			}
		}
		return "", errors.New("no weekly match")
	}

	if strings.HasPrefix(repeat, "m") {
		parts := strings.Fields(repeat)
		if len(parts) < 2 || len(parts) > 3 {
			return "", errors.New("invalid monthly")
		}
		daySpec, err := parseMonthDays(strings.ReplaceAll(parts[1], " ", ","))
		if err != nil {
			return "", err
		}
		var monthsAllow [13]bool
		anyFilter := false
		if len(parts) == 3 {
			mlist, err := parseMonthList(strings.ReplaceAll(parts[2], " ", ","))
			if err != nil {
				return "", err
			}
			for _, m := range mlist {
				monthsAllow[m] = true
			}
			anyFilter = len(mlist) > 0
		}

		cur := time.Date(threshold.Year(), threshold.Month(), 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 240; i++ {
			monthNum := int(cur.Month())
			if !anyFilter || monthsAllow[monthNum] {
				opts := monthValidDays(cur.Year(), monthNum, daySpec)
				if len(opts) > 0 {
					best := 0
					for _, d := range opts {
						c := time.Date(cur.Year(), cur.Month(), d, 0, 0, 0, 0, time.UTC)
						if c.After(threshold) && (best == 0 || d < best) {
							best = d
						}
					}
					if best != 0 {
						return time.Date(cur.Year(), cur.Month(), best, 0, 0, 0, 0, time.UTC).Format(dateLayout), nil
					}
				}
			}
			cur = time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
		}
		return "", errors.New("no monthly match")
	}

	return "", errors.New("invalid repeat")
}

func adjustYearly(base time.Time, year int) time.Time {
	m := base.Month()
	d := base.Day()
	if m == time.February && d == 29 {
		if !isLeap(year) {
			return time.Date(year, time.March, 1, 0, 0, 0, 0, time.UTC)
		}
	}
	return time.Date(year, m, d, 0, 0, 0, 0, time.UTC)
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func parseMonthDays(s string) ([]int, error) {
	if strings.Contains(s, " ") {
		s = strings.ReplaceAll(s, " ", ",")
	}
	parts := strings.Split(s, ",")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, errors.New("invalid day")
		}
		if strings.HasPrefix(p, "-") {
			if p != "-1" && p != "-2" {
				return nil, errors.New("invalid day")
			}
			v, _ := strconv.Atoi(p)
			res = append(res, v)
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 || v > 31 {
			return nil, errors.New("invalid day")
		}
		res = append(res, v)
	}
	return res, nil
}

func monthValidDays(year int, month int, spec []int) []int {
	last := lastDayOfMonth(year, time.Month(month))
	vals := make([]int, 0, len(spec))
	seen := map[int]bool{}
	for _, v := range spec {
		var d int
		if v > 0 {
			d = v
			if d > last {
				continue
			}
		} else {
			if v == -1 {
				d = last
			} else if v == -2 {
				d = last - 1
			} else {
				continue
			}
		}
		if !seen[d] {
			seen[d] = true
			vals = append(vals, d)
		}
	}
	return vals
}

func lastDayOfMonth(year int, m time.Month) int {
	next := time.Date(year, m+1, 1, 0, 0, 0, 0, time.UTC)
	prev := next.AddDate(0, 0, -1)
	return prev.Day()
}

func parseMonthList(s string) ([]int, error) {
	if strings.Contains(s, " ") {
		s = strings.ReplaceAll(s, " ", ",")
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 || v > 12 {
			return nil, errors.New("invalid month")
		}
		out = append(out, v)
	}
	return out, nil
}

func parseWeekDays(s string) ([]int, error) {
	if strings.Contains(s, " ") {
		s = strings.ReplaceAll(s, " ", ",")
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	seen := map[int]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 || v > 7 {
			return nil, errors.New("invalid weekday")
		}
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("invalid weekday")
	}
	sort.Ints(out)
	return out, nil
}

func weekday1to7(wd time.Weekday) int {
	switch wd {
	case time.Monday:
		return 1
	case time.Tuesday:
		return 2
	case time.Wednesday:
		return 3
	case time.Thursday:
		return 4
	case time.Friday:
		return 5
	case time.Saturday:
		return 6
	default:
		return 7
	}
}
