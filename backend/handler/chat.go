package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"resume-agent/config"
	"resume-agent/persona"
	"resume-agent/service"
)

type ChatHandler struct {
	chatService    *service.ChatService
	personaLoader *persona.Loader
	accessPassword string
	personaData   string
}

func NewChatHandler(chatService *service.ChatService, personaLoader *persona.Loader, cfg *config.Config) *ChatHandler {
	persona, _ := personaLoader.Load()
	return &ChatHandler{
		chatService:    chatService,
		personaLoader:  personaLoader,
		accessPassword: cfg.AccessPassword,
		personaData:    persona,
	}
}

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

type InitResponse struct {
	Name         string `json:"name"`
	Welcome      string `json:"welcome"`
	NeedPassword bool   `json:"need_password"`
}

type VerifyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// InitHandler 返回初始化信息
func (h *ChatHandler) InitHandler(c *gin.Context) {
	lines := strings.Split(h.personaData, "\n")
	var name string
	for _, line := range lines {
		if strings.Contains(line, "姓名") && len(line) < 50 {
			parts := strings.Split(line, "**")
			if len(parts) >= 3 {
				name = strings.TrimSpace(parts[2])
			} else {
				parts = strings.Split(line, ":")
				if len(parts) >= 2 {
					name = strings.TrimSpace(parts[len(parts)-1])
				}
			}
			name = strings.Trim(name, "** \t")
			if name != "" && name != "[你的名字]" {
				break
			}
		}
	}
	if name == "" {
		name = "AI助手"
	}

	c.JSON(http.StatusOK, InitResponse{
		Name:         name,
		Welcome:      "你好！我是" + name + "，很高兴认识你。有什么想了解的吗？",
		NeedPassword: h.accessPassword != "",
	})
}

// VerifyHandler 验证访问密码
func (h *ChatHandler) VerifyHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, VerifyResponse{
			Success: false,
			Message: "密码不能为空",
		})
		return
	}

	if h.accessPassword == "" {
		c.JSON(http.StatusOK, VerifyResponse{
			Success: true,
			Message: "验证成功",
		})
		return
	}

	if req.Message == h.accessPassword {
		c.JSON(http.StatusOK, VerifyResponse{
			Success: true,
			Message: "验证成功",
		})
	} else {
		c.JSON(http.StatusOK, VerifyResponse{
			Success: false,
			Message: "密码错误，请重试",
		})
	}
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

	// 保存对话日志
	go logger.Log(req.Message, reply)

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

	var fullReply string

	c.Stream(func(w io.Writer) bool {
		err := h.chatService.ChatStream(c.Request.Context(), req.Message, func(chunk string) {
			fullReply += chunk
			encoded := base64.StdEncoding.EncodeToString([]byte(chunk))
			data := fmt.Sprintf("data: %s\n\n", encoded)
			c.Writer.Write([]byte(data))
			c.Writer.Flush()
		})
		if err != nil {
			c.Writer.Write([]byte(fmt.Sprintf("data: [ERROR] %s\n\n", err.Error())))
		}
		c.Writer.Write([]byte("data: [DONE]\n\n"))

		// 保存对话日志
		go logger.Log(req.Message, fullReply)

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

	h.personaData = persona
	h.chatService.UpdatePersona(persona)
	h.chatService.ClearHistory()

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
