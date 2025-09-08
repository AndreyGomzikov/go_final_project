package server

import (
	"fmt"
	"net/http"

	"go1f/pkg/api"
)

func Run(port int) error {
	mux := http.NewServeMux()

	api.Init(mux)

	mux.Handle("/", http.FileServer(http.Dir("web")))

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
