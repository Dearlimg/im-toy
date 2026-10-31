// 页面层：主聊天页编排入口
// 职责：初始化组件、串联 store/services、注册 WS 事件、处理会话切换与发送

import { authStore, loadSession, clearSession } from '../store/auth.js';
import { setCurrentFriend, updateFriendOnline, findFriend } from '../store/friends.js';
import { setCurrentGroup, groupsStore } from '../store/groups.js';
import { conversationsStore, openConversation, pushMessage, mergeMessages, updateMessageRead, updateMessageDelivered, applyDeliveryStatuses, updateMessageContent, markMessageFailed, markMessagePending } from '../store/conversations.js';
import { ensureE2EE, decryptContent, isEncrypted } from '../services/e2ee.js';
import { unreadStore, setPrivateUnread, increment, clear, totalCount } from '../store/unread.js';

import { initAuthForm } from '../components/business/AuthForm.js';
import { setHandlers as setSidebarHandlers, loadAll, switchSide, render } from '../components/business/Sidebar.js';
import { updateHeader, renderAll, scrollToBottom, setUidProvider, setPeerOnlineProvider, clearMessages, setLoadOlderHandler } from '../components/business/ChatWindow.js';
import { initInput, setSendHandler } from '../components/business/ChatInput.js';
import { Toast } from '../components/common/Toast.js';

import { syncMessages, privateUnreadCounts, privateUndelivered, privateMessageStatuses, friendSearch, friendAdd, groupMembers } from '../services/api.js';
import { openModal } from '../components/common/Modal.js';
import { connect, disconnect, onMessage, send, isOpen } from '../services/ws.js';
import { createDeliveryQueue } from '../services/delivery.js';

const app = document.getElementById('app');
const authPage = document.getElementById('authPage');
const meNick = document.getElementById('meNick');
const meAvatar = document.getElementById('meAvatar');
const searchInput = document.getElementById('searchInput');

let myUid = null;
let pendingMessages = new Map();
let failedMessages = new Set();
let syncState = null;
let deliveryQueue = null;
let inboxLoading = false;
let inboxTimer = null;
const readReceiptsSent = new Map();
function savePending() {
  try { localStorage.setItem('im_pending:' + myUid, JSON.stringify({ pending: [...pendingMessages], failed: [...failedMessages] })); } catch (e) { console.warn('pending messages not saved', e); }
}
function loadPending() {
  try {
    const data = JSON.parse(localStorage.getItem('im_pending:' + myUid) || '{}');
    pendingMessages = new Map(Array.isArray(data) ? data : (data.pending || []));
    failedMessages = new Set(data.failed || []);
  } catch (e) { pendingMessages = new Map(); failedMessages = new Set(); }
}
window.addEventListener('im-retry-message', e => {
  const id = e.detail?.clientMsgID;
  const payload = pendingMessages.get(id);
  if (!payload) return;
  if (!send(payload)) { Toast('未连接到服务器，无法重试'); return; }
  failedMessages.delete(id);
  markMessagePending(id);
  savePending();
});
function clientMessageID() {
  if (crypto.randomUUID) return crypto.randomUUID();
  return Array.from(crypto.getRandomValues(new Uint8Array(16)), b => b.toString(16).padStart(2, '0')).join('');
}
setUidProvider(() => myUid);

// 当前会话对方是否在线（由 selectFriend/selectGroup 时设置）
let peerOnline = false;
setPeerOnlineProvider(() => peerOnline);

// ============ 启动 ============
document.addEventListener('DOMContentLoaded', () => {
  loadSession();
  initInput();
  setSendHandler(handleSend);
  setLoadOlderHandler(loadOlder);
  setSidebarHandlers({ selectFriend, selectGroup });
  onMessage(handleWsEvent);

  if (authStore.get().token) {
    enterApp();
  } else {
    initAuthForm({ onSuccess: enterApp });
  }
});

