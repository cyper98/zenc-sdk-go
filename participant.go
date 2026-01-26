package zenc

import "sync"

// Participant represents a participant in a room
type Participant struct {
	ID                string
	UserID            string
	Role              RoomRole
	Tracks            map[string]*RemoteTrack
	ConnectionQuality ConnectionQuality
	Metadata          map[string]interface{}
	JoinedAt          int64

	mu sync.RWMutex
}

// NewParticipant creates a new participant
func NewParticipant(info ParticipantInfo) *Participant {
	return &Participant{
		ID:                info.ID,
		UserID:            info.UserID,
		Role:              info.Role,
		Metadata:          info.Metadata,
		JoinedAt:          info.JoinedAt,
		Tracks:            make(map[string]*RemoteTrack),
		ConnectionQuality: ConnectionQualityGood,
	}
}

// AddTrack adds a track to the participant
func (p *Participant) AddTrack(track *RemoteTrack) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Tracks[track.ID] = track
}

// RemoveTrack removes a track from the participant
func (p *Participant) RemoveTrack(trackID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.Tracks, trackID)
}

// GetTrack gets a track by ID
func (p *Participant) GetTrack(trackID string) (*RemoteTrack, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	track, ok := p.Tracks[trackID]
	return track, ok
}

// LocalParticipant represents the local participant
type LocalParticipant struct {
	*Participant
	PublishedTracks map[string]*LocalTrack
}

// NewLocalParticipant creates a new local participant
func NewLocalParticipant(info ParticipantInfo) *LocalParticipant {
	return &LocalParticipant{
		Participant:     NewParticipant(info),
		PublishedTracks: make(map[string]*LocalTrack),
	}
}

// AddPublishedTrack adds a published track
func (p *LocalParticipant) AddPublishedTrack(track *LocalTrack) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.PublishedTracks[track.ID] = track
}

// RemovePublishedTrack removes a published track
func (p *LocalParticipant) RemovePublishedTrack(trackID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.PublishedTracks, trackID)
}
