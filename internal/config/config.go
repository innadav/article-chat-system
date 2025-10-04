package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	DatabaseURL        string
	Port               string
	InitialArticleURLs []string
	LLMProvider        string
	GoogleAPIKey       string
	OpenAIAPIKey       string
	PromptVersion      string
	WeaviateURL        string
	WeaviateAPIKey     string
}

// New loads configuration from environment variables.
func New() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	return &Config{
		DatabaseURL:    GetEnv("DATABASE_URL", "postgres://user:password@localhost:5433/articledb?sslmode=disable"),
		Port:           GetEnv("PORT", "8080"),
		LLMProvider:    GetEnv("LLM_PROVIDER", "openai"),
		GoogleAPIKey:   GetEnv("GEMINI_API_KEY", ""),
		OpenAIAPIKey:   GetEnv("OPENAI_API_KEY", ""),
		PromptVersion:  GetEnv("PROMPT_VERSION", "v1"),
		WeaviateURL:    GetEnv("WEAVIATE_URL", "http://localhost:8080"),
		WeaviateAPIKey: GetEnv("WEAVIATE_API_KEY", ""),
		InitialArticleURLs: []string{
			"https://techcrunch.com/2025/07/26/astronomer-winks-at-viral-notoriety-with-temporary-spokesperson-gwyneth-paltrow/",
			"https://techcrunch.com/2025/07/26/allianz-life-says-majority-of-customers-personal-data-stolen-in-cyberattack/",
			"https://techcrunch.com/2025/07/27/itch-io-is-the-latest-marketplace-to-crack-down-on-adult-games/",
			"https://techcrunch.com/2025/07/26/tesla-vet-says-that-reviewing-real-products-not-mockups-is-the-key-to-staying-innovative/",
			"https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/",
			"https://techcrunch.com/2025/07/26/dating-safety-app-tea-breached-exposing-72000-user-images/",
			"https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/",
			"https://techcrunch.com/2025/07/25/intel-is-spinning-off-its-network-and-edge-group/",
			"https://techcrunch.com/2025/07/27/wizard-of-oz-blown-up-by-ai-for-giant-sphere-screen/",
			"https://techcrunch.com/2025/07/27/doge-has-built-an-ai-tool-to-slash-federal-regulations/",
			"https://edition.cnn.com/2025/07/27/business/us-china-trade-talks-stockholm-intl-hnk",
			"https://edition.cnn.com/2025/07/27/business/trump-us-eu-trade-deal",
			"https://edition.cnn.com/2025/07/27/business/eu-trade-deal",
			"https://edition.cnn.com/2025/07/26/tech/daydream-ai-online-shopping",
			"https://edition.cnn.com/2025/07/25/tech/meta-ai-superintelligence-team-who-its-hiring",
			"https://edition.cnn.com/2025/07/25/tech/sequoia-islamophobia-maguire-mamdani",
			"https://edition.cnn.com/2025/07/24/tech/intel-layoffs-15-percent-q2-earnings",
		},
	}
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