// ============ 进入主界面 ============
async function enterApp() {
  authPage.classList.add('hidden');
  app.classList.remove('hidden');

  const u = authStore.get().user || {};
  myUid = u.uid;
  loadPending();
  deliveryQueue = createDeliveryQueue(myUid, send);
  meNick.textContent = u.nickname || u.username || '我';
  meAvatar.textContent = (u.nickname || u.username || '我').charAt(0).toUpperCase();
  document.getElementById('railProfile').textContent = meAvatar.textContent;
  document.getElementById('railProfile').title = u.nickname || u.username || '我';

  connect();
  await ensureE2EE(myUid, { readOnly: true });
  await loadAll();
  loadUnread();
  render();

  // 实时消息订阅 → 重渲染列表与标题未读
  unreadStore.subscribe(() => {
    const total = totalCount();
    document.title = total > 0 ? '(' + total + ') 微聊' : '微聊';
    render();
  });
  conversationsStore.subscribe(() => renderAll());

  // 侧栏 tab 切换
  document.querySelectorAll('.side-tabs button, .rail-tab').forEach(b => {
    b.addEventListener('click', () => switchSide(b.dataset.tab));
  });
}

// ============ 退出 ============
export function logout() {
  disconnect();
  clearSession();
  location.reload();
}

// ============ 会话切换 ============
async function selectFriend(f) {
  setCurrentFriend(f);
  openConversation('friend', f.friend_uid);
  clear('friend', f.friend_uid);
  peerOnline = !!f.is_online;
  updateHeader('friend', f);
  await loadHistory('friend', f.friend_uid);
  if (conversationsStore.get().chatType !== 'friend' || String(conversationsStore.get().chatId) !== String(f.friend_uid)) return;
  sendReadReceipts();
  render();
  renderAll();
}

async function selectGroup(g) {
  setCurrentGroup(g);
  groupsStore.set({ members: [] });
  openConversation('group', g.group_id);
  clear('group', g.group_id);
  peerOnline = false;
  updateHeader('group', g);
  try {
    const members = await groupMembers(g.group_id);
    if (String(groupsStore.get().current?.group_id) === String(g.group_id)) groupsStore.set({ members: members || [] });
  } catch (e) { console.warn('could not load group members', e); }
  await loadHistory('group', g.group_id);
  if (conversationsStore.get().chatType !== 'group' || String(conversationsStore.get().chatId) !== String(g.group_id)) return;
  sendReadReceipts();
  render();
  renderAll();
}

async function loadHistory(chatType, id) {
  const state = { chatType, id: String(id), cursor: '', older: '', loading: false };
  syncState = state;
  try {
    const page = await syncMessages(chatType, id, { tail: true });
    let list = (page.messages || []).map(m => ({ ...m, acked: true }));
    if (chatType === 'friend') list = await Promise.all(list.map(m => resolveCipher(m, id)));
    if (syncState !== state) return;
    if (chatType === 'friend') list.forEach(m => deliveryQueue?.enqueue(m));
    state.cursor = page.next_cursor || '';
    state.older = page.older_cursor || '';
    // Preserve real-time arrivals and pending messages while the history request was in flight.
    const existing = conversationsStore.get().messages;
    conversationsStore.set({ messages: list, hasOlder: !!page.has_older });
    mergeMessages(existing);
    for (const [clientID, payload] of pendingMessages) {
      const target = chatType === 'friend' ? payload.to_uid : payload.group_id;
      if (String(target) !== String(id) || payload.type !== (chatType === 'friend' ? 'private_message' : 'group_message')) continue;
      if (conversationsStore.get().messages.some(m => m.client_msg_id === clientID)) continue;
      pushMessage({ from_uid: myUid, content: payload.content, client_msg_id: clientID,
        send_time: payload.client_time || Date.now(), acked: false, failed: failedMessages.has(clientID) });
    }
    await syncCurrent();
    renderAll();
  } catch (e) {
    Toast('加载历史消息失败: ' + e.message);
  }
}

async function syncCurrent() {
  const state = syncState;
  if (!state || state.loading) return;
  state.loading = true;
  try {
    let more;
    do {
      const page = await syncMessages(state.chatType, state.id, { cursor: state.cursor });
      if (syncState !== state) return;
      const rows = (page.messages || []).map(raw => ({ ...raw, acked: true }));
      const messages = state.chatType === 'friend'
        ? await Promise.all(rows.map(msg => resolveCipher(msg, state.id))) : rows;
      if (syncState !== state) return;
      if (state.chatType === 'friend') messages.forEach(m => deliveryQueue?.enqueue(m));
      mergeMessages(messages);
      state.cursor = page.next_cursor || state.cursor;
      more = !!page.has_more;
    } while (more);
  } catch (e) {
    Toast('同步消息失败：' + e.message);
  } finally {
    state.loading = false;
  }
}

