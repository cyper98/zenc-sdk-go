package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	zenc "github.com/aminofox/zenc-sdk-go"
)

func main() {
	// Create client
	client := zenc.NewClient(zenc.ClientOptions{
		URL:         "ws://localhost:8080/ws",
		AutoConnect: true,
		Reconnect:   true,
	})

	// Handle connection events
	client.On("connected", func(data interface{}) {
		fmt.Println("✅ Connected to server")
	})

	client.On("disconnected", func(data interface{}) {
		fmt.Println("❌ Disconnected from server")
	})

	client.On("reconnecting", func(data interface{}) {
		fmt.Println("🔄 Reconnecting...")
	})

	// Connect to server
	fmt.Println("Connecting to Zenc server...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Join room
	fmt.Println("Joining room 'test-room'...")
	room, err := client.JoinRoom("test-room", &zenc.JoinOptions{
		Role: zenc.RolePublisher,
		Metadata: map[string]interface{}{
			"name": "Go Client",
		},
	})
	if err != nil {
		log.Fatalf("Failed to join room: %v", err)
	}
	defer room.Leave()

	fmt.Printf("✅ Joined room: %s (ID: %s)\n", room.Name, room.ID)
	fmt.Printf("   Local participant: %s\n", room.LocalParticipant.UserID)

	// Create and publish video track
	fmt.Println("\nCreating video track...")
	videoTrack, err := client.CreateVideoTrack("vp8")
	if err != nil {
		log.Fatalf("Failed to create video track: %v", err)
	}

	fmt.Println("Publishing video track...")
	if err := room.PublishTrack(videoTrack); err != nil {
		log.Fatalf("Failed to publish video track: %v", err)
	}
	fmt.Printf("✅ Published video track: %s\n", videoTrack.ID)

	// Create and publish audio track
	fmt.Println("\nCreating audio track...")
	audioTrack, err := client.CreateAudioTrack("opus")
	if err != nil {
		log.Fatalf("Failed to create audio track: %v", err)
	}

	fmt.Println("Publishing audio track...")
	if err := room.PublishTrack(audioTrack); err != nil {
		log.Fatalf("Failed to publish audio track: %v", err)
	}
	fmt.Printf("✅ Published audio track: %s\n", audioTrack.ID)

	// List participants
	fmt.Printf("\n📊 Participants in room (%d):\n", len(room.Participants)+1)
	fmt.Printf("   - %s (You)\n", room.LocalParticipant.UserID)
	for _, participant := range room.Participants {
		fmt.Printf("   - %s\n", participant.UserID)
	}

	// Test hold/resume
	fmt.Println("\n⏸️  Testing hold/resume...")
	time.Sleep(2 * time.Second)

	fmt.Println("Holding video track...")
	if err := videoTrack.Hold(); err != nil {
		log.Printf("Failed to hold video: %v", err)
	}
	fmt.Printf("   Video status: %s\n", videoTrack.Status)

	time.Sleep(2 * time.Second)

	fmt.Println("Resuming video track...")
	if err := videoTrack.Resume(); err != nil {
		log.Printf("Failed to resume video: %v", err)
	}
	fmt.Printf("   Video status: %s\n", videoTrack.Status)

	// Wait for interrupt signal
	fmt.Println("\n✨ Streaming... Press Ctrl+C to exit")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n👋 Leaving room and disconnecting...")
}
