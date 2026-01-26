package zenc

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v4"
)

// Room represents a streaming room
type Room struct {
	ID               string
	Name             string
	Participants     map[string]*Participant
	LocalParticipant *LocalParticipant
	State            ParticipantState

	signalingClient *SignalingClient
	publisherPC     *webrtc.PeerConnection
	subscriberPCs   map[string]*webrtc.PeerConnection

	mu sync.RWMutex
}

// NewRoom creates a new room
func NewRoom(roomInfo RoomInfo, participantInfo ParticipantInfo, signalingClient *SignalingClient) *Room {
	room := &Room{
		ID:               roomInfo.ID,
		Name:             roomInfo.Name,
		Participants:     make(map[string]*Participant),
		LocalParticipant: NewLocalParticipant(participantInfo),
		signalingClient:  signalingClient,
		subscriberPCs:    make(map[string]*webrtc.PeerConnection),
		State:            ParticipantStateActive,
	}

	room.setupSignalingHandlers()

	return room
}

func (r *Room) setupSignalingHandlers() {
	// Participant joined
	r.signalingClient.On("participant.joined", func(data interface{}) {
		jsonData, _ := json.Marshal(data)
		var info ParticipantInfo
		if err := json.Unmarshal(jsonData, &info); err != nil {
			return
		}

		participant := NewParticipant(info)
		r.mu.Lock()
		r.Participants[participant.ID] = participant
		r.mu.Unlock()
	})

	// Participant left
	r.signalingClient.On("participant.left", func(data interface{}) {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return
		}

		participantID, ok := dataMap["participantId"].(string)
		if !ok {
			return
		}

		r.mu.Lock()
		delete(r.Participants, participantID)
		r.mu.Unlock()
	})

	// Track published
	r.signalingClient.On("track.published", func(data interface{}) {
		jsonData, _ := json.Marshal(data)
		var trackInfo TrackInfo
		if err := json.Unmarshal(jsonData, &trackInfo); err != nil {
			return
		}

		r.mu.RLock()
		participant, ok := r.Participants[trackInfo.ParticipantID]
		r.mu.RUnlock()

		if ok {
			track := NewRemoteTrack(trackInfo.ID, trackInfo.Kind, trackInfo.Source)
			participant.AddTrack(track)
		}
	})

	// Track unpublished
	r.signalingClient.On("track.unpublished", func(data interface{}) {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return
		}

		trackID, _ := dataMap["trackId"].(string)
		participantID, _ := dataMap["participantId"].(string)

		r.mu.RLock()
		participant, ok := r.Participants[participantID]
		r.mu.RUnlock()

		if ok {
			participant.RemoveTrack(trackID)
		}
	})
}

// PublishTrack publishes a local track
func (r *Room) PublishTrack(track *LocalTrack) error {
	if r.LocalParticipant == nil {
		return fmt.Errorf("not joined to room")
	}

	// Create publisher peer connection if needed
	if r.publisherPC == nil {
		config := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}

		pc, err := webrtc.NewPeerConnection(config)
		if err != nil {
			return fmt.Errorf("failed to create peer connection: %w", err)
		}

		// Handle ICE candidates
		pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate != nil {
				r.signalingClient.Send(WSMessage{
					Type: "ice.candidate",
					Data: ICECandidateMessage{
						Candidate: candidate.ToJSON(),
						TrackID:   track.ID,
					},
				})
			}
		})

		r.publisherPC = pc
	}

	// Add track to peer connection
	if _, err := r.publisherPC.AddTrack(track.GetTrack()); err != nil {
		return fmt.Errorf("failed to add track: %w", err)
	}

	// Create offer
	offer, err := r.publisherPC.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}

	if err := r.publisherPC.SetLocalDescription(offer); err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	// Send publish request
	request := PublishTrackRequest{
		ParticipantID: r.LocalParticipant.ID,
		Track: TrackInfo{
			ID:            track.ID,
			Kind:          track.Kind,
			Source:        track.Source,
			Status:        track.Status,
			ParticipantID: r.LocalParticipant.ID,
		},
		SDPOffer: offer.SDP,
	}

	responseChan := make(chan PublishTrackResponse, 1)
	r.signalingClient.Once("track.published.response", func(data interface{}) {
		jsonData, _ := json.Marshal(data)
		var response PublishTrackResponse
		if err := json.Unmarshal(jsonData, &response); err != nil {
			return
		}
		responseChan <- response
	})

	if err := r.signalingClient.Send(WSMessage{
		Type: "track.publish",
		Data: request,
	}); err != nil {
		return err
	}

	// Wait for response
	response := <-responseChan

	// Set remote description
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  response.SDPAnswer,
	}

	if err := r.publisherPC.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	r.LocalParticipant.AddPublishedTrack(track)

	return nil
}

