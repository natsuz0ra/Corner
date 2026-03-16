package platforms

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"slimebot/backend/internal/models"
	"slimebot/backend/internal/services"
)

type Dispatcher struct {
	chat platformChatService
}

type platformChatService interface {
	EnsureMessagePlatformSession() (*models.Session, error)
	ResolvePlatformModel() (string, error)
	HandleChatStream(
		ctx context.Context,
		sessionID string,
		requestID string,
		content string,
		modelID string,
		callbacks services.AgentCallbacks,
	) (*services.ChatStreamResult, error)
}

func NewDispatcher(chat platformChatService) *Dispatcher {
	return &Dispatcher{chat: chat}
}

// HandleInbound 将平台消息转发到统一对话引擎，并把结果回传到平台。
func (d *Dispatcher) HandleInbound(ctx context.Context, message InboundMessage, sender OutboundSender) error {
	if d == nil || d.chat == nil {
		return fmt.Errorf("dispatcher 未初始化")
	}
	content := strings.TrimSpace(message.Text)
	if content == "" {
		return nil
	}
	chatID := strings.TrimSpace(message.ChatID)
	if chatID == "" {
		return fmt.Errorf("chat id 为空")
	}

	session, err := d.chat.EnsureMessagePlatformSession()
	if err != nil {
		return err
	}
	modelID, err := d.chat.ResolvePlatformModel()
	if err != nil {
		return err
	}

	callbacks := services.AgentCallbacks{
		OnChunk: func(_ string) error {
			return nil
		},
		OnToolCallStart: func(req services.ApprovalRequest) error {
			summary := formatToolStartSummary(req)
			return sender.SendText(chatID, summary)
		},
		// 平台侧暂不支持审批交互，先明确拒绝并提示在 Web 端操作。
		WaitApproval: func(_ context.Context, _ string) (*services.ApprovalResponse, error) {
			_ = sender.SendText(chatID, "检测到需要审批的工具操作，当前 Telegram 会话不支持审批，已自动拒绝。请在 Web 端完成审批后重试。")
			return &services.ApprovalResponse{Approved: false}, nil
		},
		OnToolCallResult: func(result services.ToolCallResult) error {
			return sender.SendText(chatID, formatToolResultSummary(result))
		},
	}

	streamResult, err := d.chat.HandleChatStream(
		ctx,
		session.ID,
		uuid.NewString(),
		content,
		modelID,
		callbacks,
	)
	if err != nil {
		return err
	}
	if streamResult == nil {
		return nil
	}
	answer := strings.TrimSpace(streamResult.Answer)
	if answer == "" {
		answer = "模型没有返回内容。"
	}
	return sender.SendText(chatID, answer)
}

func formatToolStartSummary(req services.ApprovalRequest) string {
	params := make([]string, 0, len(req.Params))
	for key, value := range req.Params {
		v := strings.TrimSpace(value)
		if v == "" {
			continue
		}
		if len(v) > 80 {
			v = v[:80] + "..."
		}
		params = append(params, fmt.Sprintf("%s=%s", key, v))
		if len(params) >= 3 {
			break
		}
	}
	if len(params) == 0 {
		return fmt.Sprintf("工具执行开始：%s", req.ToolName)
	}
	return fmt.Sprintf("工具执行开始：%s（%s）", req.ToolName, strings.Join(params, ", "))
}

func formatToolResultSummary(result services.ToolCallResult) string {
	statusText := "失败"
	if strings.EqualFold(strings.TrimSpace(result.Status), "completed") {
		statusText = "成功"
	}
	if strings.TrimSpace(result.Error) == "" {
		return fmt.Sprintf("工具执行%s：%s", statusText, result.ToolName)
	}
	return fmt.Sprintf("工具执行%s：%s（%s）", statusText, result.ToolName, strings.TrimSpace(result.Error))
}
