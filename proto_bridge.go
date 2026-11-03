package zenc

import (
	pb "github.com/cyper98/zenc-protocol/generated/go"
)

// ToProto converts a SDK RoomRole to protobuf RoomRole
func (r RoomRole) ToProto() pb.RoomRole {
	switch r {
	case RolePublisher:
		return pb.RoomRole_ROOM_ROLE_PUBLISHER
	case RoleSubscriber:
		return pb.RoomRole_ROOM_ROLE_SUBSCRIBER
	case RoleModerator:
		return pb.RoomRole_ROOM_ROLE_MODERATOR
	default:
		return pb.RoomRole_ROOM_ROLE_UNSPECIFIED
	}
}

// RoomRoleFromProto converts protobuf RoomRole to SDK RoomRole
func RoomRoleFromProto(r pb.RoomRole) RoomRole {
	switch r {
	case pb.RoomRole_ROOM_ROLE_PUBLISHER:
		return RolePublisher
	case pb.RoomRole_ROOM_ROLE_SUBSCRIBER:
		return RoleSubscriber
	case pb.RoomRole_ROOM_ROLE_MODERATOR:
		return RoleModerator
	default:
		return RoleSubscriber
	}
}

// ToProto converts a SDK TrackKind to protobuf TrackKind
func (k TrackKind) ToProto() pb.TrackKind {
	switch k {
	case TrackKindAudio:
		return pb.TrackKind_TRACK_KIND_AUDIO
	case TrackKindVideo:
		return pb.TrackKind_TRACK_KIND_VIDEO
	default:
		return pb.TrackKind_TRACK_KIND_UNSPECIFIED
	}
}

// ToProto converts a SDK TrackSource to protobuf TrackSource
func (s TrackSource) ToProto() pb.TrackSource {
	switch s {
	case TrackSourceCamera:
		return pb.TrackSource_TRACK_SOURCE_CAMERA
	case TrackSourceScreen:
		return pb.TrackSource_TRACK_SOURCE_SCREEN
	case TrackSourceMicrophone:
		return pb.TrackSource_TRACK_SOURCE_MICROPHONE
	default:
		return pb.TrackSource_TRACK_SOURCE_UNSPECIFIED
	}
}

// ToProto converts a SDK TrackStatus to protobuf TrackStatus
func (s TrackStatus) ToProto() pb.TrackStatus {
	switch s {
	case TrackStatusActive:
		return pb.TrackStatus_TRACK_STATUS_ACTIVE
	case TrackStatusPaused:
		return pb.TrackStatus_TRACK_STATUS_PAUSED
	case TrackStatusStopped:
		return pb.TrackStatus_TRACK_STATUS_STOPPED
	default:
		return pb.TrackStatus_TRACK_STATUS_UNSPECIFIED
	}
}

