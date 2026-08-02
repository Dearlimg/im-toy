// 页面层：主聊天页编排入口
// 职责：初始化组件、串联 store/services、注册 WS 事件、处理会话切换与发送

import { authStore, loadSession, clearSession } from '../store/auth.js';
import { setCurrentFriend, updateFriendOnline, findFriend } from '../store/friends.js';
import { setCurrentGroup } from '../store/groups.js';
import { conversationsStore, openConversation, pushMessage, updateMessageRead, updateMessageContent } from '../store/conversations.js';
import { ensureE2EE, encryptContent, decryptContent, isEncrypted } from '../services/e2ee.js';
import { unreadStore, setPrivateUnread, increment, clear, totalCount } from '../store/unread.js';

import { initAuthForm } from '../components/business/AuthForm.js';
import { setHandlers as setSidebarHandlers, loadAll, switchSide, render } from '../components/business/Sidebar.js';
import { updateHeader, renderAll, scrollToBottom, setUidProvider, setPeerOnlineProvider, clearMessages } from '../components/business/ChatWindow.js';
import { initInput, setSendHandler } from '../components/business/ChatInput.js';
import { Toast } from '../components/common/Toast.js';

import { privateConversation, groupMessages, privateUnreadCounts, friendSearch, friendAdd } from '../services/api.js';
import { openModal } from '../components/common/Modal.js';
import { connect, disconnect, onMessage, send } from '../services/ws.js';

const app = document.getElementById('app');
const authPage = document.getElementById('authPage');
const meNick = document.getElementById('meNick');
const meAvatar = document.getElementById('meAvatar');
const searchInput = document.getElementById('searchInput');

let myUid = null;
setUidProvider(() => myUid);

// 当前会话对方是否在线（由 selectFriend/selectGroup 时设置）
let peerOnline = false;
setPeerOnlineProvider(() => peerOnline);

// ============ 启动 ============
document.addEventListener('DOMContentLoaded', () => {
  loadSession();
  initInput();
  setSendHandler(handleSend);
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
  meNick.textContent = u.nickname || u.username || '我';
  meAvatar.textContent = (u.nickname || u.username || '我').charAt(0).toUpperCase();

  connect();
  await ensureE2EE();
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
  document.querySelectorAll('.side-tabs button').forEach(b => {
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
  sendReadReceipts();
  render();
  renderAll();
}

async function selectGroup(g) {
  setCurrentGroup(g);
  openConversation('group', g.group_id);
  clear('group', g.group_id);
  peerOnline = false;
  updateHeader('group', g);
  await loadHistory('group', g.group_id);
  render();
  renderAll();
}

async function loadHistory(chatType, id) {
  try {
    const msgs = chatType === 'friend'
      ? await privateConversation(id)
      : await groupMessages(id, 50);
    let list = msgs || [];
    // 好友会话历史消息解密（E2EE 密文 → 明文）
    if (chatType === 'friend') {
      const peer = findFriend(id);
      if (peer && peer.pubkey) {
        list = await Promise.all(list.map(async m => {
          if (isEncrypted(m.content)) {
            const plain = await decryptContent(peer.pubkey, m.content);
            if (plain != null) return Object.assign({}, m, { content: plain });
          }
          return m;
        }));
      }
    }
    conversationsStore.set({ messages: list });
    renderAll();
  } catch (e) {
    Toast('加载历史消息失败: ' + e.message);
  }
}

// 对当前会话中"对方发来且未读"的消息补发已读回执，
// 让对方实时看到单对号 → 双对号（覆盖离线期间收到的消息）
function sendReadReceipts() {
  const st = conversationsStore.get();
  (st.messages || []).forEach(m => {
    if (String(m.from_uid) !== String(myUid) && !m.is_read) {
      send({ type: 'read_receipt', msg_id: m.msg_id, from_uid: m.from_uid });
    }
  });
}

// ============ 发送消息 ============
function handleSend(content) {
  const st = conversationsStore.get();
  if (!st.chatType) {
    Toast('请先选择一个会话');
    return;
  }
  const ok = st.chatType === 'friend'
    ? send({ type: 'private_message', to_uid: st.chatId, content })
    : send({ type: 'group_message', group_id: st.chatId, content });
  if (!ok) {
    Toast('发送失败：未连接到服务器');
    return;
  }
  const u = authStore.get().user || {};
  pushMessage({
    from_uid: u.uid,
    from_nickname: u.nickname || u.username,
    content,
    send_time: Date.now(),
    acked: false,
    is_read: false
  });
  renderAll();
  scrollToBottom();
}

// ============ WS 事件分发 ============
function handleWsEvent(type, msg) {
  const st = conversationsStore.get();
  switch (type) {
    case 'private': {
      const fromMe = String(msg.from_uid) === String(myUid);
      const toMe = String(msg.to_uid) === String(myUid);
      if (!fromMe && !toMe) return;
      // 服务端回显给发送方时标记 acked
      if (fromMe) msg.acked = true;
      // 当前会话的 chatId 是对方 uid，回显时 from_uid=自己 to_uid=对方
      const peerUid = fromMe ? msg.to_uid : msg.from_uid;
      const isCurrent = st.chatType === 'friend' && String(st.chatId) === String(peerUid);
      if (isCurrent) {
        pushMessage(msg);
        if (!fromMe) {
          send({ type: 'read_receipt', msg_id: msg.msg_id, from_uid: msg.from_uid });
        }
        renderAll();
        scrollToBottom();
      } else {
        increment('friend', peerUid);
      }
      break;
    }
    case 'group': {
      const gid = msg.group_id;
      if (gid == null) return;
      // 服务端回显给发送方时标记 acked
      if (String(msg.from_uid) === String(myUid)) {
        msg.acked = true;
      }
      const isCurrent = st.chatType === 'group' && String(st.chatId) === String(gid);
      if (isCurrent) {
        pushMessage(msg);
        renderAll();
        scrollToBottom();
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
    }, [{ text: '关闭', onClick: () => false }]);
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
