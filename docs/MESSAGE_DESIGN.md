# 消息模型与扩展边界

## 单节点

HTTP 私聊和 WebSocket 私聊、群聊调用同一发送服务。以发送者和 `client_msg_id` 为幂等键，在 MySQL 事务中写入请求记录、递增会话序号并保存消息。重复请求若内容及目标一致，返回原来的消息 ID 和序号；内容或目标不一致则拒绝。提交成功后向发送者确认，并尽力推送在线接收者。

历史补拉使用 `(seq, id)` 作为不透明游标。迁移前的消息序号为 0，按数据库行 ID 排序；新消息在各自会话内从 1 开始。重连后按服务器提供的游标正向拉取并以消息 ID 去重。私聊已读只能由消息接收者上报；群聊已读需要验证有效成员和消息归属。

本机浏览器保存未确认发送的内容以便重连时使用原 ID 重试。私聊保存的是密文；群聊为明文。设备丢失和浏览器存储清除会丢失这些本地待发送记录。

## 多实例后续设计

当前连接表和推送都在进程内；共享 Kafka 消费组并不能将事件送到持有连接的实例。若需要多实例，建议在消息事务中加入 Outbox，独立发布器可靠投递事件；连接网关维护带租约的用户/设备到实例路由，按实例分发事件，处理重试、重复和实例退出。消息历史仍是事实来源，客户端继续用游标补拉。该方案目前只是设计方向，尚未实现或验证。

## 接口

连接 `GET /room?room=default&token=<JWT>` 后，私聊发送 `{"type":"private_message","to_uid":"用户 ID","client_msg_id":"客户端唯一 ID","content":"内容"}`，群聊使用 `type: "group_message"` 和 `group_id`。相同请求重试时必须保留原 `client_msg_id` 和原内容；服务端提交后回显同类型消息，携带 `msg_id`、`seq`、`client_msg_id`；拒绝或写入失败返回 `send_failed`、`client_msg_id` 和 `error`。HTTP `POST /api/private/send` 使用相同的 `to_uid`、`client_msg_id` 和 `content`。

认证用户可调用 `GET /api/sync/private?friend_uid=<ID>&tail=1&limit=50` 或 `GET /api/sync/group?group_id=<ID>&tail=1&limit=50` 获取最新一页。响应的 `messages` 为升序，`next_cursor` 用于增量同步，`older_cursor` 用于向前翻页，`has_older` 表示是否还有更早消息。增量请求传 `cursor=<next_cursor>`，按 `has_more` 继续；历史请求传 `before=<older_cursor>`，用返回的 `next_cursor` 继续向前。游标应原样保存并透传，不应由客户端自行计算。群历史每次查询均检查当前成员身份。

私聊已读发送 `{"type":"read_receipt","from_uid":"发送者 ID","msg_id":"消息 ID"}`；群聊发送 `{"type":"group_read_receipt","group_id":"群 ID","msg_ids":["消息 ID"]}`。私聊按照单条消息标记，不代表此前所有消息都已读；群聊每批最多 200 条。

## 升级旧部署

旧 Kafka 的消息和已读消费者以及 Redis 已读缓存刷盘器已停用。若旧 topic 或 Redis 存在尚未写入数据库的数据，切换前需要由旧版本在受控条件下完成消费并核对数据；不要直接假设所有事件均已持久化。升级时先备份数据库，再运行自动迁移并检查同步结果。
