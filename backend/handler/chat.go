package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"resume-agent/persona"
	"resume-agent/service"
)

type ChatHandler struct {
	chatService *service.ChatService
	personaLoader *persona.Loader
}

func NewChatHandler(chatService *service.ChatService, personaLoader *persona.Loader) *ChatHandler {
	return &ChatHandler{
		chatService:  chatService,
		personaLoader: personaLoader,
	}
}

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

type InitResponse struct {
	Name    string `json:"name"`
	Welcome string `json:"welcome"`
}

// InitHandler 返回初始化信息
func (h *ChatHandler) InitHandler(c *gin.Context) {
	persona, _ := h.personaLoader.Load()
	lines := strings.Split(persona, "\n")
	var name string
	for _, line := range lines {
		if strings.HasPrefix(line, "- 姓名:") || strings.HasPrefix(line, "- 姓名：") {
			name = strings.TrimPrefix(line, "- 姓名:")
			name = strings.TrimPrefix(name, "- 姓名：")
			name = strings.TrimSpace(name)
			break
		}
	}
	if name == "" || name == "[你的名字]" {
		name = "AI助手"
	}

	c.JSON(http.StatusOK, InitResponse{
		Name:    name + "的AI助手",
		Welcome: "你好！我是" + name + "，很高兴认识你。有什么想了解的吗？",
	})
}

// ChatHandler 处理普通聊天请求
func (h *ChatHandler) ChatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息不能为空"})
		return
	}

	reply, err := h.chatService.Chat(c.Request.Context(), req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{Reply: reply})
}

// StreamHandler 处理流式聊天请求
func (h *ChatHandler) StreamHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息不能为空"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 设置 SSE 友好的格式
	c.Stream(func(w io.Writer) bool {
		err := h.chatService.ChatStream(c.Request.Context(), req.Message, func(chunk string) {
			data := fmt.Sprintf("data: %s\n\n", chunk)
			c.Writer.Write([]byte(data))
			c.Writer.Flush()
		})
		if err != nil {
			c.Writer.Write([]byte(fmt.Sprintf("data: [ERROR] %s\n\n", err.Error())))
		}
		c.Writer.Write([]byte("data: [DONE]\n\n"))
		return false
	})
}

// ReloadHandler 重新加载人物设定
func (h *ChatHandler) ReloadHandler(c *gin.Context) {
	persona, err := h.personaLoader.Reload()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.chatService.UpdatePersona(persona)
	h.chatService.ClearHistory()

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
