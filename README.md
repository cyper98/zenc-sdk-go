# Zenc SDK for Go

Go SDK for the Zenc livestreaming platform.

## Installation

```bash
go get github.com/cyper98/zenc-sdk-go
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"time"

	zenc "github.com/cyper98/zenc-sdk-go"
)

func main() {
	// Create client
	client := zenc.NewClient(zenc.ClientOptions{
		URL:         "ws://localhost:8080/ws",
		AutoConnect: true,
		Reconnect:   true,
	})

	// Connect to server
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect()

	// Join room
	room, err := client.JoinRoom("test-room", &zenc.JoinOptions{
		Role: zenc.RolePublisher,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer room.Leave()

	fmt.Printf("Joined room: %s\n", room.Name)

	// Create and publish video track
	videoTrack, err := client.CreateVideoTrack("vp8")
	if err != nil {
		log.Fatal(err)
	}

	if err := room.PublishTrack(videoTrack); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Published video track")

	// Create and publish audio track
	audioTrack, err := client.CreateAudioTrack("opus")
	if err != nil {
		log.Fatal(err)
	}

	if err := room.PublishTrack(audioTrack); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Published audio track")

	// Handle events
	client.On("participantJoined", func(data interface{}) {
		fmt.Println("Participant joined")
	})

	client.On("participantLeft", func(data interface{}) {
		fmt.Println("Participant left")
	})

	// Keep running
	time.Sleep(60 * time.Second)
}
```

## Features

- ✅ WebRTC-based real-time streaming with Pion
- ✅ Auto-reconnection with exponential backoff
- ✅ Track management (publish/unpublish)
- ✅ Subscribe to remote participants
- ✅ Hold/resume functionality
- ✅ Connection state management
- ✅ Event-driven architecture
- ✅ Thread-safe operations

## API Reference

### Client

Main client for connecting to Zenc server.

```go
client := zenc.NewClient(zenc.ClientOptions{
    URL:               "ws://server:8080/ws",
    Token:             "jwt-token",
    AutoConnect:       true,
    Reconnect:         true,
    ReconnectAttempts: 5,
    ReconnectDelay:    time.Second,
})
```

**Methods:**
- `Connect() error` - Connect to server
- `Disconnect() error` - Disconnect from server
- `JoinRoom(roomID string, opts *JoinOptions) (*Room, error)` - Join a room
- `CreateVideoTrack(codec string) (*LocalTrack, error)` - Create video track
- `CreateAudioTrack(codec string) (*LocalTrack, error)` - Create audio track
- `GetConnectionState() ConnectionState` - Get connection state
- `On(event string, handler func(interface{}))` - Register event handler

### Room

Represents a streaming room.

**Properties:**
- `ID string` - Room ID
- `Name string` - Room name
- `Participants map[string]*Participant` - Room participants
- `LocalParticipant *LocalParticipant` - Local participant

**Methods:**
- `PublishTrack(track *LocalTrack) error` - Publish a track
- `UnpublishTrack(track *LocalTrack) error` - Unpublish a track
- `Subscribe(trackID string) (*RemoteTrack, error)` - Subscribe to track
- `Unsubscribe(trackID string) error` - Unsubscribe from track
- `Leave() error` - Leave room

### LocalTrack

Local media track.

**Methods:**
- `GetTrack() *webrtc.TrackLocalStaticRTP` - Get WebRTC track
- `Hold() error` - Pause sending
- `Resume() error` - Resume sending
- `Stop() error` - Stop track

### RemoteTrack

Remote participant track.

**Methods:**
- `GetTrack() *webrtc.TrackRemote` - Get WebRTC track
- `SetTrack(track *webrtc.TrackRemote)` - Set WebRTC track
- `Hold() error` - Pause receiving
- `Resume() error` - Resume receiving
- `Stop() error` - Stop track

## Examples

See the [examples](./examples) directory for complete examples:

- **basic** - Simple streaming client
- **server** - Server-to-server communication
- **cli** - Command-line tool

## Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build

# Run example
go run examples/basic/main.go
```

## License

MIT

## Links

- [Documentation](https://zenc.example.com/docs)
- [GitHub](https://github.com/cyper98/zenc-sdk-go)
- [Issues](https://github.com/cyper98/zenc-sdk-go/issues)
