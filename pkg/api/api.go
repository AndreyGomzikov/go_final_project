package api

import (
	"net/http"
)

func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/signin", signinHandler)
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.HandleFunc("/api/task", auth(taskHandler))
	mux.HandleFunc("/api/tasks", auth(tasksHandler))
	mux.HandleFunc("/api/task/done", auth(doneTaskHandler))
}
