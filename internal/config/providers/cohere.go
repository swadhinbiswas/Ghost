// cohere.go: Cohere provider (trial credits).
//
// Base URL: https://api.cohere.com/v1
// Auth: Bearer $COHERE_API_KEY
// Free: trial credits, then pay-as-you-go.
package providers

import "charm.land/catwalk/pkg/catwalk"

var CohereProvider = catwalk.Provider{
	ID:                  "cohere",
	Name:                "Cohere (trial credits)",
	APIEndpoint:         "https://api.cohere.com/v1",
	APIKey:              "$COHERE_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "command-a-03-2025",
	DefaultSmallModelID: "command-r",
	Models: []catwalk.Model{
		{ID: "command-a-03-2025", Name: "Command A 03 2025 (Cohere)", ContextWindow: 256_000, DefaultMaxTokens: 8_000},
		{ID: "command-r-plus", Name: "Command R Plus (Cohere)", ContextWindow: 128_000, DefaultMaxTokens: 4_000},
		{ID: "command-r", Name: "Command R (Cohere, free trial)", ContextWindow: 128_000, DefaultMaxTokens: 4_000},
		{ID: "command-light", Name: "Command Light (Cohere, free)", ContextWindow: 4_000, DefaultMaxTokens: 4_000},
		{ID: "embed-english-v3.0", Name: "Embed English v3 (Cohere)", ContextWindow: 512, DefaultMaxTokens: 512},
		{ID: "embed-multilingual-v3.0", Name: "Embed Multilingual v3 (Cohere)", ContextWindow: 512, DefaultMaxTokens: 512},
	},
}
