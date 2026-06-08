package channelserver

import (
	"encoding/binary"
	"errors"
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

// --- mockDistRepo ---

type mockDistRepo struct {
	distributions []Distribution
	listErr       error
	items         map[uint32][]DistributionItem
	itemsErr      error
	claimableDist map[uint32]Distribution
	description   string
	descErr       error
	recordedDist  uint32
	recordedChar  uint32
	recordErr     error
}

func (m *mockDistRepo) List(_ uint32, _ uint8) ([]Distribution, error) {
	return m.distributions, m.listErr
}

func (m *mockDistRepo) GetClaimable(distID, _ uint32, _ uint8) (Distribution, []DistributionItem, error) {
	if m.itemsErr != nil {
		return Distribution{}, nil, m.itemsErr
	}
	items := m.items[distID]
	dist := Distribution{
		ID:              distID,
		TimesAcceptable: 1,
		TimesAccepted:   0,
		MinHR:           -1,
		MaxHR:           -1,
		MinSR:           -1,
		MaxSR:           -1,
		MinGR:           -1,
		MaxGR:           -1,
	}
	if m.claimableDist != nil {
		if configured, ok := m.claimableDist[distID]; ok {
			dist = configured
		}
	}
	return dist, items, nil
}

func (m *mockDistRepo) GetItems(distID uint32) ([]DistributionItem, error) {
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}
	if m.items != nil {
		return m.items[distID], nil
	}
	return nil, nil
}

func (m *mockDistRepo) RecordAccepted(distID, charID uint32) error {
	m.recordedDist = distID
	m.recordedChar = charID
	return m.recordErr
}

func (m *mockDistRepo) GetDescription(_ uint32) (string, error) {
	return m.description, m.descErr
}

func parseSimpleAckError(t *testing.T, raw []byte) uint8 {
	t.Helper()
	if len(raw) < 8 {
		t.Fatalf("raw packet too short: %d bytes", len(raw))
	}
	return raw[7]
}

