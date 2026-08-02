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

// 更新消息内容（E2EE 解密后替换密文）
// msg：原始消息（含密文 content）；plain：解密后的明文
export function updateMessageContent(msg, plain) {
  const cur = conversationsStore.get();
  const messages = cur.messages.map(m => {
    const sameId = msg.msg_id != null && String(m.msg_id) === String(msg.msg_id);
    const sameLocal = msg.msg_id == null && !m.msg_id &&
      String(m.from_uid) === String(msg.from_uid) && m.content === msg.content;
    if (sameId || sameLocal) {
      return Object.assign({}, m, { content: plain });
    }
    return m;
  });
  conversationsStore.set({ messages });
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
