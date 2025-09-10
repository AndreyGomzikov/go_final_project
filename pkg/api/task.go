package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go1f/pkg/db"
)

const (
	dateFormat = "20060102"
)

type addTaskReq struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func parseDate(s string, now time.Time) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return now.Format(dateFormat), nil
	}
	if s == "today" {
		return now.Format(dateFormat), nil
	}
	if len(s) != 8 {
		return "", fmt.Errorf("invalid date format")
	}
	_, err := time.Parse(dateFormat, s)
	if err != nil {
		return "", fmt.Errorf("invalid date format")
	}
	return s, nil
}

func validRepeat(rep string) bool {
	rep = strings.TrimSpace(rep)
	if rep == "" { return true }
	if rep == "y" { return true }
	if strings.HasPrefix(rep, "d ") {
		parts := strings.Fields(rep)
		if len(parts) != 2 { return false }
		if _, err := strconv.Atoi(parts[1]); err != nil { return false }
		return true
	}
	if strings.HasPrefix(rep, "w ") {
		list := strings.Split(strings.TrimSpace(rep[2:]), ",")
		if len(list) == 0 { return false }
		for _, x := range list {
			x = strings.TrimSpace(x)
			if x == "" { return false }
			n, err := strconv.Atoi(x)
			if err != nil || n < 1 || n > 7 { return false }
		}
		return true
	}
	if strings.HasPrefix(rep, "m ") {
		fields := strings.Fields(strings.TrimSpace(rep[2:]))
		if len(fields) < 1 || len(fields) > 2 { return false }
		days := strings.Split(fields[0], ",")
		for _, d := range days {
			d = strings.TrimSpace(d)
			if d == "" { return false }
			if strings.HasPrefix(d, "-") {
				if d != "-1" && d != "-2" { return false }
				continue
			}
			if _, err := strconv.Atoi(d); err != nil { return false }
		}
		if len(fields) == 2 {
			months := strings.Split(fields[1], ",")
			for _, m := range months {
				m = strings.TrimSpace(m)
				if m == "" { return false }
				n, err := strconv.Atoi(m)
				if err != nil || n < 1 || n > 12 { return false }
			}
		}
		return true
	}
	return false
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req addTaskReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, errInvalidJSON)
			return
		}
		now := time.Now()
		date, err := parseDate(req.Date, now)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error()); return
		}
		title := strings.TrimSpace(req.Title)
		if title == "" {
			writeError(w, http.StatusBadRequest, "title required"); return
		}
		repeat := strings.TrimSpace(req.Repeat)
		if !validRepeat(repeat) {
			writeError(w, http.StatusBadRequest, "invalid repeat"); return
		}
		
	if date < now.Format(dateFormat) {
		if repeat == "" {
			date = now.Format(dateFormat)
		} else {
			if next, err2 := NextDate(now.Format(dateFormat), date, repeat); err2 == nil && next != "" {
				date = next
			} else {
				date = now.Format(dateFormat)
			}
		}
	}
t := &db.Task{Date: date, Title: title, Comment: strings.TrimSpace(req.Comment), Repeat: repeat}
		id, err := db.AddTask(t)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error()); return
		}
		writeJSON(w, http.StatusOK, map[string]string{"id": id})
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, errMissingID); return
		}
		t, err := db.GetTask(id)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error()); return
		}
		writeJSON(w, http.StatusOK, t)
	case http.MethodPut:
		var t db.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, errInvalidJSON); return
		}
		if t.ID == "" {
			writeError(w, http.StatusBadRequest, errMissingID); return
		}
		now := time.Now()
		date, err := parseDate(t.Date, now)
		if err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
		if strings.TrimSpace(t.Title) == "" { writeError(w, http.StatusBadRequest, "title required"); return }
		if !validRepeat(strings.TrimSpace(t.Repeat)) { writeError(w, http.StatusBadRequest, "invalid repeat"); return }
		t.Date = date
		if err := db.UpdateTask(&t); err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
		writeJSON(w, http.StatusOK, map[string]any{})
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusOK, map[string]string{"error": "id required"})
			return
		}
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, http.StatusOK, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
