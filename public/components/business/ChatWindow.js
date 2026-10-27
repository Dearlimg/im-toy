// 业务组件：聊天窗口（标题 / 消息列表 / 时间分隔 / 已读状态）
// 依赖：store、MessageBubble、Avatar、api

import { renderMessageBubble } from '../common/MessageBubble.js';
import { Avatar } from '../common/Avatar.js';
import { openModal } from '../common/Modal.js';
import { Toast } from '../common/Toast.js';
import { conversationsStore } from '../../store/conversations.js';
import { groupsStore } from '../../store/groups.js';
import { esc } from '../../utils/format.js';
import { groupMembers } from '../../services/api.js';
import { fingerprint } from '../../services/e2ee.js';

const titleEl = document.getElementById('chatTitle');
const messagesEl = document.getElementById('chatMessages');
const emptyTip = document.getElementById('chatEmptyTip');

// 由页面层注入当前登录 uid
let getUid = () => null;
export function setUidProvider(fn) {
  getUid = fn;
}

// 由页面层注入当前会话对方是否在线
let getPeerOnline = () => false;
export function setPeerOnlineProvider(fn) {
  getPeerOnline = fn;
}

let headerVersion = 0;
export function updateHeader(chatType, obj) {
  const version = ++headerVersion;
  const groupAction = document.querySelector('.header-actions .group-action');
  if (groupAction) groupAction.hidden = chatType !== 'group' || !obj;
  if (!obj) {
    titleEl.textContent = '微聊';
    return;
  }
  if (chatType === 'friend') {
    const name = esc(obj.nickname || obj.username);
    titleEl.innerHTML = name + '<small>私聊</small>';
    if (obj.pubkey) {
      // 端到端加密安全码：双方线下/语音比对，防止中间人攻击
      fingerprint(obj.pubkey).then(fp => {
        if (version === headerVersion) titleEl.innerHTML = name + '<small>公钥指纹 · ' + esc(fp) + '</small>';
      }).catch(() => {});
    }
  } else {
    titleEl.innerHTML = esc(obj.name) + '<small>群聊 · 未端到端加密</small>';
    ensureGroupAction();
    const action = document.querySelector('.header-actions .group-action');
    if (action) action.hidden = false;
  }
}

function ensureGroupAction() {
  const headerActions = document.querySelector('.header-actions');
  if (headerActions.querySelector('.group-action')) return;
  const btn = document.createElement('button');
  btn.className = 'icon-btn group-action';
  btn.title = '群成员';
  btn.innerHTML = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>';
  btn.addEventListener('click', async () => {
    const cur = groupsStore.get().current;
    if (!cur) return;
    try {
      const members = await groupMembers(cur.id);
      groupsStore.set({ members: members || [] });
      openModal('群成员', () => {
        const wrap = document.createElement('div');
        (members || []).forEach(m => {
          const row = document.createElement('div');
          row.className = 'member-check';
          row.style.cursor = 'default';
          row.appendChild(Avatar(m.nickname || m.username, { size: 32, fontSize: 13 }));
          const span = document.createElement('span');
          span.textContent = (m.nickname || m.username) + (m.username ? ' (@' + m.username + ')' : '');
          row.appendChild(span);
          wrap.appendChild(row);
        });
        return wrap;
      }, [{ text: '关闭', onClick: () => false }]);
    } catch (e) {
      Toast(e.message);
    }
  });
  headerActions.appendChild(btn);
}

export function clearMessages() {
  conversationsStore.set({ messages: [] });
  renderAll();
}

let renderedConversation = '';
let renderedMessageKeys = new Set();
let renderedFirstKey = '';
let loadOlderHandler = null;
let loadingOlder = false;

export function setLoadOlderHandler(fn) {
  loadOlderHandler = fn;
}

function messageKey(msg) {
  return String(msg.msg_id || msg.client_msg_id || (msg.from_uid + ':' + msg.send_time + ':' + msg.content));
}

export function renderAll() {
  const st = conversationsStore.get();
  const key = st.chatType ? st.chatType + ':' + st.chatId : '';
  const sameConversation = key === renderedConversation;
  const oldKeys = sameConversation ? renderedMessageKeys : new Set();
  const previousFirstKey = sameConversation ? renderedFirstKey : '';
  const previousScrollHeight = messagesEl.scrollHeight;
  const previousScrollTop = messagesEl.scrollTop;
  const wasAtBottom = previousScrollHeight - previousScrollTop - messagesEl.clientHeight < 48;
  const messages = st.messages || [];
  const prepended = previousFirstKey && messages.findIndex(msg => messageKey(msg) === previousFirstKey) > 0;
  messagesEl.innerHTML = '';
  renderedConversation = key;
  renderedMessageKeys = new Set(messages.map(messageKey));
  renderedFirstKey = messages.length ? messageKey(messages[0]) : '';
  if (!st.chatType) {
    messagesEl.appendChild(emptyTip);
    return;
  }
  if (!st.messages.length) {
    const tip = document.createElement('div');
    tip.className = 'empty-tip';
    tip.textContent = '暂无消息，发送第一条吧';
    messagesEl.appendChild(tip);
    return;
  }
  renderMessages(st, sameConversation && oldKeys.size > 0 && !prepended ? oldKeys : null);
  if (prepended) messagesEl.scrollTop = previousScrollTop + messagesEl.scrollHeight - previousScrollHeight;
  else if (!sameConversation || wasAtBottom) scrollToBottom();
  else messagesEl.scrollTop = previousScrollTop;
}

function renderMessages(st, oldKeys) {
  const fragment = document.createDocumentFragment();
  if (st.hasOlder && loadOlderHandler) {
    const button = document.createElement('button');
    button.className = 'load-older';
    button.type = 'button';
    button.textContent = loadingOlder ? '加载中…' : '加载更早消息';
    button.disabled = loadingOlder;
    button.addEventListener('click', async () => {
      if (loadingOlder) return;
      loadingOlder = true;
      button.disabled = true;
      button.textContent = '加载中…';
      try {
        await loadOlderHandler();
      } catch (err) {
        Toast('加载更早消息失败：' + (err.message || '请稍后重试'));
      } finally {
        loadingOlder = false;
        if (button.isConnected) {
          button.disabled = false;
          button.textContent = '加载更早消息';
        }
      }
    });
    fragment.appendChild(button);
  }
  let lastDay = null;
  st.messages.forEach(msg => {
    const d = new Date(msg.send_time);
    const day = d.toDateString();
    if (day !== lastDay) {
      lastDay = day;
      const div = document.createElement('div');
      div.className = 'time-divider';
      const span = document.createElement('span');
      span.textContent = formatDay(d);
      div.appendChild(span);
      fragment.appendChild(div);
    }
    const isSelf = String(msg.from_uid) === String(getUid());
    const { row } = renderMessageBubble(msg, isSelf, getPeerOnline());
    if (oldKeys && !oldKeys.has(messageKey(msg))) row.classList.add('message-enter');
    fragment.appendChild(row);
  });
  messagesEl.appendChild(fragment);
}

export function scrollToBottom() {
  requestAnimationFrame(() => {
    messagesEl.scrollTop = messagesEl.scrollHeight;
  });
}

function formatDay(d) {
  const now = new Date();
  if (d.toDateString() === now.toDateString()) return '今天';
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return y + '-' + m + '-' + day;
}
