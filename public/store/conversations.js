// 会话状态：当前会话类型 + 消息历史

import { createStore } from './index.js';

export const conversationsStore = createStore({
  chatType: null,      // 'friend' | 'group' | null
  chatId: null,        // 好友 friend_uid 或群 group_id
  messages: []         // 当前会话消息
});

export function openConversation(chatType, chatId) {
  conversationsStore.set({ chatType, chatId, messages: [], hasOlder: false });
}

function compareMessages(a, b) {
  const aSeq = a.msg_id && a.seq != null ? Number(a.seq) : Number.POSITIVE_INFINITY;
  const bSeq = b.msg_id && b.seq != null ? Number(b.seq) : Number.POSITIVE_INFINITY;
  if (aSeq !== bSeq) return aSeq - bSeq;
  if (aSeq === 0 && a.ID != null && b.ID != null) return Number(a.ID) - Number(b.ID);
  return Number(a.send_time) - Number(b.send_time);
}

export function mergeMessages(incoming) {
  const cur = conversationsStore.get();
  const messages = [...cur.messages];
  for (const msg of incoming) {
    if (msg.msg_id && messages.some(m => m.msg_id && String(m.msg_id) === String(msg.msg_id))) continue;
    const optimistic = msg.client_msg_id && messages.findIndex(m => !m.msg_id && m.client_msg_id === msg.client_msg_id);
    if (typeof optimistic === 'number' && optimistic >= 0) messages.splice(optimistic, 1);
    messages.push(msg);
  }
  messages.sort(compareMessages);
  conversationsStore.set({ messages });
}

export function setMessages(list) {
  conversationsStore.set({ messages: list || [] });
}

// 追加一条消息（去重 + 时间排序）
// 服务端回显用 client_msg_id 精确替换乐观消息，避免相同内容相互覆盖。
export function pushMessage(msg) {
  const cur = conversationsStore.get();

  if (msg.msg_id) {
    const exists = cur.messages.some(m =>
      m.msg_id && String(m.msg_id) === String(msg.msg_id));
    if (exists) {
      const next = cur.messages.filter(m => !(!m.msg_id && msg.client_msg_id && m.client_msg_id === msg.client_msg_id));
      if (next.length !== cur.messages.length) conversationsStore.set({ messages: next });
      return next;
    }

    const filtered = cur.messages.filter(m => {
      return !(!m.msg_id && msg.client_msg_id && m.client_msg_id === msg.client_msg_id);
    });

    const next = filtered.concat(msg).sort(compareMessages);
    conversationsStore.set({ messages: next });
    return next;
  }

  const next = cur.messages.concat(msg).sort(compareMessages);
  conversationsStore.set({ messages: next });
  return next;
}

export function markMessageFailed(clientMsgID) {
  const cur = conversationsStore.get();
  conversationsStore.set({ messages: cur.messages.map(m =>
    m.client_msg_id === clientMsgID && !m.msg_id ? { ...m, failed: true } : m) });
}

export function markMessagePending(clientMsgID) {
  const cur = conversationsStore.get();
  conversationsStore.set({ messages: cur.messages.map(m =>
    m.client_msg_id === clientMsgID && !m.msg_id ? { ...m, failed: false } : m) });
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