async function loadOlder() {
  const state = syncState;
  if (!state || !state.older) return;
  try {
    const page = await syncMessages(state.chatType, state.id, { before: state.older });
    const rows = (page.messages || []).map(m => ({ ...m, acked: true }));
    const list = state.chatType === 'friend'
      ? await Promise.all(rows.map(m => resolveCipher(m, state.id)))
      : rows;
    if (syncState !== state) return;
    if (state.chatType === 'friend') list.forEach(m => deliveryQueue?.enqueue(m));
    state.older = page.next_cursor || state.older;
    mergeMessages(list);
    conversationsStore.set({ hasOlder: !!page.has_more });
    sendReadReceipts();
  } catch (e) {
    Toast('加载更早消息失败：' + e.message);
  }
}

// 对当前会话中"对方发来且未读"的消息补发已读回执，
// 让对方实时看到单对号 → 双对号（覆盖离线期间收到的消息）
function sendReadReceipts() {
  if (document.visibilityState === 'hidden' || !document.hasFocus()) return;
  const st = conversationsStore.get();
  const viewport = document.getElementById('chatMessages');
  if (!viewport) return;
  const bounds = viewport.getBoundingClientRect();
  const visibleIDs = new Set();
  viewport.querySelectorAll('[data-msg-id]').forEach(row => {
    const rect = row.getBoundingClientRect();
    if (rect.bottom > bounds.top && rect.top < bounds.bottom && rect.height > 0) visibleIDs.add(row.dataset.msgId);
  });
  const now = Date.now();
  const eligible = m => m.msg_id && visibleIDs.has(String(m.msg_id)) && String(m.from_uid) !== String(myUid) &&
    now - (readReceiptsSent.get(String(m.msg_id)) || 0) >= 30000;
  if (st.chatType === 'group') {
    const ids = (st.messages || []).filter(eligible).map(m => m.msg_id);
    for (let i = 0; i < ids.length; i += 200) {
      const batch = ids.slice(i, i + 200);
      if (send({ type: 'group_read_receipt', group_id: st.chatId, msg_ids: batch })) batch.forEach(id => readReceiptsSent.set(String(id), now));
    }
    return;
  }
  (st.messages || []).forEach(m => {
    if (eligible(m) && !m.is_read && !isEncrypted(m.content)) {
      if (send({ type: 'read_receipt', msg_id: m.msg_id, from_uid: m.from_uid })) readReceiptsSent.set(String(m.msg_id), now);
    }
  });
}

async function drainUndelivered() {
  if (!deliveryQueue || inboxLoading || !isOpen()) return;
  inboxLoading = true;
  let hasMore = false;
  try {
    const page = await privateUndelivered();
    hasMore = !!page.has_more;
    for (const message of page.messages || []) deliveryQueue.enqueue(message);
  } catch (e) {
    console.warn('Could not sync undelivered messages:', e);
  } finally {
    inboxLoading = false;
    if (hasMore && !deliveryQueue.size()) scheduleInboxSync();
  }
}

function scheduleInboxSync() {
  clearTimeout(inboxTimer);
  inboxTimer = setTimeout(drainUndelivered, 150);
}

async function refreshVisibleStatuses() {
  const state = syncState;
  if (!state || state.chatType !== 'friend') return;
  const ids = conversationsStore.get().messages.filter(m => m.msg_id && String(m.from_uid) === String(myUid)).map(m => m.msg_id);
  try {
    for (let i = 0; i < ids.length; i += 100) {
      const result = await privateMessageStatuses(ids.slice(i, i + 100));
      if (syncState !== state) return;
      applyDeliveryStatuses(result.messages || []);
    }
  } catch (e) {
    console.warn('Could not refresh delivery status:', e);
  }
}

window.addEventListener('focus', () => { sendReadReceipts(); refreshVisibleStatuses(); });
let readScrollTimer;
document.getElementById('chatMessages').addEventListener('scroll', () => {
  clearTimeout(readScrollTimer);
  readScrollTimer = setTimeout(sendReadReceipts, 100);
}, { passive: true });
document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'visible') sendReadReceipts();
});
// Reconcile dropped acknowledgements and status notifications without waiting
// for the socket to disconnect. The history remains the source of truth.
setInterval(() => {
  if (!myUid || !isOpen()) return;
  deliveryQueue?.flush();
  drainUndelivered();
  if (document.visibilityState !== 'hidden') refreshVisibleStatuses();
}, 15000);

