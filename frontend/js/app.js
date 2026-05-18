class ChatApp {
    constructor() {
        this.chatMessages = document.getElementById('chat-messages');
        this.messageInput = document.getElementById('message-input');
        this.sendBtn = document.getElementById('send-btn');
        this.aiName = document.getElementById('ai-name');
        this.welcomeCard = document.getElementById('welcome-card');

        this.init();
    }

    init() {
        this.bindEvents();
        this.loadInit();
        this.adjustTextareaHeight();
        this.initQuickButtons();
    }

    bindEvents() {
        this.sendBtn.addEventListener('click', () => this.sendMessage());

        this.messageInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        this.messageInput.addEventListener('input', () => {
            this.adjustTextareaHeight();
        });
    }

    adjustTextareaHeight() {
        this.messageInput.style.height = 'auto';
        this.messageInput.style.height = Math.min(this.messageInput.scrollHeight, 150) + 'px';
    }

    initQuickButtons() {
        document.querySelectorAll('.quick-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const question = btn.dataset.question;
                this.messageInput.value = question;
                this.sendMessage();
            });
        });
    }

    async loadInit() {
        try {
            const response = await fetch('/api/init');
            const data = await response.json();

            if (data.name) {
                this.aiName.textContent = data.name;
            }
            if (data.welcome) {
                this.updateWelcomeMessage(data.welcome);
            }
        } catch (error) {
            console.error('初始化失败:', error);
        }
    }

    updateWelcomeMessage(welcome) {
        const h1 = this.welcomeCard?.querySelector('h1');
        const p = this.welcomeCard?.querySelector('p');
        if (h1) {
            h1.textContent = welcome.split('！')[0] + '！';
        }
        if (p) {
            p.textContent = '你可以直接向我提问，我会以第一人称帮你介绍正烁的背景和经历';
        }
    }

    async sendMessage() {
        const message = this.messageInput.value.trim();
        if (!message) return;

        // 清空输入框
        this.messageInput.value = '';
        this.adjustTextareaHeight();

        // 隐藏欢迎卡片
        if (this.welcomeCard) {
            this.welcomeCard.style.display = 'none';
        }

        // 添加用户消息
        this.addMessage('user', message);

        // 添加 AI 消息占位
        const aiMessageEl = this.addMessage('ai', '', true);

        // 禁用发送按钮
        this.sendBtn.disabled = true;

        try {
            await this.streamChat(message, aiMessageEl);
        } catch (error) {
            aiMessageEl.querySelector('.message-content').innerHTML = 
                `<span style="color: #ef4444;">发生错误: ${error.message}</span>`;
        } finally {
            this.sendBtn.disabled = false;
        }
    }

    async streamChat(message, messageEl) {
        const contentEl = messageEl.querySelector('.message-content');
        let fullContent = '';

        try {
            const response = await fetch('/api/chat/stream', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ message }),
            });

            const reader = response.body.getReader();
            const decoder = new TextDecoder();

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');

                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const data = line.slice(6);
                        if (data === '[DONE]') continue;
                        if (data.startsWith('[ERROR]')) {
                            throw new Error(data.slice(7));
                        }
                        fullContent += data;
                        contentEl.textContent = fullContent;
                        this.scrollToBottom();
                    }
                }
            }
        } catch (error) {
            // 如果流式失败，尝试普通请求
            try {
                const response = await fetch('/api/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ message }),
                });

                const data = await response.json();
                if (data.error) {
                    throw new Error(data.error);
                }
                contentEl.textContent = data.reply;
            } catch (retryError) {
                contentEl.innerHTML = `<span style="color: #ef4444;">发生错误: ${retryError.message}</span>`;
            }
        }

        this.scrollToBottom();
    }

    addMessage(role, content, placeholder = false) {
        const messageEl = document.createElement('div');
        messageEl.className = `message ${role}`;

        const avatar = role === 'ai' ? '🤖' : '👤';

        if (placeholder) {
            messageEl.innerHTML = `
                <div class="message-avatar">${avatar}</div>
                <div class="message-content">
                    <div class="typing-indicator">
                        <span></span>
                        <span></span>
                        <span></span>
                    </div>
                </div>
            `;
        } else {
            messageEl.innerHTML = `
                <div class="message-avatar">${avatar}</div>
                <div class="message-content">${this.escapeHtml(content)}</div>
            `;
        }

        this.chatMessages.appendChild(messageEl);
        this.scrollToBottom();

        return messageEl;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    scrollToBottom() {
        this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    window.chatApp = new ChatApp();
});
