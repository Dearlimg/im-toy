# IM-Toy

一个简洁高效的即时通讯项目，用于学习与实践 IM 系统的核心架构设计。

## 技术栈

- **后端**：Go + Gin + GORM + Kafka + Redis + MySQL
- **前端**：原生 HTML/JS（轻量演示前端）
- **消息队列**：Kafka（异步消息处理）
- **缓存**：Redis（会话缓存、在线状态、消息幂等）

## 项目结构

```
├── cmd/                  # 入口目录
│   ├── main.go          # 后端服务入口
│   └── frontend/        # 前端服务入口
├── config/              # 配置文件
├── internal/            # 核心业务代码
│   ├── application/     # 应用层（Service）
│   ├── domain/          # 领域层（Entity、Repository 接口）
│   ├── infra/           # 基础设施层（Repository 实现、Kafka、Redis、数据库）
│   └── interface/      # 接口层（HTTP Handler、WebSocket）
├── public/              # 前端静态资源
├── scripts/             # Docker Compose 文件
└── docs/                # API 文档
```

## 快速启动

### 1. 配置环境

```bash
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml，填写数据库、Redis、Kafka 等配置
```

### 2. 启动依赖服务

```bash
docker-compose -f scripts/docker-compose.yml up -d
```

### 3. 启动服务

```bash
# 后端服务
go run cmd/main.go

# 前端服务（新终端）
go run cmd/frontend/main.go
```

访问 `http://localhost:8080` 即可使用。

## 核心特性

- **Kafka 异步消息**：消息发送与存储解耦，支持消息重放
- **WebSocket 实时通讯**：支持私聊和群聊
- **消息幂等**：Redis 去重 + MySQL 唯一索引双重保障
- **离线消息**：用户上线后主动拉取未读消息
- **消息已读**：支持单聊和群聊的已读回执
- **在线状态**：基于 Redis 的用户在线状态管理