package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/cfg"
	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func BotCommand() *cli.Command {
	return &cli.Command{
		Name:        "bot",
		Usage:       "Bot configuration management",
		Description: "Manage bot configurations (add, remove, set, list, edit)",
		Commands: []*cli.Command{
			botAddCommand(),
			botRemoveCommand(),
			botSetCommand(),
			botListCommand(),
			botEditCommand(),
		},
	}
}

func botSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set bot from config by its name",
		ArgsUsage: "[bot name]",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("specify bot config name")
			}

			config, err := cfg.LoadConfig()
			if err != nil {
				return utils.ConfigErrorf("failed to load config: %w", err)
			}

			bot, err := config.GetBotByName(c.Args().Get(0))
			if err != nil {
				return utils.NotFoundErrorf("bot '%s' not found", c.Args().Get(0))
			}

			config.CurrentBot = bot.Name
			cfg.SaveConfig(config)

			outputFormat := c.String("output")
			if outputFormat == "" {
				outputFormat = "table"
			}

			if outputFormat == "table" {
				fmt.Printf("Bot '%s' set as current\n", bot.Name)
			} else {
				format, _ := dprint.ParseFormat(outputFormat)
				output := dprint.NewOutputManager(dprint.WithFormat(format))
				result := map[string]interface{}{
					"success":     true,
					"current_bot": bot.Name,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func botAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add bot to config",
		ArgsUsage: "[bot name] [token]",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 2 {
				return utils.ValidationError("specify bot config name and token")
			}

			config, err := cfg.LoadConfig()
			if err != nil {
				switch err {
				case cfg.ErrNotConfigured:
					config = &cfg.Config{
						CurrentBot: c.Args().Get(0),
					}
				default:
					return utils.ConfigErrorf("failed to load config: %w", err)
				}
			}

			for _, bot := range config.Bots {
				if bot.Name == c.Args().Get(0) {
					return utils.ValidationErrorf("bot '%s' already exists", c.Args().Get(0))
				}
			}

			config.Bots = append(config.Bots, cfg.BotConfig{
				Name: c.Args().Get(0),
				Bot: cfg.Bot{
					Token: c.Args().Get(1),
				},
			})
			cfg.SaveConfig(config)

			outputFormat := c.String("output")
			if outputFormat == "" {
				outputFormat = "table"
			}

			if outputFormat == "table" {
				fmt.Printf("Bot '%s' saved successfully\n", c.Args().Get(0))
			} else {
				format, _ := dprint.ParseFormat(outputFormat)
				output := dprint.NewOutputManager(dprint.WithFormat(format))
				result := map[string]interface{}{
					"success":  true,
					"bot_name": c.Args().Get(0),
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func botListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Returns a list of configured bots",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "tokens",
				Usage: "Show full tokens",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			config, err := cfg.LoadConfig()
			if err != nil {
				return utils.ConfigErrorf("failed to load config: %w", err)
			}

			outputFormat := c.String("output")
			if outputFormat == "" {
				outputFormat = "table"
			}

			format, parseErr := dprint.ParseFormat(outputFormat)
			if parseErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", parseErr)
				format = dprint.FormatTable
			}

			if format == dprint.FormatTable {
				data := [][]string{}
				for _, bot := range config.Bots {
					var token, selected string
					if c.Bool("tokens") {
						token = bot.Bot.Token
					} else {
						if len(bot.Bot.Token) > 5 {
							token = bot.Bot.Token[:5] + "*****"
						}
					}
					if bot.Name == config.CurrentBot {
						selected = "*"
					}
					data = append(data, []string{selected, bot.Name, token})
				}

				header := []string{"*", "Name", "Token"}
				dprint.Table(header, data)
			} else {
				type BotInfo struct {
					Name    string `json:"name" yaml:"name"`
					Token   string `json:"token,omitempty" yaml:"token,omitempty"`
					Current bool   `json:"current" yaml:"current"`
				}

				var bots []BotInfo
				for _, bot := range config.Bots {
					info := BotInfo{
						Name:    bot.Name,
						Current: bot.Name == config.CurrentBot,
					}
					if c.Bool("tokens") {
						info.Token = bot.Bot.Token
					}
					bots = append(bots, info)
				}

				output := dprint.NewOutputManager(dprint.WithFormat(format))
				if err := output.Print(bots); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func botRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a bot from config",
		ArgsUsage: "[bot-name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Skip confirmation prompt",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("bot name is required")
			}

			botName := c.Args().First()

			config, err := cfg.LoadConfig()
			if err != nil {
				return utils.ConfigErrorf("failed to load config: %w", err)
			}

			botIndex := -1
			for i, bot := range config.Bots {
				if bot.Name == botName {
					botIndex = i
					break
				}
			}

			if botIndex == -1 {
				return utils.NotFoundErrorf("bot '%s' not found", botName)
			}

			if !c.Bool("force") {
				fmt.Printf("You are about to remove bot: %s\n", botName)
				fmt.Print("Are you sure? (yes/no): ")
				var response string
				_, err := fmt.Scanln(&response)
				if err == nil && response != "yes" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			config.Bots = append(config.Bots[:botIndex], config.Bots[botIndex+1:]...)

			if config.CurrentBot == botName {
				if len(config.Bots) > 0 {
					config.CurrentBot = config.Bots[0].Name
				} else {
					config.CurrentBot = ""
				}
			}

			cfg.SaveConfig(config)

			outputFormat := c.String("output")
			if outputFormat == "" {
				outputFormat = "table"
			}

			if outputFormat == "table" {
				fmt.Printf("Bot '%s' removed successfully\n", botName)
			} else {
				format, _ := dprint.ParseFormat(outputFormat)
				output := dprint.NewOutputManager(dprint.WithFormat(format))
				result := map[string]interface{}{
					"success":  true,
					"bot_name": botName,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func botEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a bot's configuration",
		ArgsUsage: "[bot-name]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "token",
				Usage: "New bot token",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "New bot name",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("bot name is required")
			}

			botName := c.Args().First()
			newToken := c.String("token")
			newName := c.String("name")

			if newToken == "" && newName == "" {
				return utils.ValidationError("at least one of --token or --name must be specified")
			}

			config, err := cfg.LoadConfig()
			if err != nil {
				return utils.ConfigErrorf("failed to load config: %w", err)
			}

			botIndex := -1
			for i, bot := range config.Bots {
				if bot.Name == botName {
					botIndex = i
					break
				}
			}

			if botIndex == -1 {
				return utils.NotFoundErrorf("bot '%s' not found", botName)
			}

			if newToken != "" {
				config.Bots[botIndex].Bot.Token = newToken
			}
			if newName != "" {
				if newName != botName {
					for _, bot := range config.Bots {
						if bot.Name == newName {
							return utils.ValidationErrorf("bot with name '%s' already exists", newName)
						}
					}
				}
				if config.CurrentBot == botName {
					config.CurrentBot = newName
				}
				config.Bots[botIndex].Name = newName
			}

			cfg.SaveConfig(config)

			outputFormat := c.String("output")
			if outputFormat == "" {
				outputFormat = "table"
			}

			if outputFormat == "table" {
				fmt.Printf("Bot '%s' updated successfully\n", botName)
			} else {
				format, _ := dprint.ParseFormat(outputFormat)
				output := dprint.NewOutputManager(dprint.WithFormat(format))
				result := map[string]interface{}{
					"success":  true,
					"bot_name": botName,
				}
				if newName != "" {
					result["new_name"] = newName
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func ConfigValidateCommand() *cli.Command {
	return &cli.Command{
		Name:        "validate",
		Usage:       "Validate configuration file",
		Description: "Check configuration file for errors and warnings",
		Action: func(ctx context.Context, c *cli.Command) error {
			result, err := cfg.ValidateConfigFile()
			if err != nil {
				return utils.ConfigErrorf("failed to validate config: %w", err)
			}

			outputFormat := c.String("output")
			if outputFormat == "" {
				outputFormat = "table"
			}

			if outputFormat == "table" {
				fmt.Print(cfg.FormatValidationResult(result))
			} else {
				format, _ := dprint.ParseFormat(outputFormat)
				output := dprint.NewOutputManager(dprint.WithFormat(format))

				errorStrings := make([]string, len(result.Errors))
				for i, err := range result.Errors {
					errorStrings[i] = err.Error()
				}

				resultMap := map[string]interface{}{
					"valid":    result.Valid,
					"errors":   errorStrings,
					"warnings": result.Warnings,
				}
				if err := output.Print(resultMap); err != nil {
					return err
				}
			}

			if !result.Valid {
				return utils.ValidationError("configuration validation failed")
			}

			return nil
		},
	}
}