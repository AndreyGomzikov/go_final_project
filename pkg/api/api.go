package api

import (
	"net/http"
)

func Init(mux *http.ServeMux, cfg Config) {
	auth := AuthMiddleware(cfg)

	mux.Handle("/api/signin", SigninHandler(cfg))
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.HandleFunc("/api/task", auth(taskHandler))
	mux.HandleFunc("/api/tasks", auth(tasksHandler))
	mux.HandleFunc("/api/task/done", auth(doneTaskHandler))
}
