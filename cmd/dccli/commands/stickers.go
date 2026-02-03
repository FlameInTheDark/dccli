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

func StickersCommand() *cli.Command {
	return &cli.Command{
		Name:  "stickers",
		Usage: "Sticker management",
		Commands: []*cli.Command{
			StickersListCommand(),
			StickersGetCommand(),
			StickersCreateCommand(),
			StickersEditCommand(),
			StickersDeleteCommand(),
		},
	}
}

// StickersListCommand lists all stickers in a guild
func StickersListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all stickers in a guild",
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

			stickers, err := cliCtx.Client.GetGuildStickers(guildID)
			if err != nil {
				return utils.DiscordErrorf("failed to list stickers: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, sticker := range stickers {
					tags := sticker.Tags
					if tags == "" {
						tags = "-"
					}
					data = append(data, []string{
						sticker.ID,
						sticker.Name,
						stickerTypeToString(sticker.Type),
						tags,
					})
				}
				header := []string{"ID", "Name", "Type", "Tags"}
				dprint.Table(header, data)
			} else {
				if err := output.Print(stickers); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// StickersGetCommand gets sticker details
func StickersGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get sticker details",
		ArgsUsage: "[sticker-id]",
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
				return utils.ValidationError("sticker ID is required")
			}
			stickerID := c.Args().First()
			guildID := c.String("guild")

			sticker, err := cliCtx.Client.GetGuildSticker(guildID, stickerID)
			if err != nil {
				return utils.DiscordErrorf("failed to get sticker: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("ID: %s\n", sticker.ID)
				fmt.Printf("Name: %s\n", sticker.Name)
				fmt.Printf("Description: %s\n", sticker.Description)
				fmt.Printf("Tags: %s\n", sticker.Tags)
				fmt.Printf("Type: %s\n", stickerTypeToString(sticker.Type))
				fmt.Printf("Format: %s\n", stickerFormatToString(sticker.FormatType))
				if sticker.User != nil {
					fmt.Printf("Creator: %s#%s (%s)\n", sticker.User.Username, sticker.User.Discriminator, sticker.User.ID)
				}
				fmt.Printf("Available: %v\n", sticker.Available)
			} else {
				if err := output.Print(sticker); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// StickersCreateCommand creates a new sticker
func StickersCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new sticker",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Sticker name (2-30 characters)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "image",
				Usage:    "Path to image file (PNG, APNG, Lottie)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "tags",
				Usage: "Sticker tags (comma-separated, max 200 characters)",
			},
			&cli.StringFlag{
				Name:  "description",
				Usage: "Sticker description (max 100 characters)",
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

			// Validate file size (max 500KB for stickers)
			if len(imageData) > 512000 {
				return utils.ValidationError("sticker file too large (max 500KB)")
			}

			// Suppress unused variable warnings - these are used for future implementation
			_ = getStickerMimeType(imagePath)
			_ = base64.StdEncoding.EncodeToString(imageData)

			sticker, err := cliCtx.Client.CreateGuildSticker(guildID, name, c.String("description"), c.String("tags"), imagePath)
			if err != nil {
				return utils.DiscordErrorf("failed to create sticker: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Sticker created successfully!\n")
				fmt.Printf("ID: %s\n", sticker.ID)
				fmt.Printf("Name: %s\n", sticker.Name)
				fmt.Printf("Tags: %s\n", sticker.Tags)
			} else {
				result := map[string]interface{}{
					"success": true,
					"sticker": sticker,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// StickersEditCommand edits a sticker
func StickersEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a sticker",
		ArgsUsage: "[sticker-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "guild",
				Usage:    "Guild ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "New sticker name (2-30 characters)",
			},
			&cli.StringFlag{
				Name:  "tags",
				Usage: "New sticker tags (comma-separated, max 200 characters)",
			},
			&cli.StringFlag{
				Name:  "description",
				Usage: "New sticker description (max 100 characters)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if c.NArg() < 1 {
				return utils.ValidationError("sticker ID is required")
			}
			stickerID := c.Args().First()
			guildID := c.String("guild")

			// Get existing sticker first
			existingSticker, err := cliCtx.Client.GetGuildSticker(guildID, stickerID)
			if err != nil {
				return utils.DiscordErrorf("failed to get sticker: %w", err)
			}

			// Get values for editing
			newName := existingSticker.Name
			if name := c.String("name"); name != "" {
				newName = name
			}

			newDesc := existingSticker.Description
			if desc := c.String("description"); desc != "" {
				newDesc = desc
			}

			newTags := existingSticker.Tags
			if tags := c.String("tags"); tags != "" {
				newTags = tags
			}

			sticker, err := cliCtx.Client.EditGuildSticker(guildID, stickerID, newName, newDesc, newTags)
			if err != nil {
				return utils.DiscordErrorf("failed to edit sticker: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Sticker updated successfully!\n")
				fmt.Printf("ID: %s\n", sticker.ID)
				fmt.Printf("Name: %s\n", sticker.Name)
				fmt.Printf("Tags: %s\n", sticker.Tags)
			} else {
				result := map[string]interface{}{
					"success": true,
					"sticker": sticker,
				}
				if err := output.Print(result); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// StickersDeleteCommand deletes a sticker
func StickersDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a sticker",
		ArgsUsage: "[sticker-id]",
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
				return utils.ValidationError("sticker ID is required")
			}
			stickerID := c.Args().First()
			guildID := c.String("guild")

			// Get sticker info for confirmation
			sticker, err := cliCtx.Client.GetGuildSticker(guildID, stickerID)
			if err != nil {
				return utils.DiscordErrorf("failed to get sticker info: %w", err)
			}

			// Confirmation prompt
			if !c.Bool("force") {
				fmt.Printf("You are about to delete sticker: %s (ID: %s)\n", sticker.Name, stickerID)
				fmt.Print("Are you sure? (yes/no): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if response != "yes\n" && response != "yes\r\n" {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			if err := cliCtx.Client.DeleteGuildSticker(guildID, stickerID); err != nil {
				return utils.DiscordErrorf("failed to delete sticker: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				fmt.Printf("Successfully deleted sticker: %s\n", sticker.Name)
			} else {
				result := map[string]interface{}{
					"success":    true,
					"sticker_id": stickerID,
					"name":       sticker.Name,
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

func stickerTypeToString(stickerType discordgo.StickerType) string {
	switch stickerType {
	case discordgo.StickerTypeStandard:
		return "Standard"
	case discordgo.StickerTypeGuild:
		return "Guild"
	default:
		return "Unknown"
	}
}

func stickerFormatToString(format discordgo.StickerFormat) string {
	switch format {
	case discordgo.StickerFormatTypePNG:
		return "PNG"
	case discordgo.StickerFormatTypeAPNG:
		return "APNG"
	case discordgo.StickerFormatTypeLottie:
		return "Lottie"
	default:
		return "Unknown"
	}
}

func getStickerMimeType(path string) string {
	ext := strings.ToLower(path)
	if strings.HasSuffix(ext, ".json") {
		return "application/json"
	}
	if strings.HasSuffix(ext, ".apng") {
		return "image/apng"
	}
	if strings.HasSuffix(ext, ".png") {
		return "image/png"
	}
	// Default to PNG
	return "image/png"
}

func getStickerExtension(path string) string {
	ext := strings.ToLower(path)
	if strings.HasSuffix(ext, ".json") {
		return "json"
	}
	if strings.HasSuffix(ext, ".apng") {
		return "apng"
	}
	if strings.HasSuffix(ext, ".png") {
		return "png"
	}
	return "png"
}