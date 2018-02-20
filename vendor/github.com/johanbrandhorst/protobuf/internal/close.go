package internal

const closeMessage = "clos"

// IsCloseMessage says whether the byte sequence is a close message
func IsCloseMessage(msg []byte) bool {
	return string(msg) == closeMessage
}

// FormatCloseMessage gets the CloseMessage as a byte sequence
func FormatCloseMessage() []byte {
	return []byte(closeMessage)
}
