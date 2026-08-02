// 业务组件：聊天窗口（标题 / 消息列表 / 时间分隔 / 已读状态）
// 依赖：store、MessageBubble、Avatar、api

import { renderMessageBubble, setReadStatus } from '../common/MessageBubble.js';
import { Avatar } from '../common/Avatar.js';
import { openModal } from '../common/Modal.js';
import { Toast } from '../common/Toast.js';
import { conversationsStore } from '../../store/conversations.js';
import { groupsStore } from '../../store/groups.js';
import { esc } from '../../utils/format.js';
import { groupMembers } from '../../services/api.js';

const titleEl = document.getElementById('chatTitle');
const messagesEl = document.getElementById('chatMessages');
const emptyTip = document.getElementById('chatEmptyTip');

// 由页面层注入当前登录 uid
let getUid = () => null;
export function setUidProvider(fn) {
  getUid = fn;
}

export function updateHeader(chatType, obj) {
  if (!obj) {
    titleEl.innerHTML = '<small>选择一个对话开始聊天</small>';
    return;
  }
  if (chatType === 'friend') {
    titleEl.innerHTML = '<small>' + esc(obj.nickname || obj.username) + '</small>';
  } else {
    titleEl.innerHTML = '<small>' + esc(obj.name) + ' · 群聊</small>';
    ensureGroupAction();
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

export function renderAll() {
  const st = conversationsStore.get();
  messagesEl.innerHTML = '';
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
  renderMessages();
}

function renderMessages() {
  const st = conversationsStore.get();
  messagesEl.innerHTML = '';
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
      messagesEl.appendChild(div);
    }
    const isSelf = String(msg.from_uid) === String(getUid());
    const { row, status } = renderMessageBubble(msg, isSelf);
    if (isSelf) {
      if (msg.read_count && msg.total_count && msg.read_count > 0) {
        setReadStatus(status, msg.read_count, msg.total_count);
      } else if (msg.is_read) {
        status.textContent = '已读';
        status.classList.add('read');
      } else {
        status.textContent = '已发送';
      }
    }
    messagesEl.appendChild(row);
  });
  scrollToBottom();
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