// ============ 发送消息 ============
async function handleSend(content) {
  const st = conversationsStore.get();
  if (!st.chatType) {
    Toast('请先选择一个会话');
    return false;
  }
  const toSend = content;
  const current = conversationsStore.get();
  if (current.chatType !== st.chatType || String(current.chatId) !== String(st.chatId)) {
    Toast('会话已切换，消息未发送');
    return false;
  }
  const clientMsgID = clientMessageID();
  const payload = st.chatType === 'friend'
    ? { type: 'private_message', to_uid: st.chatId, content: toSend, client_msg_id: clientMsgID, client_time: Date.now() }
    : { type: 'group_message', group_id: st.chatId, content, client_msg_id: clientMsgID, client_time: Date.now() };
  const ok = send(payload);
  if (!ok) {
    Toast('发送失败：未连接到服务器');
    return false;
  }
  pendingMessages.set(clientMsgID, payload);
  savePending();
  const u = authStore.get().user || {};
  // 乐观消息用密文入列，服务端回显（同为密文）可匹配替换，回显后解密为明文
  pushMessage({
    from_uid: u.uid,
    from_nickname: u.nickname || u.username,
    content: toSend,
    client_msg_id: clientMsgID,
    send_time: Date.now(),
    acked: false,
    is_read: false
  });
  renderAll();
  scrollToBottom();
  return true;
}

// E2EE：尝试用好友公钥解密 enc. 密文；非密文或解密失败原样返回
// 说明：ECDH 共享密钥对称，发送方也能用自己的私钥 + 对方公钥解开自己发出的密文
async function resolveCipher(msg, friendUid) {
  if (!isEncrypted(msg.content)) return msg;
  const peer = findFriend(friendUid);
  if (!peer || !peer.pubkey) return msg;
  const plain = await decryptContent(peer.pubkey, msg.content);
  if (plain != null) return Object.assign({}, msg, { content: plain });
  return msg;
}

