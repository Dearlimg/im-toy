// 未读消息状态：按会话 key 计数

import { createStore } from './index.js';

export const unreadStore = createStore({
  // key: 'friend:<friend_uid>' 或 'group:<group_id>'，value: 数量
  counts: {}
});

function key(chatType, id) {
  return chatType + ':' + id;
}

// 后端 private/unread-counts 返回 { friend_uid: count }
export function setPrivateUnread(map) {
  const cur = unreadStore.get().counts;
  const next = Object.assign({}, cur);
  Object.keys(next).forEach(k => {
    if (k.startsWith('friend:')) delete next[k];
  });
  if (map) {
    Object.keys(map).forEach(fid => {
      if (map[fid]) next['friend:' + fid] = map[fid];
    });
  }
  unreadStore.set({ counts: next });
}

export function increment(chatType, id) {
  const k = key(chatType, id);
  const cur = unreadStore.get().counts;
  unreadStore.set({ counts: Object.assign({}, cur, { [k]: (cur[k] || 0) + 1 }) });
}

export function clear(chatType, id) {
  const k = key(chatType, id);
  const cur = unreadStore.get().counts;
  if (!cur[k]) return;
  const next = Object.assign({}, cur);
  delete next[k];
  unreadStore.set({ counts: next });
}

export function getCount(chatType, id) {
  return unreadStore.get().counts[key(chatType, id)] || 0;
}

export function totalCount() {
  return Object.values(unreadStore.get().counts).reduce((a, b) => a + b, 0);
}
