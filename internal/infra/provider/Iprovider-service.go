package provider

import "social-connector/internal/domain/dto"

type IWhatsAppProvider interface {
	SendTextMessage(to, message string) error
	SendAudioMessage(to, audioLink string) error
	GenerateOAuth2Token() (*dto.TokenResponse, error)
}
