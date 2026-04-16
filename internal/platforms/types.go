package platforms

// InboundMessage is the normalized inbound shape from a platform.
type InboundMessage struct {
	Platform      string
	ChatID        string
	Text          string
	Attachments   []InboundAttachment
	AttachmentIDs []string
}

// InboundAttachment holds basic metadata for a platform-side attachment.
type InboundAttachment struct {
	Source         string
	ProviderFileID string
	Name           string
	MimeType       string
	SizeBytes      int64
	Category       string
}

// OutboundSender abstracts platform send for reuse inside the dispatcher.
type OutboundSender interface {
	SendText(chatID string, text string) error
	SendApprovalKeyboard(chatID string, text string, approveData string, rejectData string) error
}
