package routes

import (
	"encoding/json"
	"net/http"
	"social-connector/internal/infra/handlers"

	"github.com/gorilla/mux"
)

type Routes struct {
	Mux         *mux.Router
	HttpHandler *handlers.HttpHandlers
}

func NewRoutes(mux *mux.Router, HttpHandler *handlers.HttpHandlers) *Routes {
	return &Routes{mux, HttpHandler}
}

func (r *Routes) Init() {
	r.Mux.HandleFunc("/webhook", r.HttpHandler.Webhook)

	r.Mux.HandleFunc("/healthCheck", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"status": "healthy"}
		json.NewEncoder(w).Encode(response)
	}).Methods(http.MethodGet)
}
