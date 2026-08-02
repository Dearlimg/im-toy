// 会话状态：当前会话类型 + 消息历史

import { createStore } from './index.js';

export const conversationsStore = createStore({
  chatType: null,      // 'friend' | 'group' | null
  chatId: null,        // 好友 friend_uid 或群 group_id
  messages: []         // 当前会话消息
});

export function openConversation(chatType, chatId) {
  conversationsStore.set({ chatType, chatId, messages: [] });
}

export function setMessages(list) {
  conversationsStore.set({ messages: list || [] });
}

// 追加一条消息（去重 + 时间排序）
// 若新消息带 msg_id，会替换同 from_uid+content 的本地乐观消息（无 msg_id），避免重复
export function pushMessage(msg) {
  const cur = conversationsStore.get();

  if (msg.msg_id) {
    const exists = cur.messages.some(m =>
      m.msg_id && String(m.msg_id) === String(msg.msg_id));
    if (exists) return cur.messages;

    // 去掉同发送者同内容的乐观消息（send_time 接近）
    const now = new Date(msg.send_time).getTime();
    const filtered = cur.messages.filter(m => {
      if (m.msg_id) return true;
      const sameSender = String(m.from_uid) === String(msg.from_uid);
      const sameContent = m.content === msg.content;
      const nearTime = Math.abs(new Date(m.send_time).getTime() - now) < 60000;
      return !(sameSender && sameContent && nearTime);
    });

    const next = filtered.concat(msg).sort((a, b) =>
      new Date(a.send_time) - new Date(b.send_time));
    conversationsStore.set({ messages: next });
    return next;
  }

  const next = cur.messages.concat(msg).sort((a, b) =>
    new Date(a.send_time) - new Date(b.send_time));
  conversationsStore.set({ messages: next });
  return next;
}

// 更新消息已读状态
export function updateMessageRead(msgId, isRead, readCount, totalCount) {
  const cur = conversationsStore.get();
  const messages = cur.messages.map(m => {
    if (String(m.msg_id) === String(msgId)) {
      return Object.assign({}, m, {
        is_read: isRead != null ? isRead : m.is_read,
        read_count: readCount != null ? readCount : m.read_count,
        total_count: totalCount != null ? totalCount : m.total_count
      });
    }
    return m;
  });
  conversationsStore.set({ messages });
}
