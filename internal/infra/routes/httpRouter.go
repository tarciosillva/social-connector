package routes

import (
	"encoding/json"
	"net/http"
	"social-connector/internal/infra/handlers"

	"github.com/gorilla/mux"
)

type Routes struct {
	Mux            *mux.Router
	HttpHandler    *handlers.HttpHandlers
	InfobipHandler *handlers.InfobipHandlers
}

func NewRoutes(mux *mux.Router, HttpHandler *handlers.HttpHandlers, InfobipHandler *handlers.InfobipHandlers) *Routes {
	return &Routes{mux, HttpHandler, InfobipHandler}
}

// Estruturas para processar o JSON recebido
type Message struct {
	From string `json:"from"`
	Text string `json:"text"`
}

type WebhookRequest struct {
	Messages []Message `json:"messages"`
}

func (r *Routes) Init() {
	r.Mux.HandleFunc("/webhook", r.HttpHandler.MetaWebhook)
	r.Mux.HandleFunc("/infobip-webhook", r.InfobipHandler.InfoBipWebhook)

	r.Mux.HandleFunc("/healthCheck", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"status": "healthy"}
		json.NewEncoder(w).Encode(response)
	}).Methods(http.MethodGet)
}
