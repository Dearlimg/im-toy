// 通用组件：单条消息气泡（区分自己/对方、文本/图片/表情）
// 返回 { row, bubble }，业务层负责插入与已读状态

import { Avatar } from './Avatar.js';
import { esc, isImageUrl, fullTime } from '../../utils/format.js';
import { isEncrypted } from '../../services/e2ee.js';
import { messageStatus } from '../../services/message-status.js';

export function renderMessageBubble(msg, isSelf, peerOnline) {
  const row = document.createElement('div');
  row.className = 'msg-row' + (isSelf ? ' self' : ' other');
  if (msg.msg_id) row.dataset.msgId = String(msg.msg_id);

  row.appendChild(Avatar(isSelf ? '' : (msg.from_nickname || msg.from_username), {
    online: !isSelf && !!peerOnline,
    offline: !isSelf && !peerOnline
  }));

  const main = document.createElement('div');
  main.className = 'msg-main';

  if (!isSelf) {
    const name = document.createElement('div');
    name.className = 'msg-name';
    name.textContent = msg.from_nickname || msg.from_username || '';
    main.appendChild(name);
  }

  const bubble = document.createElement('div');
  bubble.className = 'bubble';

  if (isEncrypted(msg.content)) {
    // E2EE 密文：解密完成后会被替换为明文并重渲染，此处占位
    const span = document.createElement('span');
    span.className = 'enc-msg';
    span.textContent = '\u{1F512} 加密消息';
    bubble.appendChild(span);
  } else if (isImageUrl(msg.content)) {
    const img = document.createElement('img');
    img.src = msg.content;
    img.alt = '图片';
    img.loading = 'lazy';
    img.addEventListener('click', () => openImageModal(img.src));
    bubble.appendChild(img);
  } else if (isPureEmoji(msg.content)) {
    const span = document.createElement('span');
    span.className = 'emoji-char';
    span.textContent = msg.content;
    bubble.appendChild(span);
  } else {
    const text = document.createElement('div');
    text.innerHTML = esc(msg.content).replace(/\n/g, '<br>');
    bubble.appendChild(text);
  }

  const meta = document.createElement('div');
  meta.className = 'msg-meta';
  const timestamp = document.createElement('time');
  timestamp.className = 'msg-time';
  const date = new Date(msg.send_time);
  if (!Number.isNaN(date.getTime())) {
    timestamp.textContent = date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false });
    timestamp.title = fullTime(msg.send_time);
  }
  meta.appendChild(timestamp);
  const status = document.createElement('div');
  status.className = 'msg-status';
  if (isSelf) {
    const state = messageStatus(msg);
    if (state === 'failed') {
      const retry = document.createElement('button');
      retry.type = 'button';
      retry.className = 'retry-message';
      retry.textContent = '发送失败，点击重试';
      retry.addEventListener('click', () => window.dispatchEvent(new CustomEvent('im-retry-message', {
        detail: { clientMsgID: msg.client_msg_id }
      })));
      status.appendChild(retry);
    } else if (state === 'group') {
      status.textContent = '已读 ' + msg.read_count + '/' + (msg.total_count ?? 0);
    } else if (state === 'read' || state === 'delivered') {
      status.appendChild(renderCheck(state));
    } else {
      status.textContent = state === 'sent' ? '已发送' : '发送中';
    }
    meta.appendChild(status);
  }
  bubble.appendChild(meta);
  main.appendChild(bubble);

  row.appendChild(main);
  return { row, status, bubble };
}

function renderCheck(state) {
  const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.setAttribute('viewBox', state === 'read' ? '0 0 20 14' : '0 0 14 14');
  svg.setAttribute('class', 'msg-check' + (state === 'read' ? ' msg-check-read' : ''));
  svg.setAttribute('role', 'img');
  svg.setAttribute('aria-label', state === 'read' ? '对方已读' : '对方已接收');
  const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
  path.setAttribute('d', state === 'read' ? 'M1 7.2l3.5 3.5L12 3M7 7.2l3.5 3.5L18 3' : 'M1 7.2l3.5 3.5L12 3');
  path.setAttribute('fill', 'none');
  path.setAttribute('stroke', 'currentColor');
  path.setAttribute('stroke-width', '1.8');
  path.setAttribute('stroke-linecap', 'round');
  path.setAttribute('stroke-linejoin', 'round');
  svg.appendChild(path);
  return svg;
}

function isPureEmoji(content) {
  if (!content || typeof content !== 'string') return false;
  const s = content.trim();
  if (!s || s.length > 8) return false;
  // 仅由 emoji 组成
  return /^(\p{Extended_Pictographic}|\p{Emoji_Presentation}|\u200d|\ufe0f)+$/u.test(s);
}

// 更新已读状态文字
export function setReadStatus(statusEl, read, total) {
  if (read > 0) {
    statusEl.textContent = '已读 ' + read + '/' + total;
    statusEl.classList.add('read');
  } else {
    statusEl.classList.remove('read');
  }
}

function openImageModal(src) {
  let mask = document.getElementById('imgModal');
  if (!mask) {
    mask = document.createElement('div');
    mask.id = 'imgModal';
    mask.addEventListener('click', () => mask.classList.remove('show'));
    document.body.appendChild(mask);
  }
  mask.innerHTML = '';
  const img = document.createElement('img');
  img.src = src;
  mask.appendChild(img);
  mask.classList.add('show');
}
