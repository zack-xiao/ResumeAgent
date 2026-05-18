# ResumeAgent - 个人 AI 展示页面

一个基于 AI 的个人展示页面，让面试官通过对话了解你。

## 技术栈

- **后端**: Go + Gin + eino
- **前端**: HTML + CSS + JavaScript
- **AI**: DeepSeek API

## 快速开始

### 1. 安装依赖

```bash
cd backend
go mod tidy
```

### 2. 配置 API Key

编辑 `.env` 文件：

```env
DEEPSEEK_API_KEY=sk-your-api-key
DEEPSEEK_MODEL=deepseek-chat
PORT=8080
```

### 3. 编辑个人信息

修改 `data/persona.md` 文件，填写你的真实信息。

### 4. 启动服务

```bash
cd backend
go run main.go
```

### 5. 访问页面

打开浏览器访问: http://localhost:8080

## 项目结构

```
ResumeAgent/
├── backend/            # 后端代码
│   ├── main.go
│   ├── config.go
│   ├── handler/        # HTTP 处理器
│   ├── service/        # AI 服务
│   └── persona/        # 人物设定加载
├── frontend/           # 前端代码
│   ├── index.html
│   ├── css/
│   └── js/
├── data/               # 数据文件
│   ├── persona.md      # 你的个人信息
│   └── prompt.md       # AI 提示词
└── .env                # 环境配置
```

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/init` | GET | 获取初始化信息 |
| `/api/chat` | POST | 普通对话 |
| `/api/chat/stream` | POST | 流式对话 |
| `/api/reload` | POST | 重新加载人物设定 |

## 注意事项

1. 请确保 DeepSeek API Key 有效
2. 定期更新 `persona.md` 中的个人信息
3. 生产环境请配置 HTTPS
