package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ChatLog 单条对话记录
type ChatLog struct {
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	Assistant string `json:"assistant"`
}

// ChatLogger 对话日志记录器
type ChatLogger struct {
	filePath string
	mu       sync.Mutex
}

var (
	logger *ChatLogger
	once    sync.Once
)

// InitChatLogger 初始化日志记录器（单例）
func InitChatLogger(logPath string) {
	once.Do(func() {
		// 确保目录存在
		dir := filepath.Dir(logPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("创建日志目录失败: %v", err)
			return
		}
		logger = &ChatLogger{filePath: logPath}
	})
}

// Log 保存一条对话记录
func (l *ChatLogger) Log(user, assistant string) {
	if l == nil {
		return
	}

	entry := ChatLog{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		User:      user,
		Assistant: assistant,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("序列化日志失败: %v", err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("打开日志文件失败: %v", err)
		return
	}
	defer f.Close()

	f.Write(data)
	f.WriteString("\n")
}

// SaveConversation 保存完整对话（包含多轮）
func (l *ChatLogger) SaveConversation(messages []string, replies []string) {
	if l == nil || len(messages) == 0 {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("打开日志文件失败: %v", err)
		return
	}
	defer f.Close()

	header := fmt.Sprintf("\n========== 对话 %s ==========\n", time.Now().Format("2006-01-02 15:04:05"))
	f.WriteString(header)

	for i := 0; i < len(messages); i++ {
		f.WriteString(fmt.Sprintf("用户: %s\n", messages[i]))
		if i < len(replies) {
			f.WriteString(fmt.Sprintf("AI: %s\n", replies[i]))
		}
	}
	f.WriteString("================================\n\n")
}
