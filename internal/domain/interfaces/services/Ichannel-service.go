package Iservices

import "social-connector/internal/domain/dto"

type IChannelServices interface {
	WebhookService(webhookDto *dto.InboundResponse)
}
