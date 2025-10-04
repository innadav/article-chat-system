package config_test

import (
	"os"
	"testing"

	"article-chat-system/internal/config"
)

func TestNew(t *testing.T) {
	// Save original environment variables
	originalDBURL := os.Getenv("DATABASE_URL")
	originalPort := os.Getenv("PORT")
	originalLLMProvider := os.Getenv("LLM_PROVIDER")
	originalGoogleAPIKey := os.Getenv("GEMINI_API_KEY")
	originalPromptVersion := os.Getenv("PROMPT_VERSION")

	// Clean up after test
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
		if originalLLMProvider != "" {
			os.Setenv("LLM_PROVIDER", originalLLMProvider)
		} else {
			os.Unsetenv("LLM_PROVIDER")
		}
		if originalGoogleAPIKey != "" {
			os.Setenv("GEMINI_API_KEY", originalGoogleAPIKey)
		} else {
			os.Unsetenv("GEMINI_API_KEY")
		}
		if originalPromptVersion != "" {
			os.Setenv("PROMPT_VERSION", originalPromptVersion)
		} else {
			os.Unsetenv("PROMPT_VERSION")
		}
	}()

	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig *config.Config
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expectedConfig: &config.Config{
				DatabaseURL:   "postgres://user:password@localhost:5433/articledb?sslmode=disable",
				Port:          "8080",
				LLMProvider:   "openai",
				GoogleAPIKey:  "",
				PromptVersion: "v1",
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
			},
		},
		{
			name: "custom values from environment",
			envVars: map[string]string{
				"DATABASE_URL":    "postgres://custom:pass@localhost:5432/testdb",
				"PORT":            "9090",
				"LLM_PROVIDER":    "openai",
				"GEMINI_API_KEY":  "test-api-key",
				"PROMPT_VERSION":  "v2",
			},
			expectedConfig: &config.Config{
				DatabaseURL:   "postgres://custom:pass@localhost:5432/testdb",
				Port:          "9090",
				LLMProvider:   "openai",
				GoogleAPIKey:  "test-api-key",
				PromptVersion: "v2",
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("DB_URL")
			os.Unsetenv("PORT")
			os.Unsetenv("LLM_PROVIDER")
			os.Unsetenv("GEMINI_API_KEY")
			os.Unsetenv("PROMPT_VERSION")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg := config.New()

			if cfg.DatabaseURL != tt.expectedConfig.DatabaseURL {
				t.Errorf("Expected DatabaseURL %s, got %s", tt.expectedConfig.DatabaseURL, cfg.DatabaseURL)
			}

			if cfg.Port != tt.expectedConfig.Port {
				t.Errorf("Expected Port %s, got %s", tt.expectedConfig.Port, cfg.Port)
			}

			if cfg.LLMProvider != tt.expectedConfig.LLMProvider {
				t.Errorf("Expected LLMProvider %s, got %s", tt.expectedConfig.LLMProvider, cfg.LLMProvider)
			}

			if cfg.GoogleAPIKey != tt.expectedConfig.GoogleAPIKey {
				t.Errorf("Expected GoogleAPIKey %s, got %s", tt.expectedConfig.GoogleAPIKey, cfg.GoogleAPIKey)
			}

			if cfg.PromptVersion != tt.expectedConfig.PromptVersion {
				t.Errorf("Expected PromptVersion %s, got %s", tt.expectedConfig.PromptVersion, cfg.PromptVersion)
			}

			if len(cfg.InitialArticleURLs) != len(tt.expectedConfig.InitialArticleURLs) {
				t.Errorf("Expected %d initial URLs, got %d", len(tt.expectedConfig.InitialArticleURLs), len(cfg.InitialArticleURLs))
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "env_value",
			setEnv:       true,
			expected:     "env_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			setEnv:       false,
			expected:     "default",
		},
		{
			name:         "empty environment variable",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			defer os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			result := config.GetEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
