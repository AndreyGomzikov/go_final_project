package api

import (
	"net/http"

	"go1f/pkg/db"
)

const (
	defaultTasksLimit = 50
)

type tasksResponse struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	tasks, err := db.GetTasks(defaultTasksLimit, search)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, tasksResponse{Tasks: []*db.Task{}})
		return
	}
	if tasks == nil {
		tasks = []*db.Task{}
	}
	writeJSON(w, http.StatusOK, tasksResponse{Tasks: tasks})
}
