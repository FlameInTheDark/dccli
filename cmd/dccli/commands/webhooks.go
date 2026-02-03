package commands

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func WebhooksCommand() *cli.Command {
	return &cli.Command{
		Name:  "webhooks",
		Usage: "Webhook management",
		Commands: []*cli.Command{
			WebhooksListCommand(),
			WebhooksGetCommand(),
			WebhooksCreateCommand(),
			WebhooksEditCommand(),
			WebhooksDeleteCommand(),
			WebhooksExecuteCommand(),
		},
	}
}

func WebhooksListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List webhooks",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "channel",
				Usage: "Filter by channel ID",
			},
			&cli.StringFlag{
				Name:  "guild",
				Usage: "Filter by guild ID",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			var webhooks []*discordgo.Webhook

			if channelID := c.String("channel"); channelID != "" {
				webhooks, err = cliCtx.Client.GetChannelWebhooks(channelID)
			} else if guildID := c.String("guild"); guildID != "" {
				webhooks, err = cliCtx.Client.GetGuildWebhooks(guildID)
			} else {
				return utils.ValidationError("either --channel or --guild is required")
			}

			if err != nil {
				return utils.DiscordErrorf("failed to list webhooks: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, wh := range webhooks {
					avatar := "No"
					if wh.Avatar != "" {
						avatar = "Yes"
					}
					data = append(data, []string{
						wh.ID,
						wh.Name,
						wh.ChannelID,
						avatar,
					})
				}
				header := []string{"ID", "Name", "Channel ID", "Avatar"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(webhooks); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// WebhooksGetCommand gets webhook details
func WebhooksGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get webhook details",
		ArgsUsage: "[webhook-id]",
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("webhook ID is required")
			}
			webhookID := c.Args().First()

			webhook, err := cliCtx.Client.GetWebhook(webhookID)
			if err != nil {
				return utils.DiscordErrorf("failed to get webhook: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", webhook.ID)
				fmt.Printf("Name: %s\n", webhook.Name)
				fmt.Printf("Token: %s\n", webhook.Token)
				fmt.Printf("Channel ID: %s\n", webhook.ChannelID)
				fmt.Printf("Guild ID: %s\n", webhook.GuildID)
				if webhook.User != nil {
					fmt.Printf("Creator: %s#%s\n", webhook.User.Username, webhook.User.Discriminator)
				}
				fmt.Printf("URL: https://discord.com/api/webhooks/%s/%s\n", webhook.ID, webhook.Token)
			} else {
				if err := output.Print(webhook); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// WebhooksCreateCommand creates a webhook
func WebhooksCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a webhook",
		ArgsUsage: "[channel-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Webhook name",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "avatar",
				Usage: "Path to avatar image file",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("channel ID is required")
			}
			channelID := c.Args().First()
			name := c.String("name")

			var avatar string
			if avatarPath := c.String("avatar"); avatarPath != "" {
				avatarData, err := os.ReadFile(avatarPath)
				if err != nil {
					return utils.ValidationErrorf("failed to read avatar file: %w", err)
				}
				avatar = "data:image/png;base64," + base64.StdEncoding.EncodeToString(avatarData)
			}

			webhook, err := cliCtx.Client.CreateWebhook(channelID, name, avatar)
			if err != nil {
				return utils.DiscordErrorf("failed to create webhook: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Webhook created successfully!\n")
				fmt.Printf("ID: %s\n", webhook.ID)
				fmt.Printf("Name: %s\n", webhook.Name)
				fmt.Printf("Token: %s\n", webhook.Token)
				fmt.Printf("URL: https://discord.com/api/webhooks/%s/%s\n", webhook.ID, webhook.Token)
			} else {
				result := map[string]interface{}{
					"success": true,
					"webhook": webhook,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// WebhooksEditCommand edits a webhook
func WebhooksEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a webhook",
		ArgsUsage: "[webhook-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "New webhook name",
			},
			&cli.StringFlag{
				Name:  "avatar",
				Usage: "Path to new avatar image file",
			},
			&cli.StringFlag{
				Name:  "channel",
				Usage: "Move webhook to this channel ID",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("webhook ID is required")
			}
			webhookID := c.Args().First()

			// Get existing webhook first
			existingWebhook, err := cliCtx.Client.GetWebhook(webhookID)
			if err != nil {
				return utils.DiscordErrorf("failed to get webhook: %w", err)
			}

			name := existingWebhook.Name
			if newName := c.String("name"); newName != "" {
				name = newName
			}

			var avatar string
			if avatarPath := c.String("avatar"); avatarPath != "" {
				avatarData, err := os.ReadFile(avatarPath)
				if err != nil {
					return utils.ValidationErrorf("failed to read avatar file: %w", err)
				}
				avatar = "data:image/png;base64," + base64.StdEncoding.EncodeToString(avatarData)
			}

			channelID := c.String("channel")

			webhook, err := cliCtx.Client.EditWebhook(webhookID, name, avatar, channelID)
			if err != nil {
				return utils.DiscordErrorf("failed to edit webhook: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Webhook updated successfully!\n")
				fmt.Printf("ID: %s\n", webhook.ID)
				fmt.Printf("Name: %s\n", webhook.Name)
				if channelID != "" {
					fmt.Printf("Channel ID: %s\n", webhook.ChannelID)
				}
			} else {
				result := map[string]interface{}{
					"success": true,
					"webhook": webhook,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// WebhooksDeleteCommand deletes a webhook
func WebhooksDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a webhook",
		ArgsUsage: "[webhook-id]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Skip confirmation prompt",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("webhook ID is required")
			}
			webhookID := c.Args().First()

			// Get webhook info for confirmation
			webhook, err := cliCtx.Client.GetWebhook(webhookID)
			if err != nil {
				return utils.DiscordErrorf("failed to get webhook info: %w", err)
			}

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete webhook: %s (ID: %s)\n", webhook.Name, webhookID)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.DeleteWebhook(webhookID); err != nil {
				return utils.DiscordErrorf("failed to delete webhook: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted webhook: %s\n", webhook.Name)
			} else {
				result := map[string]interface{}{
					"success":    true,
					"webhook_id": webhookID,
					"name":       webhook.Name,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// WebhooksExecuteCommand executes a webhook
func WebhooksExecuteCommand() *cli.Command {
	return &cli.Command{
		Name:      "execute",
		Usage:     "Execute a webhook",
		ArgsUsage: "[webhook-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "content",
				Usage: "Message content",
			},
			&cli.StringFlag{
				Name:  "embed-file",
				Usage: "JSON file containing embed data",
			},
			&cli.StringFlag{
				Name:  "username",
				Usage: "Override webhook username",
			},
			&cli.StringFlag{
				Name:  "avatar-url",
				Usage: "Override webhook avatar URL",
			},
			&cli.BoolFlag{
				Name:  "wait",
				Usage: "Wait for message and return it",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("webhook ID is required")
			}
			webhookID := c.Args().First()

			// Get webhook to retrieve token
			webhook, err := cliCtx.Client.GetWebhook(webhookID)
			if err != nil {
				return utils.DiscordErrorf("failed to get webhook: %w", err)
			}

			content := c.String("content")
			if content == "" && c.String("embed-file") == "" {
				return utils.ValidationError("either --content or --embed-file is required")
			}

			params := &discordgo.WebhookParams{
				Content:   content,
				Username:  c.String("username"),
				AvatarURL: c.String("avatar-url"),
			}

			// Load embed from file if provided
			if embedFile := c.String("embed-file"); embedFile != "" {
				data, err := os.ReadFile(embedFile)
				if err != nil {
					return utils.ValidationErrorf("failed to read embed file: %w", err)
				}
				var embed discordgo.MessageEmbed
				if err := json.Unmarshal(data, &embed); err != nil {
					return utils.ValidationErrorf("failed to parse embed file: %w", err)
				}
				params.Embeds = []*discordgo.MessageEmbed{&embed}
			}

			wait := c.Bool("wait")
			message, err := cliCtx.Client.ExecuteWebhook(webhookID, webhook.Token, wait, params)
			if err != nil {
				return utils.DiscordErrorf("failed to execute webhook: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Webhook executed successfully!\n")
				if wait && message != nil {
					fmt.Printf("Message ID: %s\n", message.ID)
				}
			} else {
				result := map[string]interface{}{
					"success": true,
				}
				if wait && message != nil {
					result["message_id"] = message.ID
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
