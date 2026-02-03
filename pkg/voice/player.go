package voice

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	// Audio frame size for Discord (20ms at 48kHz stereo)
	frameSize   = 960
	channels    = 2
	sampleRate  = 48000
	maxOpusSize = 1275 // Maximum size of an Opus frame
)

// Player handles audio playback to Discord voice channels using ffmpeg
type Player struct {
	conn     *Connection
	stopChan chan struct{}
	stopped  bool
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

// NewPlayer creates a new audio player
func NewPlayer(conn *Connection) *Player {
	return &Player{
		conn:     conn,
		stopChan: make(chan struct{}),
	}
}

// PlayFile plays an audio file using ffmpeg
// It automatically joins the channel, plays the file, and leaves when done or on interrupt
func (p *Player) PlayFile(filePath string) error {
	// Check if ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH. Please install ffmpeg to use this command")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-sigChan:
			fmt.Println("\nReceived interrupt signal, stopping playback...")
			cancel()
			p.Stop()
		case <-p.stopChan:
			cancel()
		}
	}()

	fmt.Printf("Playing %s... Press Ctrl+C to stop\n", filePath)

	err := p.play(ctx, filePath)

	p.Stop()

	return err
}

// play starts the ffmpeg process and streams audio to Discord
func (p *Player) play(ctx context.Context, filePath string) error {
	voice := p.conn.GetVoiceConnection()
	if voice == nil {
		return fmt.Errorf("voice connection is nil")
	}

	voice.Speaking(true)
	defer voice.Speaking(false)

	// ffmpeg options:
	// -i input: input file
	// -f opus: output format opus
	// -ar 48000: sample rate 48kHz
	// -ac 2: 2 channels (stereo)
	// -b:a 128k: bitrate
	// -loglevel error: only show errors
	// -: output to stdout
	// Note: We use -application voip for better voice quality if needed, but audio is generic
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", filePath,
		"-f", "opus",
		"-ar", "48000",
		"-ac", "2",
		"-b:a", "128k",
		"-application", "audio",
		"-loglevel", "error",
		"-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintf(os.Stderr, "ffmpeg: %s\n", scanner.Text())
		}
	}()

	errChan := make(chan error, 1)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		errChan <- p.streamAudio(ctx, stdout)
	}()

	select {
	case err := <-errChan:
		cmd.Wait()
		if err != nil && err != io.EOF && ctx.Err() == nil {
			return fmt.Errorf("audio streaming error: %w", err)
		}
	case <-ctx.Done():
		cmd.Process.Kill()
		cmd.Wait()
		return nil
	}

	return nil
}

// streamAudio reads Ogg/Opus data from ffmpeg and sends parsed Opus frames to Discord
func (p *Player) streamAudio(ctx context.Context, reader io.Reader) error {
	voice := p.conn.GetVoiceConnection()
	if voice == nil {
		return fmt.Errorf("voice connection is nil")
	}

	// Ticker for sending packets at 20ms intervals
	// We start the ticker after finding the first audio packet
	var ticker *time.Ticker

	bufReader := bufio.NewReader(reader)

	// Ogg Page Header structure
	// Capture Pattern (4 bytes) - OggS
	// Version (1 byte)
	// Header Type (1 byte)
	// Granule Position (8 bytes)
	// Serial Number (4 bytes)
	// Page Sequence Number (4 bytes)
	// Checksum (4 bytes)
	// Page Segments (1 byte)
	// Segment Table (Variable)

	headerBuf := make([]byte, 27) // Fixed header size

	pageCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-p.stopChan:
			return nil
		default:
		}

		// Read OggS capture pattern and fixed header
		_, err := io.ReadFull(bufReader, headerBuf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read Ogg header: %w", err)
		}

		if string(headerBuf[:4]) != "OggS" {
			// Find sync? For now just fail or skip
			// Simple re-sync strategy: scan until OggS
			// But for pipe output it should be aligned
			return fmt.Errorf("invalid Ogg capture pattern")
		}

		// Read number of segments
		numSegments := int(headerBuf[26])
		segmentTable := make([]byte, numSegments)
		_, err = io.ReadFull(bufReader, segmentTable)
		if err != nil {
			return fmt.Errorf("failed to read segment table: %w", err)
		}

		// Process segments
		var packetBuf bytes.Buffer
		for _, segmentLen := range segmentTable {
			l := int(segmentLen)
			data := make([]byte, l)
			_, err := io.ReadFull(bufReader, data)
			if err != nil {
				return fmt.Errorf("failed to read segment data: %w", err)
			}
			packetBuf.Write(data)

			// If length < 255, it's the end of the packet
			if l < 255 {
				packet := packetBuf.Bytes()
				packetBuf.Reset()

				// Skip the first two pages (ID header and Comment header)
				// Page 0: OpusHead
				// Page 1: OpusTags
				// Audio starts from Page 2
				if pageCount < 2 {
					continue
				}

				// Initialize ticker on first audio packet
				if ticker == nil {
					ticker = time.NewTicker(20 * time.Millisecond)
					defer ticker.Stop()
				}

				// Wait for tick
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return nil
				case <-p.stopChan:
					return nil
				}

				// Send Opus packet
				// We need to copy the packet because packetBuf.Reset() reuses memory?
				// Actually Bytes() returns a slice of the buffer, and Reset() truncates.
				// For safety, we copy.
				toSend := make([]byte, len(packet))
				copy(toSend, packet)
				
				voice.OpusSend <- toSend
			}
		}

		pageCount++
	}
}

// Stop stops the playback
func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopped {
		return nil
	}
	p.stopped = true

	close(p.stopChan)

	// Wait for playback goroutine to finish
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		fmt.Println("Warning: timeout waiting for playback to stop")
	}

	return nil
}

// IsPlaying returns whether the player is currently playing
func (p *Player) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return !p.stopped
}
