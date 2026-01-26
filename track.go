package zenc

import (
	"github.com/pion/webrtc/v4"
)

// Track represents a media track
type Track struct {
	ID     string
	Kind   TrackKind
	Source TrackSource
	Status TrackStatus
}

// LocalTrack represents a local media track
type LocalTrack struct {
	Track
	track *webrtc.TrackLocalStaticRTP
}

// NewLocalTrack creates a new local track
func NewLocalTrack(id string, kind TrackKind, source TrackSource, track *webrtc.TrackLocalStaticRTP) *LocalTrack {
	return &LocalTrack{
		Track: Track{
			ID:     id,
			Kind:   kind,
			Source: source,
			Status: TrackStatusActive,
		},
		track: track,
	}
}

// GetTrack returns the underlying WebRTC track
func (t *LocalTrack) GetTrack() *webrtc.TrackLocalStaticRTP {
	return t.track
}

// Hold pauses the track
func (t *LocalTrack) Hold() error {
	t.Status = TrackStatusPaused
	return nil
}

// Resume resumes the track
func (t *LocalTrack) Resume() error {
	t.Status = TrackStatusActive
	return nil
}

// Stop stops the track
func (t *LocalTrack) Stop() error {
	t.Status = TrackStatusStopped
	return nil
}

// RemoteTrack represents a remote media track
type RemoteTrack struct {
	Track
	track *webrtc.TrackRemote
}

// NewRemoteTrack creates a new remote track
func NewRemoteTrack(id string, kind TrackKind, source TrackSource) *RemoteTrack {
	return &RemoteTrack{
		Track: Track{
			ID:     id,
			Kind:   kind,
			Source: source,
			Status: TrackStatusActive,
		},
	}
}

// SetTrack sets the underlying WebRTC track
func (t *RemoteTrack) SetTrack(track *webrtc.TrackRemote) {
	t.track = track
}

// GetTrack returns the underlying WebRTC track
func (t *RemoteTrack) GetTrack() *webrtc.TrackRemote {
	return t.track
}

// Hold pauses the track
func (t *RemoteTrack) Hold() error {
	t.Status = TrackStatusPaused
	return nil
}

// Resume resumes the track
func (t *RemoteTrack) Resume() error {
	t.Status = TrackStatusActive
	return nil
}

// Stop stops the track
func (t *RemoteTrack) Stop() error {
	t.Status = TrackStatusStopped
	return nil
}
