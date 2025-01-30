package services

import (
	"fmt"
	"social-connector/internal/domain/dto"
	"social-connector/internal/domain/entities"
	Iservices "social-connector/internal/domain/interfaces/services"
	"social-connector/internal/infra/logger"
	"social-connector/internal/infra/provider"
	"strings"
	"time"
)

type ChannelService struct {
	Logger             *logger.Logger
	UserContextService Iservices.IUserContextService
	QueryAIService     Iservices.IQueryAIService
	WhatsAppProvider   provider.IWhatsAppProvider
}

func NewChannelService(logger *logger.Logger, userContextService Iservices.IUserContextService, queryAIService Iservices.IQueryAIService, whatsAppProvider provider.IWhatsAppProvider) *ChannelService {
	return &ChannelService{Logger: logger, UserContextService: userContextService, QueryAIService: queryAIService, WhatsAppProvider: whatsAppProvider}
}

func (th *ChannelService) WebhookService(webhookDto *dto.InboundResponse) {
	// messageType := webhookDto.Results[webhookDto.MessageCount-1].Message.Type
	// lastMessage := webhookDto.Results[webhookDto.MessageCount-1].Message.Text
	// userAudioUrl := webhookDto.Results[webhookDto.MessageCount-1].Message.Url
	to := webhookDto.Results[webhookDto.MessageCount-1].From

	messagesSplit := []string{
		"Olá! Chegamos ao fim da minha versão beta inicial e agora farei uma pausa para manutenção e melhorias",
		"Agradeço pela compreensão. 😊",
	}

	for i, message := range messagesSplit {
		if strings.TrimSpace(message) != "" {
			if i > 0 {
				time.Sleep(2 * time.Second)
			}
			if err := th.WhatsAppProvider.SendTextMessage(to, message); err != nil {
				th.Logger.Error(fmt.Sprintf("Failed to send WhatsApp message to %s: %s", to, err.Error()))
				return
			}
		}
	}

	return

	// userContext, err := th.UserContextService.FindContext(to)
	// if err != nil {
	// 	th.Logger.Warn(fmt.Sprintf("Context not found for conversation ID %s. Initializing new context.", to))
	// 	userContext = entities.UserContext{
	// 		ConversationID: to,
	// 		Transcript:     []entities.Transcript{},
	// 		Context:        "",
	// 	}

	// 	err := th.UserContextService.Create(userContext)
	// 	if err != nil {
	// 		th.Logger.Error(fmt.Sprintf("Error to create a new context to %s. Err: %v", to, err))
	// 	}
	// }

	// switch messageType {
	// case "TEXT":
	// 	th.processText(lastMessage, userContext, to)
	// case "AUDIO":
	// 	th.processAudio(userAudioUrl, userContext, to)
	// default:
	// 	th.Logger.Warn("Unavailable message type")
	// }
}

func (cs *ChannelService) processText(lastMessage string, userContext entities.UserContext, to string) {
	userContext.Transcript = append(userContext.Transcript, entities.Transcript{
		Role:      "user",
		Message:   lastMessage,
		Timestamp: time.Now(),
	})

	result, err := cs.QueryAIService.ExecuteQueryAI(lastMessage, userContext.Context)
	if err != nil {
		cs.Logger.Error(fmt.Sprintf("Failed to execute AI query: %s", err.Error()))
		return
	}

	userContext.Transcript = append(userContext.Transcript, entities.Transcript{
		Role:      "agent",
		Message:   result.Response,
		Timestamp: time.Now(),
	})

	if len(userContext.Transcript) > 1 {
		lastAgentMessage := userContext.Transcript[len(userContext.Transcript)-1].Message
		userContext.Context = lastAgentMessage
	} else {
		userContext.Context = result.Response
	}

	userContext.UpdatedAt = time.Now()
	if _, err := cs.UserContextService.UpdateUserContext(to, userContext); err != nil {
		cs.Logger.Error(fmt.Sprintf("Failed to update user context: %s", err.Error()))
		return
	}

	messagesSplit := strings.Split(result.Response, ".")

	cs.Logger.Info(fmt.Sprintf("Sending AI response messages to WhatsApp number: %s", to))
	for i, message := range messagesSplit {
		if strings.TrimSpace(message) != "" {
			if i > 0 {
				time.Sleep(2 * time.Second)
			}
			if err := cs.WhatsAppProvider.SendTextMessage(to, message); err != nil {
				cs.Logger.Error(fmt.Sprintf("Failed to send WhatsApp message to %s: %s", to, err.Error()))
				return
			}
		}
	}
}

func (cs *ChannelService) processAudio(userAudioUrl string, userContext entities.UserContext, to string) error {
	authToken, err := cs.WhatsAppProvider.GenerateOAuth2Token()
	if err != nil {
		cs.Logger.Error(fmt.Sprintf("Error to generate OAuth2 token %v", err))
		return err
	}

	result, err := cs.QueryAIService.ExecuteAudioQueryAI(userAudioUrl, authToken.AccessToken, userContext.Context)
	if err != nil {
		cs.Logger.Error(fmt.Sprintf("Failed to execute AI query: %v", err))
		return err
	}

	userContext.Transcript = append(userContext.Transcript, entities.Transcript{
		Role:      "user",
		Message:   result.QueryText,
		Audio:     userAudioUrl,
		Timestamp: time.Now(),
	})

	userContext.Transcript = append(userContext.Transcript, entities.Transcript{
		Role:      "agent",
		Message:   result.Response,
		Audio:     result.AudioLink,
		Timestamp: time.Now(),
	})

	if len(userContext.Transcript) > 1 {
		lastAgentMessage := userContext.Transcript[len(userContext.Transcript)-1].Message
		userContext.Context = lastAgentMessage
	} else {
		userContext.Context = result.Response
	}

	userContext.UpdatedAt = time.Now()
	if _, err := cs.UserContextService.UpdateUserContext(to, userContext); err != nil {
		cs.Logger.Error(fmt.Sprintf("Failed to update user context: %s", err.Error()))
		return err
	}

	cs.Logger.Info(fmt.Sprintf("Sending AI response audio message to WhatsApp number: %s", to))
	if err := cs.WhatsAppProvider.SendAudioMessage(to, result.AudioLink); err != nil {
		cs.Logger.Error(fmt.Sprintf("Failed to send WhatsApp audio message to %s: %s", to, err.Error()))
		return err
	}

	return nil
}
