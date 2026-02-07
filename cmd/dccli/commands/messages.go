package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
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
			MessagesValidateEmbedCommand(),
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
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Stop listening after receiving N messages",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("channel ID is required")
			}
			channelID := c.Args().First()
			onlyMentions := c.Bool("mentions")
			filterUsers := c.StringSlice("users")
			limit := c.Int("limit")

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

			if !cliCtx.Quiet {
				if limit > 0 {
					fmt.Printf("Listening for %d messages in channel %s... Press Ctrl+C to stop.\n", limit, channelID)
				} else {
					fmt.Printf("Listening for messages in channel %s... Press Ctrl+C to stop.\n", channelID)
				}
			}

			type ListenMessageOutput struct {
				Username    string   `json:"username"`
				UserID      string   `json:"user_id"`
				Content     string   `json:"content"`
				Attachments []string `json:"attachments,omitempty"`
			}

			done := make(chan struct{})
			count := 0

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
				} else {
					fmt.Printf("%s (%s): %s\n", m.Author.Username, m.Author.ID, m.Content)
					for _, att := range m.Attachments {
						fmt.Printf("Attachment: %s\n", att.URL)
					}
				}

				if limit > 0 {
					count++
					if count >= limit {
						// Use a non-blocking send or close check to avoid panic if called multiple times rapidly
						select {
						case <-done:
						default:
							close(done)
						}
					}
				}
			})

			// Wait for interrupt signal or limit reached
			sc := make(chan os.Signal, 1)
			signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

			select {
			case <-sc:
				// Signal received
			case <-done:
				// Limit reached
			}

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
				Name:  "embed",
				Usage: "JSON string containing embed data",
			},
			&cli.StringFlag{
				Name:  "embed-file",
				Usage: "JSON file containing embed data",
			},
			&cli.StringSliceFlag{
				Name:  "file",
				Usage: "Attach file(s) to the message",
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
			embedJSON := c.String("embed")
			embedFile := c.String("embed-file")
			files := c.StringSlice("file")

			if content == "" && embedJSON == "" && embedFile == "" && len(files) == 0 {
				return utils.ValidationError("either --content, --embed, --embed-file, or --file is required")
			}

			msg := &discordgo.MessageSend{
				Content: content,
				TTS:     c.Bool("tts"),
			}

			// Handle file attachments
			var fileHandles []*os.File
			if len(files) > 0 {
				for _, file := range files {
					f, err := os.Open(file)
					if err != nil {
						// Close already opened files
						for _, fh := range fileHandles {
							fh.Close()
						}
						return utils.ValidationErrorf("failed to open file %s: %w", file, err)
					}
					fileHandles = append(fileHandles, f)
					msg.Files = append(msg.Files, &discordgo.File{
						Name:   filepath.Base(file),
						Reader: f,
					})
				}
			}
			// Ensure files are closed after sending
			defer func() {
				for _, f := range fileHandles {
					f.Close()
				}
			}()

			// Load embed from string or file
			if embedJSON != "" {
				if err := json.Unmarshal([]byte(embedJSON), &msg.Embed); err != nil {
					// Try to reconstruct JSON if it was split by shell
					if c.Args().Len() > 0 {
						if _, ok := utils.ReconstructJSON(embedJSON, c.Args().Slice(), &msg.Embed); ok {
							fmt.Println("Warning: JSON argument appeared to be split by shell. Successfully reconstructed.")
						} else {
							return utils.ValidationErrorf("failed to parse embed JSON: %w (Hint: check shell quoting or use --embed-file)", err)
						}
					} else {
						return utils.ValidationErrorf("failed to parse embed JSON: %w (Hint: check shell quoting or use --embed-file)", err)
					}
				}
			} else if embedFile != "" {
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

// MessagesValidateEmbedCommand validates embed JSON
func MessagesValidateEmbedCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate-embed",
		Usage: "Validate embed JSON structure",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "embed",
				Usage: "JSON string containing embed data",
			},
			&cli.StringFlag{
				Name:  "embed-file",
				Usage: "JSON file containing embed data",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			embedJSON := c.String("embed")
			embedFile := c.String("embed-file")

			if embedJSON == "" && embedFile == "" {
				return utils.ValidationError("either --embed or --embed-file is required")
			}

			var embed discordgo.MessageEmbed
			var data []byte
			var err error

			if embedJSON != "" {
				data = []byte(embedJSON)
			} else {
				data, err = os.ReadFile(embedFile)
				if err != nil {
					return utils.ValidationErrorf("failed to read embed file: %w", err)
				}
			}

			if err := json.Unmarshal(data, &embed); err != nil {
				// Try to reconstruct JSON if it was split by shell
				if embedJSON != "" && c.Args().Len() > 0 {
					if _, ok := utils.ReconstructJSON(embedJSON, c.Args().Slice(), &embed); ok {
						fmt.Println("Warning: JSON argument appeared to be split by shell. Successfully reconstructed.")
						// Use the reconstructed embed for validation output
					} else {
						return utils.ValidationErrorf("invalid embed JSON: %w (Hint: check shell quoting or use --embed-file)", err)
					}
				} else {
					return utils.ValidationErrorf("invalid embed JSON: %w (Hint: check shell quoting or use --embed-file)", err)
				}
			}

			fmt.Println("Embed JSON is valid.")
			
			// Pretty print the parsed embed
			output, _ := json.MarshalIndent(embed, "", "  ")
			fmt.Println(string(output))

			return nil
		},
	}
}