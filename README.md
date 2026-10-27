# IM-Toy

使用 Go、Gin、MySQL、Redis、WebSocket 和原生 JavaScript 实现的单节点 IM 学习项目。支持账号、好友、私聊、群聊、历史分页和已读回执。当前实现以 MySQL 消息记录为事实来源，实时推送是尽力而为的加速路径。

## 本地开发

需要 Go 1.25+ 和 Docker Compose。下面的配置仅供本机开发，包含公开的测试口令。

```bash
docker compose -f docker-compose.dev.yml up -d --build
```

访问 `http://127.0.0.1:8080`。此方式由后端同时提供静态页面和 API；MySQL、Redis 分别映射到本机 3307、6380，当前消息路径不使用 Kafka。停止服务可执行 `docker compose -f docker-compose.dev.yml down`；数据库卷默认保留。

也可以只启动依赖，在主机运行后端：

```bash
docker compose -f docker-compose.dev.yml up -d mysql redis
cp config/config.example.yaml config/config.yaml
go run ./cmd
```

Windows PowerShell 可用 `Copy-Item config/config.example.yaml config/config.yaml`。配置示例与开发 Compose 的本机端口一致。现有根目录 `docker-compose.yml` 面向已有外部网络和部署环境，不是上述开发配置。

## 消息协议和迁移

- 私聊和群聊发送共用消息服务；客户端发送唯一的 `client_msg_id`。消息、幂等记录及会话序号在同一事务提交后确认。
- `/api/sync/private` 和 `/api/sync/group` 使用服务端游标补拉，支持正向分页和 `before` 历史分页。WebSocket 在线推送不保证到达，客户端重连后补拉。
- 当前只支持单进程连接路由，不应通过增加后端副本来扩容。跨实例设计见 [消息设计](docs/MESSAGE_DESIGN.md)。
- 启动时自动迁移数据库结构，并将旧消息的空序号填为 0；旧消息用数据库行 ID 排序。升级已有数据前应备份数据库。
- 新版本不再启动旧 Kafka 生产者、消费者和 Redis 已读缓存刷盘器。切换前需核对并处理旧 topic 或 Redis 中尚未落库的积压；新版本不会自动导入这些事件。

## 加密边界

私聊要求浏览器 Web Crypto 和对方公钥；加密不可用时阻止发送，不自动降级为明文。首次使用或对方公钥变化时，应通过独立可信渠道核对界面显示的指纹后再确认。群聊未端到端加密。

当前私钥保存在本机浏览器存储中，清除数据或更换设备会丢失密钥；本项目没有密钥备份、多设备同步、前向保密或完整的密钥恢复机制。新设备替换已注册公钥需要用户明确确认，但旧历史密文可能无法解密。浏览器中的恶意脚本和被控制的页面仍可能读取密钥。非本机访问应部署 HTTPS。

## 验证

```bash
go test ./...
node --test tests/*.test.mjs
```

Go 测试包含 SQL 模拟测试，不替代真实 MySQL、浏览器和网络故障场景下的端到端验证。
