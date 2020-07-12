package websocket

type (
	Message interface {
		GetMessage() []byte
	}

	internalMessage struct {
		payload []byte
	}
)

// GetMessage returns a message.
func (im *internalMessage) GetMessage() []byte {
	return im.payload
}