// UnpublishTrack unpublishes a local track
func (r *Room) UnpublishTrack(track *LocalTrack) error {
	if r.LocalParticipant == nil {
		return fmt.Errorf("not joined to room")
	}

	r.signalingClient.Send(WSMessage{
		Type: "track.unpublish",
		Data: map[string]string{"trackId": track.ID},
	})

	r.LocalParticipant.RemovePublishedTrack(track.ID)
	track.Stop()

	return nil
}

// Subscribe subscribes to a remote track
func (r *Room) Subscribe(trackID string) (*RemoteTrack, error) {
	if r.LocalParticipant == nil {
		return nil, fmt.Errorf("not joined to room")
	}

	// Find the track
	var remoteTrack *RemoteTrack
	r.mu.RLock()
	for _, participant := range r.Participants {
		if track, ok := participant.GetTrack(trackID); ok {
			remoteTrack = track
			break
		}
	}
	r.mu.RUnlock()

	if remoteTrack == nil {
		return nil, fmt.Errorf("track not found: %s", trackID)
	}

	// Create subscriber peer connection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	// Handle ICE candidates
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			r.signalingClient.Send(WSMessage{
				Type: "ice.candidate",
				Data: ICECandidateMessage{
					Candidate: candidate.ToJSON(),
					TrackID:   trackID,
				},
			})
		}
	})

	// Handle incoming tracks
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		remoteTrack.SetTrack(track)
	})

	r.mu.Lock()
	r.subscriberPCs[trackID] = pc
	r.mu.Unlock()

	// Send subscribe request
	request := SubscribeTrackRequest{
		ParticipantID: r.LocalParticipant.ID,
		TrackID:       trackID,
	}

	if err := r.signalingClient.Send(WSMessage{
		Type: "track.subscribe",
		Data: request,
	}); err != nil {
		return nil, err
	}

	return remoteTrack, nil
}

// Unsubscribe unsubscribes from a remote track
func (r *Room) Unsubscribe(trackID string) error {
	r.signalingClient.Send(WSMessage{
		Type: "track.unsubscribe",
		Data: map[string]string{"trackId": trackID},
	})

	r.mu.Lock()
	if pc, ok := r.subscriberPCs[trackID]; ok {
		pc.Close()
		delete(r.subscriberPCs, trackID)
	}
	r.mu.Unlock()

	return nil
}

// Leave leaves the room
func (r *Room) Leave() error {
	// Unpublish all tracks
	if r.LocalParticipant != nil {
		for _, track := range r.LocalParticipant.PublishedTracks {
			track.Stop()
		}
	}

	// Close all peer connections
	if r.publisherPC != nil {
		r.publisherPC.Close()
	}

	r.mu.Lock()
	for _, pc := range r.subscriberPCs {
		pc.Close()
	}
	r.subscriberPCs = make(map[string]*webrtc.PeerConnection)
	r.mu.Unlock()

	// Send leave request
	r.signalingClient.Send(WSMessage{
		Type: "room.leave",
		Data: map[string]string{"roomId": r.ID},
	})

	return nil
}

// Hold puts the room on hold
func (r *Room) Hold() error {
	if r.LocalParticipant == nil {
		return fmt.Errorf("not connected to room")
	}

	// Pause all local tracks
	for _, track := range r.LocalParticipant.PublishedTracks {
		if err := track.Hold(); err != nil {
			return fmt.Errorf("failed to hold track: %w", err)
		}
	}

	// Send hold message to server
	if err := r.signalingClient.Send(WSMessage{
		Type: "room.hold",
		Data: map[string]string{
			"roomId":        r.ID,
			"participantId": r.LocalParticipant.ID,
		},
	}); err != nil {
		return err
	}

	r.State = ParticipantStateOnHold
	return nil
}

// Resume resumes the room from hold
func (r *Room) Resume() error {
	if r.LocalParticipant == nil {
		return fmt.Errorf("not connected to room")
	}

	// Send resume message to server
	if err := r.signalingClient.Send(WSMessage{
		Type: "room.resume",
		Data: map[string]string{
			"roomId":        r.ID,
			"participantId": r.LocalParticipant.ID,
		},
	}); err != nil {
		return err
	}

	// Resume all tracks
	for _, track := range r.LocalParticipant.PublishedTracks {
		if err := track.Resume(); err != nil {
			return fmt.Errorf("failed to resume track: %w", err)
		}
	}

	r.State = ParticipantStateActive
	return nil
}

// IsOnHold returns whether the room is on hold
func (r *Room) IsOnHold() bool {
	return r.State == ParticipantStateOnHold
}
