package zenc

import "time"

// RoomRole represents the role of a participant in a room
type RoomRole string

const (
	RolePublisher  RoomRole = "publisher"
	RoleSubscriber RoomRole = "subscriber"
	RoleModerator  RoomRole = "moderator"
)

// TrackKind represents the type of media track
type TrackKind string

const (
	TrackKindAudio TrackKind = "audio"
	TrackKindVideo TrackKind = "video"
)

// TrackSource represents the source of a media track
type TrackSource string

const (
	TrackSourceCamera     TrackSource = "camera"
	TrackSourceScreen     TrackSource = "screen"
	TrackSourceMicrophone TrackSource = "microphone"
)

// TrackStatus represents the status of a track
type TrackStatus string

const (
	TrackStatusActive  TrackStatus = "active"
	TrackStatusPaused  TrackStatus = "paused"
	TrackStatusStopped TrackStatus = "stopped"
)

// ParticipantState represents the state of a participant
type ParticipantState string

const (
	ParticipantStateActive    ParticipantState = "active"
	ParticipantStateOnHold    ParticipantState = "on_hold"
	ParticipantStateInTransit ParticipantState = "in_transit"
)

// ConnectionState represents the state of the connection
type ConnectionState string

const (
	ConnectionStateDisconnected ConnectionState = "disconnected"
	ConnectionStateConnecting   ConnectionState = "connecting"
	ConnectionStateConnected    ConnectionState = "connected"
	ConnectionStateReconnecting ConnectionState = "reconnecting"
	ConnectionStateFailed       ConnectionState = "failed"
)

// ConnectionQuality represents the quality of the connection
type ConnectionQuality string

const (
	ConnectionQualityExcellent ConnectionQuality = "excellent"
	ConnectionQualityGood      ConnectionQuality = "good"
	ConnectionQualityFair      ConnectionQuality = "fair"
	ConnectionQualityPoor      ConnectionQuality = "poor"
)

// ClientOptions represents options for creating a new client
type ClientOptions struct {
	URL               string
	Token             string
	AutoConnect       bool
	Reconnect         bool
	ReconnectAttempts int
	ReconnectDelay    time.Duration
}

// JoinOptions represents options for joining a room
type JoinOptions struct {
	Audio    bool
	Video    bool
	Role     RoomRole
	Metadata map[string]interface{}
}

// ScreenShareOptions represents options for screen sharing
type ScreenShareOptions struct {
	FrameRate int
	Width     int
	Height    int
}

// RoomInfo represents information about a room
type RoomInfo struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	MaxUsers int                    `json:"maxUsers,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ParticipantInfo represents information about a participant
type ParticipantInfo struct {
	ID       string                 `json:"id"`
	UserID   string                 `json:"userId"`
	Role     RoomRole               `json:"role"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	JoinedAt int64                  `json:"joinedAt"`
}

// TrackInfo represents information about a track
type TrackInfo struct {
	ID            string      `json:"id"`
	Kind          TrackKind   `json:"kind"`
	Source        TrackSource `json:"source"`
	Status        TrackStatus `json:"status"`
	ParticipantID string      `json:"participantId"`
}

// WebSocket message types
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// Request/Response types
type JoinRoomRequest struct {
	RoomID   string                 `json:"roomId"`
	UserID   string                 `json:"userId"`
	Role     RoomRole               `json:"role"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type JoinRoomResponse struct {
	ParticipantID string            `json:"participantId"`
	Room          RoomInfo          `json:"room"`
	Participants  []ParticipantInfo `json:"participants"`
}

type PublishTrackRequest struct {
	ParticipantID string    `json:"participantId"`
	Track         TrackInfo `json:"track"`
	SDPOffer      string    `json:"sdpOffer"`
}

type PublishTrackResponse struct {
	TrackID   string `json:"trackId"`
	SDPAnswer string `json:"sdpAnswer"`
}

type SubscribeTrackRequest struct {
	ParticipantID string `json:"participantId"`
	TrackID       string `json:"trackId"`
}

type SubscribeTrackResponse struct {
	SDPOffer string `json:"sdpOffer"`
}

type ICECandidateMessage struct {
	Candidate interface{} `json:"candidate"`
	TrackID   string      `json:"trackId,omitempty"`
}
