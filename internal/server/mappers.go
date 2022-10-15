package server

import (
	"github.com/kristina71/otus_project/internal/server/pb"
	"github.com/kristina71/otus_project/internal/storage"
)

func MapBannerToPb(banner storage.Banner) *pb.Banner {
	return &pb.Banner{
		Id:          banner.ID,
		Description: banner.Description,
	}
}

func MapSlotToPb(slot storage.Slot) *pb.Slot {
	return &pb.Slot{
		Id:          slot.ID,
		Description: slot.Description,
	}
}

func MapGroupToPb(group storage.SocialGroup) *pb.Group {
	return &pb.Group{
		Id:          group.ID,
		Description: group.Description,
	}
}
