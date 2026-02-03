package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/FlameInTheDark/dccli/pkg/dprint"
	"github.com/FlameInTheDark/dccli/pkg/utils"
	voicepkg "github.com/FlameInTheDark/dccli/pkg/voice"
)

// VoiceRootCommand returns the voice command group
func VoiceRootCommand() *cli.Command {
	return VoiceCommand()
}

// VoiceRegionsCommand lists available voice regions
func VoiceRegionsCommand() *cli.Command {
	return &cli.Command{
		Name:  "regions",
		Usage: "List available voice regions",
		Action: func(ctx context.Context, c *cli.Command) error {
			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			regions, err := cliCtx.Client.GetVoiceRegions()
			if err != nil {
				return utils.DiscordErrorf("failed to list voice regions: %w", err)
			}

			output := cliCtx.GetOutputManager()

			if output.GetFormat() == dprint.FormatTable {
				data := [][]string{}
				for _, region := range regions {
					data = append(data, []string{
						region.ID,
						region.Name,
					})
				}
				header := []string{"ID", "Name"}
				dprint.Table(header, data)
			} else {
				for _, region := range regions {
					dprint.Print(fmt.Sprintf("%s: %s\n", region.ID, region.Name))
				}
			}

			return nil
		},
	}
}

// VoiceJoinCommand joins a voice channel
func VoiceJoinCommand() *cli.Command {
	return &cli.Command{
		Name:      "join",
		Usage:     "Join a voice channel",
		ArgsUsage: "<channel-id> [guild-id]",
		Description: `Joins a voice channel by ID.
If guild ID is not provided, it will be auto-detected from the channel.
The bot will stay connected until manually disconnected or the program exits.`,
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("channel ID is required")
			}

			channelID := c.Args().First()
			var guildID string
			if c.NArg() > 1 {
				guildID = c.Args().Get(1)
			}

			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if guildID == "" {
				channel, err := cliCtx.Client.GetChannel(channelID)
				if err != nil {
					return utils.DiscordErrorf("failed to get channel info: %w", err)
				}
				guildID = channel.GuildID
			}

			voiceConn, err := cliCtx.Client.Session().ChannelVoiceJoin(guildID, channelID, false, false)
			if err != nil {
				return utils.DiscordErrorf("failed to join voice channel: %w", err)
			}

			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			ready := false

			for !ready {
				select {
				case <-timeout:
					voiceConn.Disconnect()
					return utils.DiscordErrorf("voice connection timeout")
				case <-ticker.C:
					if voiceConn.Ready {
						ready = true
					}
				}
			}
			ticker.Stop()

			fmt.Printf("Successfully joined voice channel %s\n", channelID)
			fmt.Println("Press Ctrl+C to disconnect and exit")

			// Wait for interrupt signal
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan

			fmt.Println("\nDisconnecting...")
			voiceConn.Disconnect()
			fmt.Println("Disconnected from voice channel")
			return nil
		},
	}
}

// VoiceLeaveCommand leaves a voice channel
func VoiceLeaveCommand() *cli.Command {
	return &cli.Command{
		Name:      "leave",
		Usage:     "Leave a voice channel",
		ArgsUsage: "<guild-id>",
		Description: `Disconnects the bot from a voice channel in the specified guild.
If the bot is not connected, this command will return an error.`,
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("guild ID is required")
			}

			guildID := c.Args().First()

			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if err := cliCtx.Client.LeaveVoiceChannel(guildID); err != nil {
				return utils.DiscordErrorf("failed to leave voice channel: %w", err)
			}

			fmt.Printf("Disconnected from voice channel in guild %s\n", guildID)
			return nil
		},
	}
}