func TestHandleMsgMhfEnumerateDistItem_Empty(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.S6
	server.distRepo = &mockDistRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateDistItem{AckHandle: 100, DistType: 0}
	handleMsgMhfEnumerateDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, errCode, ackData := parseAckBufData(t, p.data)
		if errCode != 0 {
			t.Errorf("ErrorCode = %d, want 0", errCode)
		}
		count := binary.BigEndian.Uint16(ackData[:2])
		if count != 0 {
			t.Errorf("dist count = %d, want 0", count)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfEnumerateDistItem_WithDistributions(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.S6
	server.distRepo = &mockDistRepo{
		distributions: []Distribution{
			{
				ID:              1,
				Deadline:        time.Unix(1000000, 0),
				Rights:          0,
				TimesAcceptable: 1,
				TimesAccepted:   0,
				MinHR:           1,
				MaxHR:           999,
				EventName:       "Test",
			},
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateDistItem{AckHandle: 100, DistType: 0}
	handleMsgMhfEnumerateDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		count := binary.BigEndian.Uint16(ackData[:2])
		if count != 1 {
			t.Errorf("dist count = %d, want 1", count)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfApplyDistItem_Empty(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.S6
	server.distRepo = &mockDistRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfApplyDistItem{
		AckHandle:      100,
		DistributionID: 42,
	}
	handleMsgMhfApplyDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		// 4 (distID) + 2 (count=0) = 6
		distID := binary.BigEndian.Uint32(ackData[:4])
		if distID != 42 {
			t.Errorf("distID = %d, want 42", distID)
		}
		itemCount := binary.BigEndian.Uint16(ackData[4:6])
		if itemCount != 0 {
			t.Errorf("item count = %d, want 0", itemCount)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfApplyDistItem_WithItems(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.S6
	server.distRepo = &mockDistRepo{
		items: map[uint32][]DistributionItem{
			10: {
				{ItemType: 1, ID: 100, ItemID: 200, Quantity: 5},
				{ItemType: 2, ID: 101, ItemID: 300, Quantity: 3},
			},
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfApplyDistItem{
		AckHandle:      100,
		DistributionID: 10,
	}
	handleMsgMhfApplyDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		itemCount := binary.BigEndian.Uint16(ackData[4:6])
		if itemCount != 2 {
			t.Errorf("item count = %d, want 2", itemCount)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_ZeroID(t *testing.T) {
	server := createMockServer()
	server.distRepo = &mockDistRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:      100,
		DistributionID: 0,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Should respond")
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_RecordAccepted(t *testing.T) {
	server := createMockServer()
	distRepo := &mockDistRepo{
		items: map[uint32][]DistributionItem{
			5: {{ItemType: 7, ItemID: 100, Quantity: 5}},
		},
	}
	server.distRepo = distRepo
	server.charRepo = newMockCharacterRepo()
	server.houseRepo = newMockHouseRepoForItems()
	session := createMockSession(1, server)
	session.charID = 42

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:      100,
		DistributionID: 5,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	if distRepo.recordedDist != 5 {
		t.Errorf("recorded dist ID = %d, want 5", distRepo.recordedDist)
	}
	if distRepo.recordedChar != 42 {
		t.Errorf("recorded char ID = %d, want 42", distRepo.recordedChar)
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Should respond")
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_ItemRewardGiftBox(t *testing.T) {
	server := createMockServer()
	distRepo := &mockDistRepo{
		items: map[uint32][]DistributionItem{
			5: {{ItemType: 7, ItemID: 2210, Quantity: 10}},
		},
	}
	houseRepo := newMockHouseRepoForItems()
	server.distRepo = distRepo
	server.charRepo = newMockCharacterRepo()
	server.houseRepo = houseRepo
	session := createMockSession(1, server)
	session.charID = 42

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:        100,
		DistributionType: 1,
		DistributionID:   5,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	giftBox := houseRepo.setData[10]
	if len(giftBox) == 0 {
		t.Fatal("gift box was not updated")
	}
	bf := byteframe.NewByteFrameFromBytes(giftBox)
	count := bf.ReadUint16()
	bf.ReadUint16()
	if count != 1 {
		t.Fatalf("gift box item count = %d, want 1", count)
	}
	bf.ReadUint32()
	itemID := bf.ReadUint16()
	quantity := bf.ReadUint16()
	if itemID != 2210 || quantity != 10 {
		t.Errorf("gift box item = (%d, %d), want (2210, 10)", itemID, quantity)
	}
	if distRepo.recordedDist != 5 || distRepo.recordedChar != 42 {
		t.Errorf("accepted record = (%d, %d), want (5, 42)", distRepo.recordedDist, distRepo.recordedChar)
	}
	select {
	case p := <-session.sendPackets:
		if errCode := parseSimpleAckError(t, p.data); errCode != 0 {
			t.Errorf("ErrorCode = %d, want 0", errCode)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_RankGateFailsWithoutGrant(t *testing.T) {
	server := createMockServer()
	distRepo := &mockDistRepo{
		items: map[uint32][]DistributionItem{
			5: {{ItemType: 7, ItemID: 2210, Quantity: 10}},
		},
		claimableDist: map[uint32]Distribution{
			5: {
				ID:              5,
				TimesAcceptable: 1,
				MinHR:           5,
				MaxHR:           999,
				MinSR:           -1,
				MaxSR:           -1,
				MinGR:           -1,
				MaxGR:           -1,
			},
		},
	}
	charRepo := newMockCharacterRepo()
	charRepo.ints["hr"] = 4
	houseRepo := newMockHouseRepoForItems()
	server.distRepo = distRepo
	server.charRepo = charRepo
	server.houseRepo = houseRepo
	session := createMockSession(1, server)
	session.charID = 42

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:        100,
		DistributionType: 1,
		DistributionID:   5,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	if distRepo.recordedDist != 0 || distRepo.recordedChar != 0 {
		t.Errorf("accepted record = (%d, %d), want none", distRepo.recordedDist, distRepo.recordedChar)
	}
	if len(houseRepo.setData[10]) != 0 {
		t.Fatal("gift box should not be updated")
	}
	select {
	case p := <-session.sendPackets:
		if errCode := parseSimpleAckError(t, p.data); errCode != 1 {
			t.Errorf("ErrorCode = %d, want 1", errCode)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_UnsupportedRewardFailsWithoutRecord(t *testing.T) {
	server := createMockServer()
	distRepo := &mockDistRepo{
		items: map[uint32][]DistributionItem{
			5: {{ItemType: 99, Quantity: 1}},
		},
	}
	server.distRepo = distRepo
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(1, server)
	session.charID = 42

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:        100,
		DistributionType: 1,
		DistributionID:   5,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	if distRepo.recordedDist != 0 || distRepo.recordedChar != 0 {
		t.Errorf("accepted record = (%d, %d), want none", distRepo.recordedDist, distRepo.recordedChar)
	}
	select {
	case p := <-session.sendPackets:
		if errCode := parseSimpleAckError(t, p.data); errCode != 1 {
			t.Errorf("ErrorCode = %d, want 1", errCode)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfAcquireDistItem_RecordError(t *testing.T) {
	server := createMockServer()
	server.distRepo = &mockDistRepo{
		items: map[uint32][]DistributionItem{
			5: {{ItemType: 7, ItemID: 100, Quantity: 5}},
		},
		recordErr: errors.New("db error"),
	}
	server.charRepo = newMockCharacterRepo()
	server.houseRepo = newMockHouseRepoForItems()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireDistItem{
		AckHandle:      100,
		DistributionID: 5,
	}
	handleMsgMhfAcquireDistItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if errCode := parseSimpleAckError(t, p.data); errCode != 1 {
			t.Errorf("ErrorCode = %d, want 1", errCode)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfGetDistDescription_Success(t *testing.T) {
	server := createMockServer()
	server.distRepo = &mockDistRepo{description: "Test event description"}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetDistDescription{
		AckHandle:      100,
		DistributionID: 1,
	}
	handleMsgMhfGetDistDescription(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, errCode, ackData := parseAckBufData(t, p.data)
		if errCode != 0 {
			t.Errorf("ErrorCode = %d, want 0", errCode)
		}
		if len(ackData) == 0 {
			t.Fatal("AckData should not be empty")
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfGetDistDescription_Error(t *testing.T) {
	server := createMockServer()
	server.distRepo = &mockDistRepo{descErr: errors.New("not found")}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetDistDescription{
		AckHandle:      100,
		DistributionID: 999,
	}
	handleMsgMhfGetDistDescription(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, errCode, ackData := parseAckBufData(t, p.data)
		if errCode != 0 {
			t.Errorf("ErrorCode = %d, want 0 (still buf succeed)", errCode)
		}
		if len(ackData) != 4 {
			t.Errorf("AckData len = %d, want 4 (fallback)", len(ackData))
		}
	default:
		t.Fatal("No response queued")
	}
}
