# 项目记忆

## Badcase 沉淀

### 1. 前后端字段名不一致导致静默失败 (2026-08-02)
- **问题**：前端 chat.js 中 `isCurrent` 用 `chatId === msg.from_uid` 判断当前会话。但 WS 回显给发送方时 `from_uid` 是自己，`chatId` 是对方，导致回显不被处理、acked 永远不更新、消息不显示对号。
- **教训**：涉及双方 uid 的消息流转，必须同时检查 `from_uid` 和 `to_uid`。`peerUid = fromMe ? to_uid : from_uid` 模式。
- **同样问题**：readReceipt payload 字段是 `read_uid` 不是 `from_uid`，代码写死 `msg.from_uid` 导致已读回执永远匹配不到。

### 2. 部署时 config.js 被本地开发配置覆盖 (2026-08-02)
- **问题**：`scp -r public\*` 会把本地 `config.js`（`API_ORIGIN: http://127.0.0.1:8080`）覆盖服务器上的生产配置（`API_ORIGIN: ''`）。
- **教训**：部署前端文件时必须单独处理 config.js，或者在 scp 后立即修复。

### 3. Go 编译 -a 全量重编导致服务器 OOM (2026-08-02)
- **问题**：Dockerfile 中 `go build -a` 强制全量重编标准库，1.6G 内存服务器上 MySQL/Redis/Kafka/pgvector 全开时编译必 OOM。
- **修复**：去掉 `-a -installsuffix cgo`，加 BuildKit `--mount=type=cache`，构建前停基础服务 + 建 swap。

### 4. nginx volume 挂载 Permission denied (2026-08-02)
- **问题**：nginx:alpine 默认以 nginx 用户运行，对 volume 挂载的文件无读权限，返回 403。
- **修复**：docker-compose nginx command 加 `chmod -R 755 /usr/share/nginx/html`。

### 5. YAML key 与 Go struct tag 不匹配 (2026-08-02)
- **问题**：config.yaml 用 key `mysql` 但 Go struct tag 是 `database`，导致 `host` 解析为空，backend 报 `dial tcp :0`。
- **教训**：写配置文件时必须对照 Go struct 的 `mapstructure` tag。

## 部署信息

- 服务器：121.40.235.227，root，SSH key `~/.ssh/id_ed25519`
- SSH alias：`aliyun`
- 项目路径：/root/workspace/im-toy
- 基础服务 compose：/root/workspace/docker-compose.yaml
- 部署 skill：C:\Users\14925\.codebuddy\skills\im-toy-deploy\SKILL.md
