package server

import (
	"fmt"
	"log"
	"net/http"

	"go1f/pkg/api"
)

func Run(port int, cfg api.Config) error {
    mux := http.NewServeMux()

    api.Init(mux, cfg) // передаём конфиг в Init

    mux.Handle("/", http.FileServer(http.Dir("web")))

    addr := fmt.Sprintf(":%d", port)
    log.Printf("Server is starting on http://localhost:%d\n", port)

    return http.ListenAndServe(addr, mux)
}
