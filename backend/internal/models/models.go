package models

import "time"

type Session struct {
	ID            string     `gorm:"primaryKey;size:36" json:"id"`
	Name          string     `gorm:"size:128;not null" json:"name"`
	IsTitleLocked bool       `gorm:"default:false;not null" json:"isTitleLocked"`
	ModelConfigID *string    `gorm:"size:36" json:"modelConfigId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"index" json:"updatedAt"`
	DeletedAt     *time.Time `gorm:"index" json:"-"`
}

type Message struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	SessionID string    `gorm:"size:36;index;not null" json:"sessionId"`
	Role      string    `gorm:"size:16;index;not null" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`
}

// ToolCallRecord 持久化一次工具调用完整链路，支持历史会话回放工具详情。
type ToolCallRecord struct {
	ID                 string     `gorm:"primaryKey;size:36" json:"id"`
	SessionID          string     `gorm:"size:36;index;not null;uniqueIndex:idx_tool_call_request,priority:1" json:"sessionId"`
	RequestID          string     `gorm:"size:36;index;not null;uniqueIndex:idx_tool_call_request,priority:2" json:"requestId"`
	AssistantMessageID *string    `gorm:"size:36;index" json:"assistantMessageId,omitempty"`
	ToolCallID         string     `gorm:"size:128;index;not null;uniqueIndex:idx_tool_call_request,priority:3" json:"toolCallId"`
	ToolName           string     `gorm:"size:128;not null" json:"toolName"`
	Command            string     `gorm:"size:128;not null" json:"command"`
	ParamsJSON         string     `gorm:"type:text;not null" json:"paramsJson"`
	Status             string     `gorm:"size:32;index;not null" json:"status"`
	RequiresApproval   bool       `gorm:"not null;default:false" json:"requiresApproval"`
	Output             string     `gorm:"type:text" json:"output,omitempty"`
	Error              string     `gorm:"type:text" json:"error,omitempty"`
	StartedAt          time.Time  `gorm:"index;not null" json:"startedAt"`
	FinishedAt         *time.Time `gorm:"index" json:"finishedAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type AppSetting struct {
	Key       string    `gorm:"primaryKey;size:64" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type LLMConfig struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	BaseURL   string    `gorm:"size:512;not null" json:"baseUrl"`
	APIKey    string    `gorm:"size:512;not null" json:"apiKey"`
	Model     string    `gorm:"size:128;not null" json:"model"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type MCPConfig struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Config    string    `gorm:"type:text;not null" json:"config"`
	IsEnabled bool      `gorm:"default:true;not null" json:"isEnabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Skill struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Name         string    `gorm:"size:64;not null;uniqueIndex" json:"name"`
	RelativePath string    `gorm:"size:512;not null" json:"relativePath"`
	Description  string    `gorm:"type:text;not null" json:"description"`
	UploadedAt   time.Time `gorm:"index;not null" json:"uploadedAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