// ============ WS 事件分发 ============
function handleWsEvent(type, msg) {
  const st = conversationsStore.get();
  switch (type) {
    case 'connect':
      readReceiptsSent.clear();
      for (const [id, payload] of pendingMessages) if (!failedMessages.has(id)) send(payload);
      deliveryQueue?.flush();
      drainUndelivered();
      syncCurrent().then(() => { sendReadReceipts(); return refreshVisibleStatuses(); });
      break;
    case 'deliveryReceipt':
      updateMessageDelivered(msg.msg_id);
      break;
    case 'deliveryAck':
      deliveryQueue?.acknowledge(msg.msg_id);
      scheduleInboxSync();
      break;
    case 'sendFailed':
      failedMessages.add(msg.client_msg_id);
      savePending();
      markMessageFailed(msg.client_msg_id);
      Toast('发送失败：' + (msg.error || '请重试'));
      break;
    case 'private': {
      const fromMe = String(msg.from_uid) === String(myUid);
      const toMe = String(msg.to_uid) === String(myUid);
      if (!fromMe && !toMe) return;
      if (toMe) deliveryQueue?.enqueue(msg);
      // 服务端回显给发送方时标记 acked
      if (fromMe) {
        msg.acked = true;
        pendingMessages.delete(msg.client_msg_id);
        failedMessages.delete(msg.client_msg_id);
        savePending();
      }
      // 当前会话的 chatId 是对方 uid，回显时 from_uid=自己 to_uid=对方
      const peerUid = fromMe ? msg.to_uid : msg.from_uid;
      const isCurrent = st.chatType === 'friend' && String(st.chatId) === String(peerUid);
      if (isCurrent) {
        if (fromMe) {
          // 自己的消息：密文回显会替换本地乐观密文消息，随后解密为明文展示
          pushMessage(msg);
          resolveCipher(msg, peerUid).then(m => {
            if (m.content !== msg.content) updateMessageContent(msg, m.content);
            renderAll();
            scrollToBottom();
          });
        } else {
          // 对方消息：解密后入列（DB 存密文，界面展示明文）
          resolveCipher(msg, peerUid).then(m => {
            const current = conversationsStore.get();
            if (current.chatType !== 'friend' || String(current.chatId) !== String(peerUid)) {
              increment('friend', peerUid);
              return;
            }
            pushMessage(m);
            sendReadReceipts();
            renderAll();
            scrollToBottom();
            requestAnimationFrame(sendReadReceipts);
          });
        }
      } else {
        if (!fromMe) increment('friend', peerUid);
      }
      break;
    }
    case 'group': {
      const gid = msg.group_id;
      if (gid == null) return;
      // 服务端回显给发送方时标记 acked
      if (String(msg.from_uid) === String(myUid)) {
        msg.acked = true;
        pendingMessages.delete(msg.client_msg_id);
        failedMessages.delete(msg.client_msg_id);
        savePending();
      }
      const isCurrent = st.chatType === 'group' && String(st.chatId) === String(gid);
      if (isCurrent) {
        pushMessage(msg);
        if (String(msg.from_uid) !== String(myUid)) sendReadReceipts();
        renderAll();
        scrollToBottom();
        requestAnimationFrame(sendReadReceipts);
      } else {
        increment('group', gid);
      }
      break;
    }
    case 'readReceipt': {
      updateMessageRead(msg.msg_id, true);
      renderAll();
      break;
    }
    case 'groupReadReceipt': {
      if (st.chatType === 'group') {
        updateMessageRead(msg.msg_id, true, msg.read_count, msg.total_count);
        renderAll();
      }
      break;
    }
    case 'presence': {
      // 好友上下线实时更新：侧栏绿/灰点 + 当前会话对方头像
      updateFriendOnline(msg.uid, msg.online);
      groupsStore.set({ members: groupsStore.get().members.map(m => String(m.uid) === String(msg.uid) ? { ...m, is_online: !!msg.online } : m) });
      if (st.chatType === 'friend' && String(st.chatId) === String(msg.uid)) {
        peerOnline = !!msg.online;
      }
      render();
      renderAll();
      break;
    }
    default:
      break;
  }
}

// ============ 未读数 ============
async function loadUnread() {
  try {
    const counts = await privateUnreadCounts();
    setPrivateUnread(counts);
  } catch (e) {
    // 忽略
  }
}

// ============ 搜索用户 ============
export async function searchUser() {
  const keyword = searchInput.value.trim();
  if (!keyword) return;
  try {
    const results = await friendSearch(keyword);
    openModal('搜索结果', () => {
      const wrap = document.createElement('div');
      if (!results || !results.length) {
        wrap.textContent = '未找到用户';
        return wrap;
      }
      results.forEach(u => {
        const row = document.createElement('div');
        row.className = 'search-result-row';
        const info = document.createElement('span');
        info.textContent = (u.nickname || u.username) + ' (@' + u.username + ')';
        const addBtn = document.createElement('button');
        addBtn.className = 'add-btn';
        addBtn.textContent = '添加';
        addBtn.addEventListener('click', async () => {
          try {
            await friendAdd(u.uid);
            addBtn.textContent = '已发送';
            addBtn.disabled = true;
          } catch (e) {
            Toast(e.message);
          }
        });
        row.appendChild(info);
        row.appendChild(addBtn);
        wrap.appendChild(row);
      });
      return wrap;
    }, [{ text: '关闭', onClick: () => true }]);
  } catch (e) {
    Toast(e.message);
  }
}

// ============ 全局事件绑定 ============
document.getElementById('searchBtn')?.addEventListener('click', searchUser);
searchInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') searchUser();
});
document.getElementById('logoutBtn')?.addEventListener('click', logout);
document.getElementById('railLogout')?.addEventListener('click', logout);
document.getElementById('clearBtn')?.addEventListener('click', clearMessages);
document.getElementById('menuBtn')?.addEventListener('click', () => toggleSidebar(true));
document.getElementById('mobileBackdrop')?.addEventListener('click', () => toggleSidebar(false));

function toggleSidebar(open) {
  const sidebar = document.getElementById('sidebar');
  const backdrop = document.getElementById('mobileBackdrop');
  sidebar.classList.toggle('open', open);
  backdrop.classList.toggle('show', open);
  backdrop.style.opacity = open ? '1' : '0';
}