// VoiceRecordCommand records audio from a voice channel
func VoiceRecordCommand() *cli.Command {
	return &cli.Command{
		Name:      "record",
		Usage:     "Record audio from a voice channel",
		ArgsUsage: "<channel-id> [guild-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "out",
				Aliases: []string{"o"},
				Usage:   "Output file path (without extension, .ogg will be added)",
				Value:   "recording",
			},
		},
		Description: `Joins a voice channel and records audio to OGG files.
Press Ctrl+C to stop recording and save the files.
Each user will have their own file named: <timestamp>_<SSRC>_<output>.ogg`,
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("channel ID is required")
			}

			channelID := c.Args().First()
			var guildID string
			if c.NArg() > 1 {
				guildID = c.Args().Get(1)
			}

			outputPath := c.String("out")
			if outputPath == "" {
				outputPath = "recording"
			}

			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if guildID == "" {
				channel, err := cliCtx.Client.GetChannel(channelID)
				if err != nil {
					return utils.DiscordErrorf("failed to get channel info: %w", err)
				}
				guildID = channel.GuildID
			}

			voiceConn, err := cliCtx.Client.Session().ChannelVoiceJoin(guildID, channelID, true, false)
			if err != nil {
				return utils.DiscordErrorf("failed to join voice channel: %w", err)
			}

			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			ready := false

			for !ready {
				select {
				case <-timeout:
					voiceConn.Disconnect()
					return utils.DiscordErrorf("voice connection timeout")
				case <-ticker.C:
					if voiceConn.Ready {
						ready = true
					}
				}
			}
			ticker.Stop()

			conn := voicepkg.NewConnection(voiceConn)

			recorder, err := voicepkg.NewRecorder(conn)
			if err != nil {
				voiceConn.Disconnect()
				return utils.DiscordErrorf("failed to create recorder: %w", err)
			}

			if err := recorder.RecordToFile(outputPath); err != nil {
				voiceConn.Disconnect()
				return utils.DiscordErrorf("recording failed: %w", err)
			}

			voiceConn.Disconnect()
			return nil
		},
	}
}

// VoicePlayCommand plays an audio file in a voice channel
func VoicePlayCommand() *cli.Command {
	return &cli.Command{
		Name:      "play",
		Usage:     "Play an audio file in a voice channel",
		ArgsUsage: "<channel-id> [guild-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "Audio file to play (MP3, WAV, etc.)",
				Required: true,
			},
		},
		Description: `Joins a voice channel and plays an audio file.
The bot will automatically leave the channel when playback completes
or when Ctrl+C is pressed.
Supports any audio format that ffmpeg can decode.`,
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return utils.ValidationError("channel ID is required")
			}

			channelID := c.Args().First()
			var guildID string
			if c.NArg() > 1 {
				guildID = c.Args().Get(1)
			}
			filePath := c.String("file")

			if filePath == "" {
				return utils.ValidationError("audio file path is required (--file)")
			}

			cliCtx, err := utils.NewCLIContext(c)
			if err != nil {
				return err
			}
			defer cliCtx.Close()

			if guildID == "" {
				channel, err := cliCtx.Client.GetChannel(channelID)
				if err != nil {
					return utils.DiscordErrorf("failed to get channel info: %w", err)
				}
				guildID = channel.GuildID
			}

			voiceConn, err := cliCtx.Client.Session().ChannelVoiceJoin(guildID, channelID, false, false)
			if err != nil {
				return utils.DiscordErrorf("failed to join voice channel: %w", err)
			}

			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			ready := false

			for !ready {
				select {
				case <-timeout:
					voiceConn.Disconnect()
					return utils.DiscordErrorf("voice connection timeout")
				case <-ticker.C:
					if voiceConn.Ready {
						ready = true
					}
				}
			}
			ticker.Stop()

			conn := voicepkg.NewConnection(voiceConn)

			player := voicepkg.NewPlayer(conn)

			if err := player.PlayFile(filePath); err != nil {
				voiceConn.Disconnect()
				return utils.DiscordErrorf("playback failed: %w", err)
			}

			fmt.Println("Playback finished, leaving voice channel...")
			voiceConn.Disconnect()

			return nil
		},
	}
}

// VoiceCommand returns the main voice command group
func VoiceCommand() *cli.Command {
	return &cli.Command{
		Name:        "voice",
		Usage:       "Voice channel management",
		Description: "Join, leave, record, and play audio in voice channels",
		Commands: []*cli.Command{
			VoiceRegionsCommand(),
			VoiceJoinCommand(),
			VoiceLeaveCommand(),
			VoiceRecordCommand(),
			VoicePlayCommand(),
		},
	}
}
