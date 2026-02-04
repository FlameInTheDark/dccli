package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/cfg"
	"github.com/FlameInTheDark/dccli/pkg/discord"
	"github.com/FlameInTheDark/dccli/pkg/dprint"
)

// Context keys for storing values in context
type contextKey string

const (
	outputFormatKey contextKey = "output_format"
	botConfigKey    contextKey = "bot_config"
	clientKey       contextKey = "discord_client"
	tokenKey        contextKey = "token"
)

// CLIContext holds all the shared context for CLI commands
type CLIContext struct {
	OutputFormat dprint.OutputFormat
	BotConfig    *cfg.BotConfig
	Client       *discord.DiscordClient
	Token        string // Direct token override
	Quiet        bool   // Suppress status messages
}

// NewCLIContext creates a CLIContext from urfave/cli context
func NewCLIContext(c *cli.Command) (*CLIContext, error) {
	ctx := &CLIContext{}

	// Parse quiet flag
	ctx.Quiet = c.Bool("quiet")

	// Parse output format
	formatStr := c.String("output")
	if formatStr == "" {
		formatStr = "table"
	}
	format, err := dprint.ParseFormat(formatStr)
	if err != nil {
		// Log the error but continue with default format
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	ctx.OutputFormat = format

	// Get token override
	ctx.Token = c.String("token")

	// Get bot configuration
	botConfig, err := GetBotConfig(c, ctx.Token)
	if err != nil {
		return nil, err
	}
	ctx.BotConfig = botConfig

	// Create Discord client if we have a token
	if ctx.Token != "" {
		client, err := discord.NewClient(ctx.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to create Discord client: %w", err)
		}
		ctx.Client = client
	} else if botConfig != nil {
		client, err := discord.NewClient(botConfig.Bot.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to create Discord client: %w", err)
		}
		ctx.Client = client
	}

	return ctx, nil
}

// GetBotConfig determines which bot configuration to use
// Priority: 1) --token flag, 2) --bot flag, 3) current from config
func GetBotConfig(c *cli.Command, tokenOverride string) (*cfg.BotConfig, error) {
	// If token is provided directly, we don't need a config
	if tokenOverride != "" {
		return &cfg.BotConfig{
			Name: "override",
			Bot: cfg.Bot{
				Token: tokenOverride,
			},
		}, nil
	}

	config, err := cfg.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Check for --bot flag override
	botName := c.String("bot")
	if botName != "" {
		bot, err := config.GetBotByName(botName)
		if err != nil {
			return nil, fmt.Errorf("bot '%s' not found in config: %w", botName, err)
		}
		return bot, nil
	}

	// Use current bot from config
	return config.GetCurrent()
}

// GetOutputManager creates an OutputManager from CLI context
func (ctx *CLIContext) GetOutputManager() *dprint.OutputManager {
	return dprint.NewOutputManager(dprint.WithFormat(ctx.OutputFormat))
}

// Close cleans up resources (like Discord client connection)
func (ctx *CLIContext) Close() {
	if ctx.Client != nil {
		// Discord client session should be closed
		// This would require adding a Close method to DiscordClient
		// For now, we assume the session will be garbage collected
	}
}

// Context methods for compatibility with standard context.Context

// WithCLIContext adds CLIContext to a standard context
func WithCLIContext(ctx context.Context, cliCtx *CLIContext) context.Context {
	ctx = context.WithValue(ctx, outputFormatKey, cliCtx.OutputFormat)
	ctx = context.WithValue(ctx, botConfigKey, cliCtx.BotConfig)
	ctx = context.WithValue(ctx, clientKey, cliCtx.Client)
	ctx = context.WithValue(ctx, tokenKey, cliCtx.Token)
	return ctx
}

// GetOutputFormatFromContext retrieves output format from context
func GetOutputFormatFromContext(ctx context.Context) dprint.OutputFormat {
	if format, ok := ctx.Value(outputFormatKey).(dprint.OutputFormat); ok {
		return format
	}
	return dprint.FormatTable
}

// GetBotConfigFromContext retrieves bot config from context
func GetBotConfigFromContext(ctx context.Context) *cfg.BotConfig {
	if bot, ok := ctx.Value(botConfigKey).(*cfg.BotConfig); ok {
		return bot
	}
	return nil
}

// GetClientFromContext retrieves Discord client from context
func GetClientFromContext(ctx context.Context) *discord.DiscordClient {
	if client, ok := ctx.Value(clientKey).(*discord.DiscordClient); ok {
		return client
	}
	return nil
}

// GetTokenFromContext retrieves token from context
func GetTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(tokenKey).(string); ok {
		return token
	}
	return ""
}
