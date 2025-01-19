package handlers

import (
	"encoding/json"
	"net/http"
	"social-connector/internal/domain/dto"
	Iservices "social-connector/internal/domain/interfaces/services"
	"social-connector/internal/infra/logger"
)

type InfobipHandlers struct {
	Logger         *logger.Logger
	ChannelService Iservices.IChannelServices
}

func NewInfobipHandlers(logger *logger.Logger, channelService Iservices.IChannelServices) *InfobipHandlers {
	return &InfobipHandlers{Logger: logger, ChannelService: channelService}
}

func (th *InfobipHandlers) InfoBipWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Mwtod not allowed", http.StatusMethodNotAllowed)
		return
	}

	var webhookRequest *dto.InboundResponse
	err := json.NewDecoder(r.Body).Decode(&webhookRequest)
	if err != nil {
		http.Error(w, "Error to process JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	go th.ChannelService.WebhookService(webhookRequest)

	w.WriteHeader(http.StatusOK)
}
