// 好友状态：列表、请求、当前会话目标

import { createStore } from './index.js';

export const friendsStore = createStore({
  friends: [],          // [{ uid, username, nickname }]
  requests: [],         // [{ id, from_uid, from_username, from_nickname, status, created_at }]
  current: null         // 当前私聊对象 { uid, username, nickname }
});

export function setFriends(list) {
  friendsStore.set({ friends: list || [] });
}

export function setRequests(list) {
  friendsStore.set({ requests: list || [] });
}

export function setCurrentFriend(friend) {
  friendsStore.set({ current: friend });
}

// 按 uid 查找好友（用于渲染）
export function findFriend(uid) {
  return friendsStore.get().friends.find(f => String(f.uid) === String(uid)) || null;
}

// 更新好友在线状态（presence 推送实时更新）
export function updateFriendOnline(uid, online) {
  const friends = friendsStore.get().friends.map(f =>
    String(f.friend_uid) === String(uid)
      ? Object.assign({}, f, { is_online: !!online })
      : f
  );
  friendsStore.set({ friends });
}
