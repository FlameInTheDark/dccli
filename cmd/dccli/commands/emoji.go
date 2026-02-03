package commands

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
)

func EmojiCommand() *cli.Command {
	return &cli.Command{
		Name:  "emoji",
		Usage: "Emoji management",
		Commands: []*cli.Command{
			EmojiListCommand(),
			EmojiGetCommand(),
			EmojiCreateCommand(),
			EmojiEditCommand(),
			EmojiDeleteCommand(),
		},
	}
}

// EmojiListCommand lists all emojis in a guild
func EmojiListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all emojis in a guild",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			guildID := c.String("guild")

			emojis, err := cliCtx.Client.GetGuildEmojis(guildID)
			if err != nil {
				return utils.DiscordErrorf("failed to list emojis: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, emoji := range emojis {
					animated := "No"
					if emoji.Animated {
						animated = "Yes"
					}
					managed := "No"
					if emoji.Managed {
						managed = "Yes"
					}
					data = append(data, []string{
						emoji.ID,
						emoji.Name,
						animated,
						managed,
					})
				}
				header := []string{"ID", "Name", "Animated", "Managed"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(emojis); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// EmojiGetCommand gets emoji details
func EmojiGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get emoji details",
		ArgsUsage: "[emoji-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
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
				return utils.ValidationError("emoji ID is required")
			}
			emojiID := c.Args().First()
			guildID := c.String("guild")

			emoji, err := cliCtx.Client.GetGuildEmoji(guildID, emojiID)
			if err != nil {
				return utils.DiscordErrorf("failed to get emoji: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", emoji.ID)
				fmt.Printf("Name: %s\n", emoji.Name)
				fmt.Printf("Animated: %v\n", emoji.Animated)
				fmt.Printf("Managed: %v\n", emoji.Managed)
				fmt.Printf("Require Colons: %v\n", emoji.RequireColons)
				if emoji.User != nil {
					fmt.Printf("Creator: %s#%s (%s)\n", emoji.User.Username, emoji.User.Discriminator, emoji.User.ID)
				}
				fmt.Printf("URL: https://cdn.discordapp.com/emojis/%s.%s\n", emoji.ID, getEmojiExtension(emoji))
			} else {
				if err := output.Print(emoji); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// EmojiCreateCommand creates a new emoji
func EmojiCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new emoji",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Emoji name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "image",
				Usage:    "Path to image file (PNG, JPG, GIF)",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:  "roles",
				Usage: "Roles that can use this emoji (comma-separated IDs)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			guildID := c.String("guild")
			name := c.String("name")
			imagePath := c.String("image")

			// Read image file
			imageData, err := os.ReadFile(imagePath)
			if err != nil {
				return utils.ValidationErrorf("failed to read image file: %w", err)
			}

			// Determine MIME type from file extension
			mimeType := getMimeType(imagePath)
			imageBase64 := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageData))

			params := &discordgo.EmojiParams{
				Name:  name,
				Image: imageBase64,
			}

			// Add role restrictions if specified
			if roles := c.StringSlice("roles"); len(roles) > 0 {
				params.Roles = roles
			}

			emoji, err := cliCtx.Client.CreateGuildEmoji(guildID, params)
			if err != nil {
				return utils.DiscordErrorf("failed to create emoji: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Emoji created successfully!\n")
				fmt.Printf("ID: %s\n", emoji.ID)
				fmt.Printf("Name: %s\n", emoji.Name)
				fmt.Printf("URL: https://cdn.discordapp.com/emojis/%s.%s\n", emoji.ID, getEmojiExtension(emoji))
			} else {
				result := map[string]interface{}{
					"success": true,
					"emoji":   emoji,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// EmojiEditCommand edits an emoji
func EmojiEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit an emoji",
		ArgsUsage: "[emoji-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "New emoji name",
			},
			&cli.StringSliceFlag{
				Name:  "roles",
				Usage: "Roles that can use this emoji (comma-separated IDs)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("emoji ID is required")
			}
			emojiID := c.Args().First()
			guildID := c.String("guild")

			// Get existing emoji first
			existingEmoji, err := cliCtx.Client.GetGuildEmoji(guildID, emojiID)
			if err != nil {
				return utils.DiscordErrorf("failed to get emoji: %w", err)
			}

			params := &discordgo.EmojiParams{}

			// Update name if provided
			if newName := c.String("name"); newName != "" {
				params.Name = newName
			} else {
				params.Name = existingEmoji.Name
			}

			// Update roles if provided
			if roles := c.StringSlice("roles"); len(roles) > 0 {
				params.Roles = roles
			} else {
				roleIDs := make([]string, len(existingEmoji.Roles))
				copy(roleIDs, existingEmoji.Roles)
				params.Roles = roleIDs
			}

			emoji, err := cliCtx.Client.EditGuildEmoji(guildID, emojiID, params)
			if err != nil {
				return utils.DiscordErrorf("failed to edit emoji: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Emoji updated successfully!\n")
				fmt.Printf("ID: %s\n", emoji.ID)
				fmt.Printf("Name: %s\n", emoji.Name)
			} else {
				result := map[string]interface{}{
					"success": true,
					"emoji":   emoji,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// EmojiDeleteCommand deletes an emoji
func EmojiDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete an emoji",
		ArgsUsage: "[emoji-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
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
				return utils.ValidationError("emoji ID is required")
			}
			emojiID := c.Args().First()
			guildID := c.String("guild")

			// Get emoji info for confirmation
			emoji, err := cliCtx.Client.GetGuildEmoji(guildID, emojiID)
			if err != nil {
				return utils.DiscordErrorf("failed to get emoji info: %w", err)
			}

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete emoji: :%s: (ID: %s)\n", emoji.Name, emojiID)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.DeleteGuildEmoji(guildID, emojiID); err != nil {
				return utils.DiscordErrorf("failed to delete emoji: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted emoji: :%s:\n", emoji.Name)
			} else {
				result := map[string]interface{}{
					"success":  true,
					"emoji_id": emojiID,
					"name":     emoji.Name,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// Helper functions

func getEmojiExtension(emoji *discordgo.Emoji) string {
	if emoji.Animated {
		return "gif"
	}
	return "png"
}

func getMimeType(path string) string {
	ext := strings.ToLower(path)
	if strings.HasSuffix(ext, ".gif") {
		return "image/gif"
	}
	if strings.HasSuffix(ext, ".png") {
		return "image/png"
	}
	if strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		return "image/jpeg"
	}
	// Default to PNG
	return "image/png"
}