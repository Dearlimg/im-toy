-- =====================================================
-- IM 聊天系统索引优化 SQL
-- 执行时间: 2026-04-05
-- 说明: 优化群聊消息查询、私聊会话查询性能
-- =====================================================

-- ==================== 1. 群消息表索引优化 ====================
-- 原因: 拉取群历史消息需要按 (group_id, send_time DESC) 排序
--       加上索引后 200ms+ 慢SQL 直接变成 5ms 内

ALTER TABLE `group_messages`
ADD INDEX `idx_group_id_send_time` (`group_id`, `send_time` DESC);


-- ==================== 2. 私聊消息表索引优化 ====================
-- 原因: 拉取两人会话记录需要按 (from_uid, to_uid, send_time DESC) 排序

ALTER TABLE `private_messages`
ADD INDEX `idx_private_conversation` (`from_uid`, `to_uid`, `send_time` DESC);


-- ==================== 3. 好友关系表唯一索引 ====================
-- 原因: 防止同一用户重复添加同一好友
--       业务逻辑上 uid + friend_uid 应该是唯一组合

ALTER TABLE `friends`
ADD UNIQUE INDEX `idx_uid_friend_uid` (`uid`, `friend_uid`);


-- =====================================================
-- 验证索引是否创建成功
-- =====================================================
SHOW INDEX FROM `group_messages`;
SHOW INDEX FROM `private_messages`;
SHOW INDEX FROM `friends`;
