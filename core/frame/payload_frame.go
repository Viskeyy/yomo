package frame

import (
	"github.com/yomorun/y3"
)

// Tag is used for representing user defined data, system use Tag to route data to different handlers.
type Tag uint32

// PayloadFrame is a Y3 encoded bytes, .
type PayloadFrame struct {
	Tag      Tag    // Tag is defined by user
	Carriage []byte // Carriage is user's payload
}

// NewPayloadFrame creates a new PayloadFrame with a given TagID of user's data
func NewPayloadFrame(tag Tag) *PayloadFrame {
	return &PayloadFrame{
		Tag: tag,
	}
}

// SetCarriage sets the user's raw data
func (m *PayloadFrame) SetCarriage(buf []byte) *PayloadFrame {
	m.Carriage = buf
	return m
}

// Encode to Y3 encoded bytes
func (m *PayloadFrame) Encode() []byte {
	tag := y3.NewPrimitivePacketEncoder(byte(TagOfPayloadDataTag))
	tag.SetUInt32Value(uint32(m.Tag))

	carriage := y3.NewPrimitivePacketEncoder(byte(TagOfPayloadCarriage))
	carriage.SetBytesValue(m.Carriage)

	payload := y3.NewNodePacketEncoder(byte(TagOfPayloadFrame))
	payload.AddPrimitivePacket(tag)
	payload.AddPrimitivePacket(carriage)

	return payload.Encode()
}

// DecodeToPayloadFrame decodes Y3 encoded bytes to PayloadFrame
func DecodeToPayloadFrame(buf []byte) (*PayloadFrame, error) {
	nodeBlock := y3.NodePacket{}
	_, err := y3.DecodeToNodePacket(buf, &nodeBlock)
	if err != nil {
		return nil, err
	}

	payload := &PayloadFrame{}
	if p, ok := nodeBlock.PrimitivePackets[byte(TagOfPayloadDataTag)]; ok {
		tag, err := p.ToUInt32()
		if err != nil {
			return nil, err
		}
		payload.Tag = Tag(tag)
	}

	if p, ok := nodeBlock.PrimitivePackets[byte(TagOfPayloadCarriage)]; ok {
		payload.Carriage = p.GetValBuf()
	}

	return payload, nil
}
