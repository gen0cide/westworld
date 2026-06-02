package action

import (
	"context"

	v235 "github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Appearance is a character-customization selection. The server derives the
// head/body SPRITE ids by adding 1 to HeadType/BodyType; the colours index the
// appearance palette. Male maps to the server's headRestrictions byte (1 = male,
// 0 = female).
type Appearance struct {
	Male          bool
	HeadType      int // 0-based; headSprite = HeadType+1
	BodyType      int // 0-based; bodySprite = BodyType+1
	HairColour    int
	TopColour     int
	TrouserColour int
	SkinColour    int
}

// DefaultAppearance is a valid default male character. It reproduces the values
// OpenRSC itself stores for a freshly-created account (headSprite=1, bodySprite=2,
// hair=2, top=8, trousers=14, skin=0), so the server's PlayerAppearance.isValid
// check passes and the account is not flagged suspicious.
func DefaultAppearance() Appearance {
	return Appearance{
		Male:          true,
		HeadType:      0,
		BodyType:      1,
		HairColour:    2,
		TopColour:     8,
		TrouserColour: 14,
		SkinColour:    0,
	}
}

// ConfirmAppearance sends OutPlayerAppearance (opcode 235) to confirm the
// character's appearance. This clears the server's "changing appearance" hold on
// a fresh account (it also removes the tutorial_appearance cache key server-side),
// after which the server begins streaming the world update packets (own position,
// NPCs, objects) it withholds while the appearance screen is up.
//
// Payload is 8 bytes (Payload235Parser case 235, read as plain bytes):
// headRestrictions, headType, bodyType, mustEqual2(=2), hairColour, topColour,
// trouserColour, skinColour.
//
// Only send this in response to SEND_APPEARANCE_SCREEN — the server flags a
// player as suspicious if it arrives while not changing appearance.
func ConfirmAppearance(ctx context.Context, conn *session.Conn, a Appearance) error {
	headRestrictions := byte(0)
	if a.Male {
		headRestrictions = 1
	}
	buf := v235.NewBuffer(8)
	buf.WriteByte(headRestrictions)
	buf.WriteByte(byte(a.HeadType))
	buf.WriteByte(byte(a.BodyType))
	buf.WriteByte(2) // mustEqual2 — the server rejects any other value here
	buf.WriteByte(byte(a.HairColour))
	buf.WriteByte(byte(a.TopColour))
	buf.WriteByte(byte(a.TrouserColour))
	buf.WriteByte(byte(a.SkinColour))
	return conn.Send(v235.OutPlayerAppearance, buf.Bytes())
}
