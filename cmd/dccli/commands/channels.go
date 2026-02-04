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

func channelTypeString(channelType discordgo.ChannelType) string {
	switch channelType {
	case discordgo.ChannelTypeGuildText:
		return "Text"
	case discordgo.ChannelTypeGuildVoice:
		return "Voice"
	case discordgo.ChannelTypeGuildCategory:
		return "Category"
	case discordgo.ChannelTypeGuildNews:
		return "News"
	case discordgo.ChannelTypeGuildStageVoice:
		return "Stage"
	case discordgo.ChannelTypeGuildForum:
		return "Forum"
	default:
		return fmt.Sprintf("Unknown (%d)", channelType)
	}
}

func ChannelsCommand() *cli.Command {
	return &cli.Command{
		Name:  "channels",
		Usage: "Channel operations",
		Commands: []*cli.Command{
			ChannelsListCommand(),
			ChannelsGetCommand(),
			ChannelsCreateCommand(),
			ChannelsEditCommand(),
			ChannelsDeleteCommand(),
			ChannelsMessagesCommand(),
			ChannelsWebhooksCommand(),
		},
	}
}

func ChannelsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List channels",
		Flags: []cli.Flag{
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

			guildID := c.String("guild")
			if guildID == "" {
				return utils.ValidationError("guild ID is required (--guild)")
			}

			channels, err := cliCtx.Client.GetGuildChannels(guildID)
			if err != nil {
				return utils.DiscordErrorf("failed to list channels: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, ch := range channels {
					chType := channelTypeString(ch.Type)
					data = append(data, []string{ch.ID, ch.Name, chType})
				}
				header := []string{"ID", "Name", "Type"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(channels); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func ChannelsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get channel details",
		ArgsUsage: "[channel-id]",
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

			channel, err := cliCtx.Client.GetChannel(channelID)
			if err != nil {
				return utils.DiscordErrorf("failed to get channel: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", channel.ID)
				fmt.Printf("Name: %s\n", channel.Name)
				fmt.Printf("Type: %s\n", channelTypeString(channel.Type))
				fmt.Printf("Guild ID: %s\n", channel.GuildID)
				if channel.Topic != "" {
					fmt.Printf("Topic: %s\n", channel.Topic)
				}
				if channel.ParentID != "" {
					fmt.Printf("Parent ID: %s\n", channel.ParentID)
				}
				fmt.Printf("NSFW: %v\n", channel.NSFW)
				fmt.Printf("Position: %d\n", channel.Position)
			} else {
				if err := output.Print(channel); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func ChannelsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new channel",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Channel name",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "type",
				Usage: "Channel type (0=Text, 2=Voice, 4=Category, 5=News, 13=Stage)",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "topic",
				Usage: "Channel topic",
			},
			&cli.StringFlag{
				Name:  "parent",
				Usage: "Parent category ID",
			},
			&cli.BoolFlag{
				Name:  "nsfw",
				Usage: "Mark channel as NSFW",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			guildID := c.String("guild")

			data := &discordgo.GuildChannelCreateData{
				Name: c.String("name"),
				Type: discordgo.ChannelType(c.Int("type")),
			}

			if topic := c.String("topic"); topic != "" {
				data.Topic = topic
			}
			if parent := c.String("parent"); parent != "" {
				data.ParentID = parent
			}
			if c.Bool("nsfw") {
				data.NSFW = true
			}

			channel, err := cliCtx.Client.CreateGuildChannel(guildID, data)
			if err != nil {
				return utils.DiscordErrorf("failed to create channel: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Channel created successfully!\n")
				fmt.Printf("ID: %s\n", channel.ID)
				fmt.Printf("Name: %s\n", channel.Name)
				fmt.Printf("Type: %s\n", channelTypeString(channel.Type))
			} else {
				result := map[string]interface{}{
					"success": true,
					"channel": channel,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ChannelsEditCommand edits a channel
func ChannelsEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a channel",
		ArgsUsage: "[channel-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "New channel name",
			},
			&cli.StringFlag{
				Name:  "topic",
				Usage: "New channel topic",
			},
			&cli.StringFlag{
				Name:  "parent",
				Usage: "New parent category ID",
			},
			&cli.BoolFlag{
				Name:  "nsfw",
				Usage: "Set NSFW status",
			},
			&cli.IntFlag{
				Name:  "position",
				Usage: "Channel position",
			},
			&cli.IntFlag{
				Name:  "bitrate",
				Usage: "Bitrate for voice channels",
			},
			&cli.IntFlag{
				Name:  "user-limit",
				Usage: "User limit for voice channels",
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

			data := &discordgo.ChannelEdit{}

			if name := c.String("name"); name != "" {
				data.Name = name
			}
			if topic := c.String("topic"); topic != "" {
				data.Topic = topic
			}
			if parent := c.String("parent"); parent != "" {
				data.ParentID = parent
			}
			if c.IsSet("nsfw") {
				nsfw := c.Bool("nsfw")
				data.NSFW = &nsfw
			}
			if position := c.Int("position"); c.IsSet("position") {
				data.Position = &position
			}
			if bitrate := c.Int("bitrate"); c.IsSet("bitrate") {
				data.Bitrate = bitrate
			}
			if userLimit := c.Int("user-limit"); c.IsSet("user-limit") {
				data.UserLimit = userLimit
			}

			channel, err := cliCtx.Client.EditChannel(channelID, data)
			if err != nil {
				return utils.DiscordErrorf("failed to edit channel: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Channel updated successfully!\n")
				fmt.Printf("ID: %s\n", channel.ID)
				fmt.Printf("Name: %s\n", channel.Name)
			} else {
				result := map[string]interface{}{
					"success": true,
					"channel": channel,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ChannelsDeleteCommand deletes a channel
func ChannelsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a channel",
		ArgsUsage: "[channel-id]",
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
				return utils.ValidationError("channel ID is required")
			}
			channelID := c.Args().First()

			// Get channel info for confirmation
			channel, err := cliCtx.Client.GetChannel(channelID)
			if err != nil {
				return utils.DiscordErrorf("failed to get channel info: %w", err)
			}

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete channel: %s (ID: %s)\n", channel.Name, channelID)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.DeleteChannel(channelID); err != nil {
				return utils.DiscordErrorf("failed to delete channel: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted channel: %s\n", channel.Name)
			} else {
				result := map[string]interface{}{
					"success":    true,
					"channel_id": channelID,
					"name":       channel.Name,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ChannelsMessagesCommand manages channel messages
func ChannelsMessagesCommand() *cli.Command {
	return &cli.Command{
		Name:      "messages",
		Usage:     "Manage channel messages",
		ArgsUsage: "[channel-id]",
		Commands: []*cli.Command{
			ChannelMessagesListCommand(),
			ChannelMessagesGetCommand(),
			ChannelMessagesSendCommand(),
			ChannelMessagesEditCommand(),
			ChannelMessagesDeleteCommand(),
			ChannelMessagesBulkDeleteCommand(),
		},
	}
}

// ChannelMessagesListCommand lists messages in a channel
func ChannelMessagesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List messages in the channel",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "limit",
				Usage: "Number of messages to retrieve (max 100)",
				Value: 50,
			},
			&cli.StringFlag{
				Name:  "before",
				Usage: "Get messages before this message ID",
			},
			&cli.StringFlag{
				Name:  "after",
				Usage: "Get messages after this message ID",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			channelID := c.String("channel")
			limit := c.Int("limit")
			if limit > 100 {
				limit = 100
			}

			messages, err := cliCtx.Client.GetChannelMessages(channelID, limit, c.String("before"), c.String("after"), "")
			if err != nil {
				return utils.DiscordErrorf("failed to list messages: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, m := range messages {
					author := m.Author.Username
					content := m.Content
					if len(content) > 50 {
						content = content[:47] + "..."
					}
					if content == "" {
						content = "[embed/attachment]"
					}
					data = append(data, []string{
						m.ID,
						author,
						content,
						m.Timestamp.Format("2006-01-02 15:04"),
					})
				}
				header := []string{"ID", "Author", "Content", "Time"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(messages); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ChannelMessagesGetCommand gets a specific message
func ChannelMessagesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific message",
		ArgsUsage: "[message-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
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
				return utils.ValidationError("message ID is required")
			}
			messageID := c.Args().First()
			channelID := c.String("channel")

			message, err := cliCtx.Client.GetChannelMessage(channelID, messageID)
			if err != nil {
				return utils.DiscordErrorf("failed to get message: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", message.ID)
				fmt.Printf("Author: %s#%s (%s)\n", message.Author.Username, message.Author.Discriminator, message.Author.ID)
				fmt.Printf("Content: %s\n", message.Content)
				fmt.Printf("Timestamp: %s\n", message.Timestamp.Format("2006-01-02 15:04:05"))
				if len(message.Embeds) > 0 {
					fmt.Printf("Embeds: %d\n", len(message.Embeds))
				}
				if len(message.Attachments) > 0 {
					fmt.Printf("Attachments: %d\n", len(message.Attachments))
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

// ChannelMessagesSendCommand sends a message to a channel
func ChannelMessagesSendCommand() *cli.Command {
	return &cli.Command{
		Name:  "send",
		Usage: "Send a message to the channel",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
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

			channelID := c.String("channel")
			content := c.String("content")
			embedJSON := c.String("embed")
			embedFile := c.String("embed-file")

			if content == "" && embedJSON == "" && embedFile == "" {
				return utils.ValidationError("either --content, --embed, or --embed-file is required")
			}

			msg := &discordgo.MessageSend{
				Content: content,
				TTS:     c.Bool("tts"),
			}

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

// ChannelMessagesEditCommand edits a message
func ChannelMessagesEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a message",
		ArgsUsage: "[message-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
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

			if c.NArg() < 1 {
				return utils.ValidationError("message ID is required")
			}
			messageID := c.Args().First()
			channelID := c.String("channel")

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

// ChannelMessagesDeleteCommand deletes a message
func ChannelMessagesDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a message",
		ArgsUsage: "[message-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
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
				return utils.ValidationError("message ID is required")
			}
			messageID := c.Args().First()
			channelID := c.String("channel")

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete message: %s\n", messageID)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
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

// ChannelMessagesBulkDeleteCommand bulk deletes messages
func ChannelMessagesBulkDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "bulk-delete",
		Usage:     "Bulk delete messages",
		ArgsUsage: "[message-ids...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
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
				return utils.ValidationError("at least one message ID is required")
			}

			channelID := c.String("channel")
			messageIDs := c.Args().Slice()

			if len(messageIDs) > 100 {
				return utils.ValidationError("cannot delete more than 100 messages at once")
			}

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete %d messages\n", len(messageIDs))
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.BulkDeleteMessages(channelID, messageIDs); err != nil {
				return utils.DiscordErrorf("failed to bulk delete messages: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted %d messages\n", len(messageIDs))
			} else {
				result := map[string]interface{}{
					"success":     true,
					"deleted":     len(messageIDs),
					"message_ids": messageIDs,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ChannelsWebhooksCommand manages channel webhooks
func ChannelsWebhooksCommand() *cli.Command {
	return &cli.Command{
		Name:      "webhooks",
		Usage:     "Manage channel webhooks",
		ArgsUsage: "[channel-id]",
		Commands: []*cli.Command{
			ChannelWebhooksListCommand(),
			ChannelWebhooksGetCommand(),
			ChannelWebhooksCreateCommand(),
			ChannelWebhooksEditCommand(),
			ChannelWebhooksDeleteCommand(),
		},
	}
}

// ChannelWebhooksListCommand lists webhooks in a channel
func ChannelWebhooksListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List webhooks in the channel",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
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

			webhooks, err := cliCtx.Client.GetChannelWebhooks(channelID)
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
						wh.Token,
						avatar,
					})
				}
				header := []string{"ID", "Name", "Token", "Avatar"}
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

// ChannelWebhooksGetCommand gets a specific webhook
func ChannelWebhooksGetCommand() *cli.Command {
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
			} else {
				if err := output.Print(webhook); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ChannelWebhooksCreateCommand creates a webhook
func ChannelWebhooksCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a webhook",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Channel ID",
				Required: true,
			},
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

			channelID := c.String("channel")
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

// ChannelWebhooksEditCommand edits a webhook
func ChannelWebhooksEditCommand() *cli.Command {
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

			name := ""
			var avatar string

			if n := c.String("name"); n != "" {
				name = n
			}

			if avatarPath := c.String("avatar"); avatarPath != "" {
				avatarData, err := os.ReadFile(avatarPath)
				if err != nil {
					return utils.ValidationErrorf("failed to read avatar file: %w", err)
				}
				avatar = "data:image/png;base64," + base64.StdEncoding.EncodeToString(avatarData)
			}

			webhook, err := cliCtx.Client.EditWebhook(webhookID, name, avatar, "")
			if err != nil {
				return utils.DiscordErrorf("failed to edit webhook: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Webhook updated successfully!\n")
				fmt.Printf("ID: %s\n", webhook.ID)
				fmt.Printf("Name: %s\n", webhook.Name)
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

// ChannelWebhooksDeleteCommand deletes a webhook
func ChannelWebhooksDeleteCommand() *cli.Command {
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
