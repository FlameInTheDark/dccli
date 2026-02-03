package cfg

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationResult contains the results of configuration validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []string
}

// ValidateConfig validates the entire configuration
func ValidateConfig(config *Config) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]string, 0),
	}

	if config == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "config",
			Message: "configuration is nil",
		})
		return result
	}

	// Validate bots
	if len(config.Bots) == 0 {
		result.Warnings = append(result.Warnings, "no bots configured")
	} else {
		botNames := make(map[string]bool)
		for i, bot := range config.Bots {
			validateBotConfig(&bot, i, result, botNames)
		}
	}

	// Validate current bot
	if config.CurrentBot != "" {
		if _, err := config.GetBotByName(config.CurrentBot); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "current-bot",
				Message: fmt.Sprintf("current bot '%s' not found in configured bots", config.CurrentBot),
			})
		}
	} else if len(config.Bots) > 0 {
		result.Warnings = append(result.Warnings, "no current bot set, using first bot")
	}

	return result
}

// validateBotConfig validates a single bot configuration
func validateBotConfig(bot *BotConfig, index int, result *ValidationResult, botNames map[string]bool) {
	prefix := fmt.Sprintf("bots[%d]", index)

	// Validate name
	if bot.Name == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   prefix + ".name",
			Message: "bot name is required",
		})
	} else {
		// Check for duplicate names
		if botNames[bot.Name] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   prefix + ".name",
				Message: fmt.Sprintf("duplicate bot name '%s'", bot.Name),
			})
		}
		botNames[bot.Name] = true

		// Validate name format (alphanumeric, hyphens, underscores)
		if !isValidBotName(bot.Name) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("bot name '%s' should only contain alphanumeric characters, hyphens, and underscores", bot.Name))
		}
	}

	// Validate token
	if bot.Bot.Token == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   prefix + ".bot.token",
			Message: "bot token is required",
		})
	} else {
		// Check token format
		tokenValidation := ValidateBotToken(bot.Bot.Token)
		if !tokenValidation.Valid {
			result.Valid = false
			for _, err := range tokenValidation.Errors {
				result.Errors = append(result.Errors, ValidationError{
					Field:   prefix + ".bot.token",
					Message: err.Message,
				})
			}
		}
		for _, warning := range tokenValidation.Warnings {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("bot '%s' token: %s", bot.Name, warning))
		}
	}
}

// TokenValidationResult contains the results of token validation
type TokenValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []string
}

// ValidateBotToken validates a Discord bot token format
func ValidateBotToken(token string) *TokenValidationResult {
	result := &TokenValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]string, 0),
	}

	if token == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "token",
			Message: "token is empty",
		})
		return result
	}

	// Check for common token prefixes
	if strings.HasPrefix(token, "Bot ") {
		result.Warnings = append(result.Warnings, "token includes 'Bot ' prefix which will be added automatically by the library")
	}

	// Discord bot tokens typically have the format: base64.base64.base64
	// or for newer tokens, they may have different formats
	// Basic validation: check length and structure

	// Remove "Bot " prefix if present for validation
	cleanToken := strings.TrimPrefix(token, "Bot ")
	cleanToken = strings.TrimSpace(cleanToken)

	// Check minimum length (Discord tokens are typically quite long)
	if len(cleanToken) < 50 {
		result.Warnings = append(result.Warnings, "token seems shorter than expected for a Discord bot token")
	}

	// Check for common issues
	if strings.Contains(cleanToken, " ") {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "token",
			Message: "token contains spaces",
		})
	}

	if strings.Contains(cleanToken, "\n") || strings.Contains(cleanToken, "\r") {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "token",
			Message: "token contains newline characters",
		})
	}

	// Check for placeholder tokens
	placeholders := []string{"YOUR_TOKEN_HERE", "YOUR_BOT_TOKEN", "PLACEHOLDER", "EXAMPLE", "TOKEN_HERE"}
	upperToken := strings.ToUpper(cleanToken)
	for _, placeholder := range placeholders {
		if strings.Contains(upperToken, placeholder) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "token",
				Message: fmt.Sprintf("token appears to be a placeholder (%s)", placeholder),
			})
			break
		}
	}

	// Basic format check: Discord tokens typically have 2 dots separating 3 parts
	parts := strings.Split(cleanToken, ".")
	if len(parts) != 3 {
		result.Warnings = append(result.Warnings, "token format doesn't match expected Discord token structure (should have 3 parts separated by dots)")
	} else {
		// Validate each part is base64-like
		base64Pattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
		for i, part := range parts {
			if !base64Pattern.MatchString(part) {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("token part %d contains characters not typical for base64 encoding", i+1))
			}
		}
	}

	return result
}

// isValidBotName checks if a bot name contains only valid characters
func isValidBotName(name string) bool {
	// Allow alphanumeric, hyphens, and underscores
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validPattern.MatchString(name)
}

// ValidateConfigFile loads and validates the configuration file
func ValidateConfigFile() (*ValidationResult, error) {
	config, err := LoadConfig()
	if err != nil {
		if err == ErrNotConfigured {
			return &ValidationResult{
				Valid:    false,
				Errors:   []ValidationError{{Field: "config", Message: "configuration file not found"}},
				Warnings: []string{},
			}, nil
		}
		return nil, err
	}

	return ValidateConfig(config), nil
}

// FormatValidationResult formats a validation result as a string
func FormatValidationResult(result *ValidationResult) string {
	var output strings.Builder

	if result.Valid && len(result.Warnings) == 0 {
		output.WriteString("✓ Configuration is valid\n")
		return output.String()
	}

	if result.Valid {
		output.WriteString("✓ Configuration is valid with warnings:\n")
	} else {
		output.WriteString("✗ Configuration is invalid:\n")
	}

	if len(result.Errors) > 0 {
		output.WriteString("\nErrors:\n")
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  - %s\n", err.Error()))
		}
	}

	if len(result.Warnings) > 0 {
		output.WriteString("\nWarnings:\n")
		for _, warning := range result.Warnings {
			output.WriteString(fmt.Sprintf("  - %s\n", warning))
		}
	}

	return output.String()
}
