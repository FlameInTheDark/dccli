package voice

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

// Recorder handles audio recording from Discord voice channels
type Recorder struct {
	conn     *Connection
	stopChan chan struct{}
	stopped  bool
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

// NewRecorder creates a new audio recorder
func NewRecorder(conn *Connection) (*Recorder, error) {
	return &Recorder{
		conn:     conn,
		stopChan: make(chan struct{}),
	}, nil
}

// createRTPPacket creates a pion RTP packet from a DiscordGo packet
func createRTPPacket(p *discordgo.Packet) *rtp.Packet {
	return &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    0x78,
			SequenceNumber: p.Sequence,
			Timestamp:      p.Timestamp,
			SSRC:           p.SSRC,
		},
		Payload: p.Opus,
	}
}

// RecordToFile starts recording audio to an OGG file
func (r *Recorder) RecordToFile(outputPath string) error {
	// Extract directory and base filename from outputPath
	// If outputPath is a directory, use "recording" as base name
	// If outputPath has an extension, remove it to use as base name
	dir := filepath.Dir(outputPath)
	baseName := filepath.Base(outputPath)

	// Remove .ogg extension if present to get clean base name
	baseName = strings.TrimSuffix(baseName, ".ogg")

	// If baseName is empty or just ".ogg" was stripped, use "recording"
	if baseName == "" || baseName == "." {
		baseName = "recording"
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for directory %s: %w", dir, err)
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", absDir, err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	voice := r.conn.GetVoiceConnection()
	if voice == nil {
		return fmt.Errorf("voice connection is nil")
	}

	fmt.Printf("Recording... Press Ctrl+C to stop\n")
	startTime := time.Now().Unix()
	fmt.Printf("Files will be saved to: %s with pattern %d_<SSRC>_%s.ogg\n", absDir, startTime, baseName)

	files := make(map[uint32]media.Writer)
	var filesMu sync.Mutex

	r.wg.Add(1)
	done := make(chan struct{})
	go func() {
		defer r.wg.Done()
		defer close(done)

		for {
			select {
			case <-r.stopChan:
				return
			case packet, ok := <-voice.OpusRecv:
				if !ok {
					return
				}

				filesMu.Lock()
				file, exists := files[packet.SSRC]
				if !exists {
					// <startTime>_<SSRC>_<baseName>.ogg
					filename := filepath.Join(absDir, fmt.Sprintf("%d_%d_%s.ogg", startTime, packet.SSRC, baseName))
					var err error
					oggFile, err := os.Create(filename)
					if err != nil {
						fmt.Printf("failed to create file %s for SSRC %d: %v\n", filename, packet.SSRC, err)
						filesMu.Unlock()
						continue
					}
					file, err = oggwriter.NewWith(oggFile, 48000, 2)
					if err != nil {
						fmt.Printf("failed to create OGG writer for %s (SSRC %d): %v\n", filename, packet.SSRC, err)
						oggFile.Close()
						filesMu.Unlock()
						continue
					}
					fmt.Printf("Created file: %s\n", filename)
					files[packet.SSRC] = file
					fmt.Printf("Started recording user with SSRC %d\n", packet.SSRC)
				}
				filesMu.Unlock()

				rtpPacket := createRTPPacket(packet)
				err := file.WriteRTP(rtpPacket)
				if err != nil {
					fmt.Printf("failed to write RTP packet for SSRC %d: %v\n", packet.SSRC, err)
				}
			}
		}
	}()

	select {
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal, stopping recording...")
	case <-r.stopChan:
	}

	r.Stop()
	voice.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		fmt.Println("Warning: timeout waiting for recording to stop")
	}

	filesMu.Lock()
	for ssrc, f := range files {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close file for SSRC %d: %v\n", ssrc, err)
		}
	}
	filesMu.Unlock()

	fmt.Println("Recording saved successfully")
	return nil
}

// Stop stops the recording
func (r *Recorder) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		return nil
	}
	r.stopped = true

	close(r.stopChan)
	return nil
}

// IsRecording returns whether the recorder is currently recording
func (r *Recorder) IsRecording() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return !r.stopped
}
