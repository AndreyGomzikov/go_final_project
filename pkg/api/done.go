package api

import (
	"net/http"
	"time"

	"go1f/pkg/db"
)

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, errMissingID)
		return
	}
	t, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	
	if t.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{})
		return
	}
now := time.Now()
	next, err := NextDate(now.Format("20060102"), t.Date, t.Repeat)
	if err != nil || next == "" {
		writeError(w, http.StatusBadRequest, "cannot compute next date")
		return
	}
	if err := db.UpdateDate(id, next); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}
