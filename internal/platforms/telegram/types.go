package telegram

type getUpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []update `json:"result"`
}

type update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *message       `json:"message"`
	CallbackQuery *callbackQuery `json:"callback_query"`
}

type message struct {
	MessageID int64            `json:"message_id"`
	Chat      chat             `json:"chat"`
	Text      string           `json:"text"`
	Caption   string           `json:"caption"`
	Photo     []photoSize      `json:"photo"`
	Voice     *voiceAttachment `json:"voice"`
	Audio     *audioAttachment `json:"audio"`
	Document  *docAttachment   `json:"document"`
}

type photoSize struct {
	FileID   string `json:"file_id"`
	FileSize int64  `json:"file_size"`
}

type voiceAttachment struct {
	FileID   string `json:"file_id"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

type audioAttachment struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

type docAttachment struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

type chat struct {
	ID int64 `json:"id"`
}

type callbackQuery struct {
	ID      string   `json:"id"`
	From    user     `json:"from"`
	Message *message `json:"message"`
	Data    string   `json:"data"`
}

type user struct {
	ID int64 `json:"id"`
}

type getFileResponse struct {
	OK     bool          `json:"ok"`
	Result *telegramFile `json:"result"`
}

type telegramFile struct {
	FilePath string `json:"file_path"`
}

type mediaCandidate struct {
	Source         string
	ProviderFileID string
	Name           string
	MimeType       string
	SizeBytes      int64
}
