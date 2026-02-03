package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func MessagesCommand() *cli.Command {
	return &cli.Command{
		Name:  "messages",
		Usage: "Message operations",
		Commands: []*cli.Command{
			MessagesGetCommand(),
			MessagesSendCommand(),
			MessagesEditCommand(),
			MessagesDeleteCommand(),
			MessagesListenCommand(),
			MessagesReactionsCommand(),
		},
	}
}

func MessagesListenCommand() *cli.Command {
	return &cli.Command{
		Name:      "listen",
		Usage:     "Listen for new messages in a channel",
		ArgsUsage: "<channel-id>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "mentions",
				Usage: "Only show messages that mention the bot",
			},
			&cli.StringSliceFlag{
				Name:  "users",
				Usage: "Filter messages by user IDs (comma-separated)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: text, json",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("channel ID is required")
			}
			channelID := c.Args().First()
			onlyMentions := c.Bool("mentions")
			filterUsers := c.StringSlice("users")

			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			// Override output format if specified via flag
			if c.String("format") == "json" {
				cliCtx.OutputFormat = dprint.FormatJSON
			}

			// Get current bot ID for mention check
			me, err := cliCtx.Client.GetCurrentUser()
			if err != nil {
				return utils.DiscordErrorf("failed to get bot user: %w", err)
			}
			botID := me.ID

			output := cliCtx.GetOutputManager()

			fmt.Printf("Listening for messages in channel %s... Press Ctrl+C to stop.\n", channelID)

			type ListenMessageOutput struct {
				Username    string   `json:"username"`
				UserID      string   `json:"user_id"`
				Content     string   `json:"content"`
				Attachments []string `json:"attachments,omitempty"`
			}

			cliCtx.Client.Session().AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
				if m.ChannelID != channelID {
					return
				}

				if len(filterUsers) > 0 {
					found := false
					for _, userID := range filterUsers {
						if m.Author.ID == userID {
							found = true
							break
						}
					}
					if !found {
						return
					}
				}

				if onlyMentions {
					isMentioned := false
					for _, user := range m.Mentions {
						if user.ID == botID {
							isMentioned = true
							break
						}
					}
					if !isMentioned {
						return
					}
				}

				if output.GetFormat() == dprint.FormatJSON {
					msg := ListenMessageOutput{
						Username: m.Author.Username,
						UserID:   m.Author.ID,
						Content:  m.Content,
					}
					for _, att := range m.Attachments {
						msg.Attachments = append(msg.Attachments, att.URL)
					}

					data, err := json.Marshal(msg)
					if err == nil {
						fmt.Println(string(data))
					}
					return
				}

				fmt.Printf("%s (%s): %s\n", m.Author.Username, m.Author.ID, m.Content)
				for _, att := range m.Attachments {
					fmt.Printf("Attachment: %s\n", att.URL)
				}
			})

			// Wait for interrupt signal
			sc := make(chan os.Signal, 1)
			signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
			<-sc

			return nil
		},
	}
}

func MessagesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific message",
		ArgsUsage: "[channel-id] [message-id]",
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 2 {
				return utils.ValidationError("channel ID and message ID are required")
			}
			channelID := c.Args().Get(0)
			messageID := c.Args().Get(1)

			message, err := cliCtx.Client.GetChannelMessage(channelID, messageID)
			if err != nil {
				return utils.DiscordErrorf("failed to get message: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", message.ID)
				fmt.Printf("Channel ID: %s\n", message.ChannelID)
				fmt.Printf("Author: %s#%s (%s)\n", message.Author.Username, message.Author.Discriminator, message.Author.ID)
				fmt.Printf("Content: %s\n", message.Content)
				fmt.Printf("Timestamp: %s\n", message.Timestamp.Format("2006-01-02 15:04:05"))
				if len(message.Embeds) > 0 {
					fmt.Printf("Embeds: %d\n", len(message.Embeds))
				}
				if len(message.Attachments) > 0 {
					fmt.Printf("Attachments: %d\n", len(message.Attachments))
				}
				if len(message.Reactions) > 0 {
					fmt.Printf("Reactions: %d\n", len(message.Reactions))
				}
			} else {
				if err := output.Print(message); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MessagesSendCommand() *cli.Command {
	return &cli.Command{
		Name:      "send",
		Usage:     "Send a message to a channel",
		ArgsUsage: "[channel-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "content",
				Usage: "Message content",
			},
			&cli.StringFlag{
				Name:  "embed-file",
				Usage: "JSON file containing embed data",
			},
			&cli.BoolFlag{
				Name:  "tts",
				Usage: "Send as TTS message",
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
			content := c.String("content")

			if content == "" && c.String("embed-file") == "" {
				return utils.ValidationError("either --content or --embed-file is required")
			}

			msg := &discordgo.MessageSend{
				Content: content,
				TTS:     c.Bool("tts"),
			}

			// Load embed from file if provided
			if embedFile := c.String("embed-file"); embedFile != "" {
				data, err := os.ReadFile(embedFile)
				if err != nil {
					return utils.ValidationErrorf("failed to read embed file: %w", err)
				}
				if err := json.Unmarshal(data, &msg.Embed); err != nil {
					return utils.ValidationErrorf("failed to parse embed file: %w", err)
				}
			}

			message, err := cliCtx.Client.SendChannelMessage(channelID, msg)
			if err != nil {
				return utils.DiscordErrorf("failed to send message: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Message sent successfully!\n")
				fmt.Printf("ID: %s\n", message.ID)
				fmt.Printf("Channel: %s\n", channelID)
			} else {
				result := map[string]interface{}{
					"success":    true,
					"message_id": message.ID,
					"channel_id": channelID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func MessagesEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a message",
		ArgsUsage: "[channel-id] [message-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "content",
				Usage: "New message content",
			},
			&cli.StringFlag{
				Name:  "embed-file",
				Usage: "JSON file containing new embed data",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 2 {
				return utils.ValidationError("channel ID and message ID are required")
			}
			channelID := c.Args().Get(0)
			messageID := c.Args().Get(1)

			if c.String("content") == "" && c.String("embed-file") == "" {
				return utils.ValidationError("either --content or --embed-file is required")
			}

			msg := &discordgo.MessageEdit{
				ID:      messageID,
				Channel: channelID,
			}

			if content := c.String("content"); content != "" {
				msg.Content = &content
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
				msg.Embed = &embed
			}

			message, err := cliCtx.Client.EditChannelMessage(channelID, messageID, msg)
			if err != nil {
				return utils.DiscordErrorf("failed to edit message: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Message edited successfully!\n")
				fmt.Printf("ID: %s\n", message.ID)
			} else {
				result := map[string]interface{}{
					"success":    true,
					"message_id": message.ID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// MessagesDeleteCommand deletes a message
func MessagesDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a message",
		ArgsUsage: "[channel-id] [message-id]",
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

			if c.NArg() < 2 {
				return utils.ValidationError("channel ID and message ID are required")
			}
			channelID := c.Args().Get(0)
			messageID := c.Args().Get(1)

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete message: %s\n", messageID)
				fmt.Print("Are you sure? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.DeleteChannelMessage(channelID, messageID); err != nil {
				return utils.DiscordErrorf("failed to delete message: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted message: %s\n", messageID)
			} else {
				result := map[string]interface{}{
					"success":    true,
					"message_id": messageID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// MessagesReactionsCommand manages message reactions
func MessagesReactionsCommand() *cli.Command {
	return &cli.Command{
		Name:      "reactions",
		Usage:     "Manage message reactions",
		ArgsUsage: "[channel-id] [message-id]",
		Commands: []*cli.Command{
			ReactionsListCommand(),
			ReactionsAddCommand(),
			ReactionsRemoveCommand(),
		},
	}
}

// ReactionsListCommand lists reactions on a message
func ReactionsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List reactions on a message",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "message",
				Usage:    "Message ID",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			channelID := c.String("channel")
			messageID := c.String("message")

			message, err := cliCtx.Client.GetChannelMessage(channelID, messageID)
			if err != nil {
				return utils.DiscordErrorf("failed to get message: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				if len(message.Reactions) == 0 {
					fmt.Println("No reactions on this message")
					return nil
				}

				data := [][]string{}
				for _, r := range message.Reactions {
					data = append(data, []string{
						r.Emoji.Name,
						r.Emoji.ID,
						strconv.Itoa(r.Count),
					})
				}
				header := []string{"Emoji", "Emoji ID", "Count"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(message.Reactions); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ReactionsAddCommand adds a reaction to a message
func ReactionsAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a reaction to a message",
		ArgsUsage: "[emoji]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "message",
				Usage:    "Message ID",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("emoji is required")
			}
			emoji := c.Args().First()
			channelID := c.String("channel")
			messageID := c.String("message")

			if err := cliCtx.Client.AddReaction(channelID, messageID, emoji); err != nil {
				return utils.DiscordErrorf("failed to add reaction: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully added reaction: %s\n", emoji)
			} else {
				result := map[string]interface{}{
					"success":    true,
					"emoji":      emoji,
					"message_id": messageID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ReactionsRemoveCommand removes a reaction from a message
func ReactionsRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Usage:     "Remove a reaction from a message",
		ArgsUsage: "[emoji]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "message",
				Usage:    "Message ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "user",
				Usage: "User ID to remove reaction from (omit for self)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("emoji is required")
			}
			emoji := c.Args().First()
			channelID := c.String("channel")
			messageID := c.String("message")
			userID := c.String("user")

			if err := cliCtx.Client.RemoveReaction(channelID, messageID, emoji, userID); err != nil {
				return utils.DiscordErrorf("failed to remove reaction: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				if userID != "" {
					fmt.Printf("Successfully removed reaction %s from user %s\n", emoji, userID)
				} else {
					fmt.Printf("Successfully removed reaction: %s\n", emoji)
				}
			} else {
				result := map[string]interface{}{
					"success":    true,
					"emoji":      emoji,
					"message_id": messageID,
					"user_id":    userID,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
