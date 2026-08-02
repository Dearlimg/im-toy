// 通用组件：单条消息气泡（区分自己/对方、文本/图片/表情）
// 返回 { row, bubble }，业务层负责插入与已读状态

import { Avatar } from './Avatar.js';
import { esc, isImageUrl, fullTime } from '../../utils/format.js';

export function renderMessageBubble(msg, isSelf, peerOnline) {
  const row = document.createElement('div');
  row.className = 'msg-row' + (isSelf ? ' self' : ' other');

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

  if (isImageUrl(msg.content)) {
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

  main.appendChild(bubble);

  const status = document.createElement('div');
  status.className = 'msg-status';
  if (isSelf) {
    // 纸飞机风格：单对号=已发送(acked)，双对号=对方已读(is_read)
    const check = document.createElement('span');
    check.className = 'msg-check';
    check.textContent = '\u2713';
    if (msg.is_read) {
      const check2 = document.createElement('span');
      check2.className = 'msg-check msg-check-read';
      check2.textContent = '\u2713';
      status.appendChild(check);
      status.appendChild(check2);
    } else if (msg.acked) {
      status.appendChild(check);
    } else {
      status.textContent = '\u53d1\u9001\u4e2d...';
    }
  } else {
    status.textContent = fullTime(msg.send_time);
  }
  main.appendChild(status);

  row.appendChild(main);
  return { row, status, bubble };
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
