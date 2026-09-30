package ai

import (
	"context"
	"fmt"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"google.golang.org/genai"
)

func NewGenAICli(ctx context.Context, cfg *config.Config, log *logger.Logger) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiApiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	log.Info("Gemini client initialized successfully")
	return client, nil
}
