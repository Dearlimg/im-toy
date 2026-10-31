// REST API 业务接口：auth / friend / group / message
// 返回后端 data 字段（已按 {code,message,data} 约定解包），失败抛 ApiError

import { http } from './http.js';

// ---- 认证 ----
export const register = (username, nickname, password) =>
  http.post('/api/user/register', { username, nickname, password });

export const login = (username, password) =>
  http.post('/api/user/login', { username, password });

// ---- E2EE ----
export const getPubKey = () => http.get('/api/e2ee/pubkey');
export const registerPubKey = (pubkey, replace = false) => http.post('/api/e2ee/pubkey', { pubkey, replace });

// ---- 好友 ----
export const friendList = () => http.get('/api/friend/list');
export const friendDelete = (friendUid) => http.post('/api/friend/delete', { friend_uid: friendUid });
export const friendRequests = () => http.get('/api/friend/requests');
export const friendAccept = (friendUid) => http.post('/api/friend/accept', { friend_uid: friendUid });
export const friendSearch = (keyword) => http.post('/api/friend/search', { keyword });
export const friendAdd = (friendUid) => http.post('/api/friend/add', { friend_uid: friendUid });

// ---- 群组 ----
export const groupList = () => http.get('/api/group/list');
export const groupMembers = (groupId) => http.get('/api/group/members?group_id=' + encodeURIComponent(groupId));
export const groupCreate = (name, memberUids) => http.post('/api/group/create', { name, member_uids: memberUids });

// ---- 消息 ----
export const privateConversation = (friendUid) =>
  http.get('/api/private/conversation?friend_uid=' + encodeURIComponent(friendUid));

export const groupMessages = (groupId, limit) =>
  http.get('/api/group/messages?group_id=' + encodeURIComponent(groupId) + '&limit=' + (limit || 50));

export const syncMessages = (type, id, options = {}) => {
  const path = type === 'friend' ? '/api/sync/private?friend_uid=' : '/api/sync/group?group_id=';
  const query = new URLSearchParams({ limit: String(options.limit || 50) });
  if (options.tail) query.set('tail', '1');
  if (options.cursor) query.set('cursor', options.cursor);
  if (options.before) query.set('before', options.before);
  return http.get(path + encodeURIComponent(id) + '&' + query.toString());
};

export const readStatus = (roomId) =>
  http.get('/api/message/read-status?room_id=' + encodeURIComponent(roomId));

export const privateUnreadCounts = () => http.get('/api/private/unread-counts');

export const privateUndelivered = () => http.get('/api/private/undelivered?limit=100');
export const privateMessageStatuses = ids =>
  http.get('/api/private/status?msg_ids=' + encodeURIComponent(ids.join(',')));
