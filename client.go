package zenc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

// Client is the main SDK client
type Client struct {
	signalingClient *SignalingClient
	currentRoom     *Room
	autoConnect     bool
}

// NewClient creates a new Zenc client
func NewClient(opts ClientOptions) *Client {
	client := &Client{
		signalingClient: NewSignalingClient(opts),
		autoConnect:     opts.AutoConnect,
	}

	if client.autoConnect {
		go func() {
			if err := client.Connect(); err != nil {
				// Log error or emit event
			}
		}()
	}

	return client
}

// Connect connects to the Zenc server
func (c *Client) Connect() error {
	return c.signalingClient.Connect()
}

// Disconnect disconnects from the Zenc server
func (c *Client) Disconnect() error {
	if c.currentRoom != nil {
		if err := c.currentRoom.Leave(); err != nil {
			return err
		}
		c.currentRoom = nil
	}

	c.signalingClient.Disconnect()
	return nil
}

// JoinRoom joins a room
func (c *Client) JoinRoom(roomID string, opts *JoinOptions) (*Room, error) {
	if c.currentRoom != nil {
		return nil, fmt.Errorf("already in a room, leave current room first")
	}

	connectionState := c.signalingClient.GetConnectionState()
	if connectionState != ConnectionStateConnected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	role := RolePublisher
	if opts != nil && opts.Role != "" {
		role = opts.Role
	}

	metadata := make(map[string]interface{})
	if opts != nil && opts.Metadata != nil {
		metadata = opts.Metadata
	}

	request := JoinRoomRequest{
		RoomID:   roomID,
		UserID:   fmt.Sprintf("user_%d_%s", time.Now().Unix(), uuid.New().String()[:8]),
		Role:     role,
		Metadata: metadata,
	}

	responseChan := make(chan JoinRoomResponse, 1)
	errorChan := make(chan error, 1)

	c.signalingClient.Once("room.joined", func(data interface{}) {
		jsonData, _ := json.Marshal(data)
		var response JoinRoomResponse
		if err := json.Unmarshal(jsonData, &response); err != nil {
			errorChan <- err
			return
		}
		responseChan <- response
	})

	if err := c.signalingClient.Send(WSMessage{
		Type: "room.join",
		Data: request,
	}); err != nil {
		return nil, err
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		room := NewRoom(
			response.Room,
			ParticipantInfo{
				ID:       response.ParticipantID,
				UserID:   request.UserID,
				Role:     request.Role,
				JoinedAt: time.Now().Unix(),
			},
			c.signalingClient,
		)

		// Add existing participants
		for _, participantInfo := range response.Participants {
			participant := NewParticipant(participantInfo)
			room.Participants[participant.ID] = participant
		}

		c.currentRoom = room
		return room, nil

	case err := <-errorChan:
		return nil, err

	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("join room timeout")
	}
}

// CreateVideoTrack creates a new video track
func (c *Client) CreateVideoTrack(codec string) (*LocalTrack, error) {
	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		uuid.New().String(),
		uuid.New().String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create video track: %w", err)
	}

	return NewLocalTrack(
		track.ID(),
		TrackKindVideo,
		TrackSourceCamera,
		track,
	), nil
}

// CreateAudioTrack creates a new audio track
func (c *Client) CreateAudioTrack(codec string) (*LocalTrack, error) {
	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		uuid.New().String(),
		uuid.New().String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio track: %w", err)
	}

	return NewLocalTrack(
		track.ID(),
		TrackKindAudio,
		TrackSourceMicrophone,
		track,
	), nil
}

// GetCurrentRoom returns the current room
func (c *Client) GetCurrentRoom() *Room {
	return c.currentRoom
}

// GetConnectionState returns the current connection state
func (c *Client) GetConnectionState() ConnectionState {
	return c.signalingClient.GetConnectionState()
}

// On registers an event handler
func (c *Client) On(event string, handler func(interface{})) {
	c.signalingClient.On(event, handler)
}
