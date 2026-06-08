package channelserver

import (
	"fmt"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfitem"
	ps "erupe-ce/common/pascalstring"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

// Distribution represents an item distribution event.
type Distribution struct {
	ID              uint32    `db:"id"`
	Deadline        time.Time `db:"deadline"`
	Rights          uint32    `db:"rights"`
	TimesAcceptable uint16    `db:"times_acceptable"`
	TimesAccepted   uint16    `db:"times_accepted"`
	MinHR           int16     `db:"min_hr"`
	MaxHR           int16     `db:"max_hr"`
	MinSR           int16     `db:"min_sr"`
	MaxSR           int16     `db:"max_sr"`
	MinGR           int16     `db:"min_gr"`
	MaxGR           int16     `db:"max_gr"`
	EventName       string    `db:"event_name"`
	Description     string    `db:"description"`
	Selection       bool      `db:"selection"`
}

func handleMsgMhfEnumerateDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateDistItem)

	bf := byteframe.NewByteFrame()
	itemDists, err := s.server.distRepo.List(s.charID, pkt.DistType)
	if err != nil {
		s.logger.Error("Failed to list item distributions", zap.Error(err))
	}

	bf.WriteUint16(uint16(len(itemDists)))
	for _, dist := range itemDists {
		bf.WriteUint32(dist.ID)
		bf.WriteUint32(uint32(dist.Deadline.Unix()))
		bf.WriteUint32(dist.Rights)
		bf.WriteUint16(dist.TimesAcceptable)
		bf.WriteUint16(dist.TimesAccepted)
		if s.server.erupeConfig.RealClientMode >= cfg.G9 {
			bf.WriteUint16(0) // Unk
		}
		bf.WriteInt16(dist.MinHR)
		bf.WriteInt16(dist.MaxHR)
		bf.WriteInt16(dist.MinSR)
		bf.WriteInt16(dist.MaxSR)
		bf.WriteInt16(dist.MinGR)
		bf.WriteInt16(dist.MaxGR)
		if s.server.erupeConfig.RealClientMode >= cfg.G7 {
			bf.WriteUint8(0) // Unk
		}
		if s.server.erupeConfig.RealClientMode >= cfg.G6 {
			bf.WriteUint16(0) // Unk
		}
		if s.server.erupeConfig.RealClientMode >= cfg.G8 {
			if dist.Selection {
				bf.WriteUint8(2) // Selection
			} else {
				bf.WriteUint8(0)
			}
		}
		if s.server.erupeConfig.RealClientMode >= cfg.G7 {
			bf.WriteUint16(0) // Unk
			bf.WriteUint16(0) // Unk
		}
		if s.server.erupeConfig.RealClientMode >= cfg.G10 {
			bf.WriteUint8(0) // Unk
		}
		ps.Uint8(bf, dist.EventName, true)
		k := 6
		if s.server.erupeConfig.RealClientMode >= cfg.G8 {
			k = 13
		}
		for i := 0; i < 6; i++ {
			for j := 0; j < k; j++ {
				bf.WriteUint8(0)
				bf.WriteUint32(0)
			}
		}
		if s.server.erupeConfig.RealClientMode >= cfg.Z2 {
			i := uint8(0)
			bf.WriteUint8(i)
			if i <= 10 {
				for j := uint8(0); j < i; j++ {
					bf.WriteUint32(0)
					bf.WriteUint32(0)
					bf.WriteUint32(0)
				}
			}
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// DistributionItem represents a single item in a distribution.
type DistributionItem struct {
	ItemType uint8  `db:"item_type"`
	ID       uint32 `db:"id"`
	ItemID   uint32 `db:"item_id"`
	Quantity uint32 `db:"quantity"`
}

func handleMsgMhfApplyDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyDistItem)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(pkt.DistributionID)
	distItems, err := s.server.distRepo.GetItems(pkt.DistributionID)
	if err != nil {
		s.logger.Error("Failed to get distribution items", zap.Error(err))
	}
	bf.WriteUint16(uint16(len(distItems)))
	for _, item := range distItems {
		bf.WriteUint8(item.ItemType)
		bf.WriteUint32(item.ItemID)
		bf.WriteUint32(item.Quantity)
		if s.server.erupeConfig.RealClientMode >= cfg.G8 {
			bf.WriteUint32(item.ID)
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireDistItem)

	if pkt.DistributionID == 0 {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	dist, distItems, err := s.server.distRepo.GetClaimable(pkt.DistributionID, s.charID, pkt.DistributionType)
	if err != nil {
		s.logger.Warn("Distribution is not claimable",
			zap.Uint32("distribution_id", pkt.DistributionID),
			zap.Uint8("distribution_type", pkt.DistributionType),
			zap.Error(err),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	hr, gr, err := distributionCharacterRanks(s)
	if err != nil {
		s.logger.Error("Failed to load character ranks for distribution claim", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if err := distributionRankAllowed(dist, hr, gr); err != nil {
		s.logger.Warn("Distribution rank check failed",
			zap.Uint32("distribution_id", pkt.DistributionID),
			zap.Uint16("hr", hr),
			zap.Uint16("gr", gr),
			zap.Error(err),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	for _, item := range distItems {
		if err := grantDistributionItem(s, item); err != nil {
			s.logger.Error("Failed to grant distribution item",
				zap.Uint32("distribution_id", pkt.DistributionID),
				zap.Uint8("item_type", item.ItemType),
				zap.Uint32("item_id", item.ItemID),
				zap.Uint32("quantity", item.Quantity),
				zap.Error(err),
			)
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	if err := s.server.distRepo.RecordAccepted(pkt.DistributionID, s.charID); err != nil {
		s.logger.Error("Failed to record accepted distribution", zap.Error(err))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func distributionCharacterRanks(s *Session) (uint16, uint16, error) {
	hr, err := s.server.charRepo.ReadInt(s.charID, "hr")
	if err != nil {
		return 0, 0, fmt.Errorf("read HR: %w", err)
	}
	gr, err := s.server.charRepo.ReadInt(s.charID, "gr")
	if err != nil {
		return 0, 0, fmt.Errorf("read GR: %w", err)
	}
	return uint16(hr), uint16(gr), nil
}

func distributionRankAllowed(dist Distribution, hr, gr uint16) error {
	if dist.MinHR >= 0 && hr < uint16(dist.MinHR) {
		return fmt.Errorf("HR %d below minimum %d", hr, dist.MinHR)
	}
	if dist.MaxHR >= 0 && hr > uint16(dist.MaxHR) {
		return fmt.Errorf("HR %d above maximum %d", hr, dist.MaxHR)
	}
	if dist.MinGR >= 0 && gr < uint16(dist.MinGR) {
		return fmt.Errorf("GR %d below minimum %d", gr, dist.MinGR)
	}
	if dist.MaxGR >= 0 && gr > uint16(dist.MaxGR) {
		return fmt.Errorf("GR %d above maximum %d", gr, dist.MaxGR)
	}
	if dist.MinSR >= 0 || dist.MaxSR >= 0 {
		return fmt.Errorf("SR-gated distributions are not supported")
	}
	return nil
}

func grantDistributionItem(s *Session, item DistributionItem) error {
	switch item.ItemType {
	case 0, 1, 2, 3, 4, 5, 6:
		if item.ItemID == 0 || item.ItemID > 0xffff || item.Quantity == 0 || item.Quantity > 0xffff {
			return fmt.Errorf("invalid equipment reward item_type=%d item_id=%d quantity=%d", item.ItemType, item.ItemID, item.Quantity)
		}
		for i := uint32(0); i < item.Quantity; i++ {
			if err := addWarehouseEquipmentErr(s, newDistributionEquipment(item.ItemType, uint16(item.ItemID))); err != nil {
				return fmt.Errorf("add equipment to gift box: %w", err)
			}
		}
	case 7:
		if item.ItemID == 0 || item.ItemID > 0xffff || item.Quantity == 0 || item.Quantity > 0xffff {
			return fmt.Errorf("invalid item reward item_id=%d quantity=%d", item.ItemID, item.Quantity)
		}
		if err := addWarehouseItemErr(s, mhfitem.MHFItemStack{
			Item:     mhfitem.MHFItem{ItemID: uint16(item.ItemID)},
			Quantity: uint16(item.Quantity),
		}); err != nil {
			return fmt.Errorf("add item to gift box: %w", err)
		}
	case 17:
		if err := addPointNetcafe(s, int(item.Quantity)); err != nil {
			return fmt.Errorf("add netcafe points: %w", err)
		}
	case 19:
		if err := s.server.userRepo.AddPremiumCoins(s.userID, item.Quantity); err != nil {
			return fmt.Errorf("add premium coins: %w", err)
		}
	case 20:
		if err := s.server.userRepo.AddTrialCoins(s.userID, item.Quantity); err != nil {
			return fmt.Errorf("add trial coins: %w", err)
		}
	case 21:
		if err := s.server.userRepo.AddFrontierPoints(s.userID, item.Quantity); err != nil {
			return fmt.Errorf("add frontier points: %w", err)
		}
	case 23:
		if item.Quantity > 0xffff {
			return fmt.Errorf("RP quantity too large: %d", item.Quantity)
		}
		saveData, err := GetCharacterSaveData(s, s.charID)
		if err != nil {
			return fmt.Errorf("load savedata for RP: %w", err)
		}
		saveData.RP += uint16(item.Quantity)
		if err := saveData.Save(s); err != nil {
			return fmt.Errorf("save RP: %w", err)
		}
	case 30, 31:
		// The client unlocks extra item/equipment box pages from the claim result.
		// Warehouse storage is already provisioned server-side.
	default:
		return fmt.Errorf("unsupported distribution item type: %d", item.ItemType)
	}
	return nil
}

func newDistributionEquipment(itemType uint8, itemID uint16) mhfitem.MHFEquipment {
	equipment := mhfitem.MHFEquipment{
		ItemType:    itemType,
		ItemID:      itemID,
		Level:       1,
		Decorations: make([]mhfitem.MHFItem, 3),
		Sigils:      make([]mhfitem.MHFSigil, 3),
	}
	for i := range equipment.Sigils {
		equipment.Sigils[i].Effects = make([]mhfitem.MHFSigilEffect, 3)
	}
	return equipment
}

func handleMsgMhfGetDistDescription(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetDistDescription)
	desc, err := s.server.distRepo.GetDescription(pkt.DistributionID)
	if err != nil {
		s.logger.Error("Error parsing item distribution description", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrame()
	ps.Uint16(bf, desc, true)
	ps.Uint16(bf, "", false)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
