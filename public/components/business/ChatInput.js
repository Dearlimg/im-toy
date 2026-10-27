// 业务组件：输入区（文本 / 表情 / 图片 / 回车发送）
// 依赖：EmojiPicker、Toast、validator

import { createEmojiPanel } from '../common/EmojiPicker.js';
import { Toast } from '../common/Toast.js';
import { validMessage } from '../../utils/validator.js';

const inputEl = document.getElementById('messageInput');
const sendBtn = document.getElementById('sendBtn');
const emojiBtn = document.getElementById('emojiBtn');
const imageBtn = document.getElementById('imageBtn');
const fileInput = document.getElementById('fileInput');

let panel = null;
let onSend = null;

export function setSendHandler(fn) {
  onSend = fn;
}

export function initInput() {
  const chatArea = document.getElementById('chatArea');
  panel = createEmojiPanel(pickEmoji);
  panel.style.display = 'none';
  chatArea.appendChild(panel);

  emojiBtn.addEventListener('click', () => {
    togglePanel();
  });

  imageBtn.addEventListener('click', () => {
    fileInput.click();
  });

  fileInput.addEventListener('change', (e) => {
    const file = e.target.files && e.target.files[0];
    if (!file) return;
    if (!file.type.startsWith('image/')) {
      Toast('请选择图片文件');
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      compressImage(reader.result, async (dataUrl) => {
        if (onSend && await onSend(dataUrl)) Toast('图片正在发送');
      });
    };
    reader.readAsDataURL(file);
    e.target.value = '';
  });

  inputEl.addEventListener('input', () => {
    autoResize();
    togglePanel(false);
  });

  inputEl.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      doSend();
    }
  });

  sendBtn.addEventListener('click', doSend);

  document.addEventListener('click', (e) => {
    if (panel && panel.classList.contains('show') &&
      !panel.contains(e.target) && e.target !== emojiBtn && !emojiBtn.contains(e.target)) {
      togglePanel(false);
    }
  });
}

async function doSend() {
  const content = inputEl.value.trim();
  if (!validMessage(content)) return;
  if (onSend && await onSend(content)) {
    // Preserve new typing entered while encryption/verification was pending.
    if (inputEl.value.trim() === content) inputEl.value = '';
    autoResize();
  }
}

function pickEmoji(e) {
  inputEl.value += e;
  inputEl.focus();
  autoResize();
}

function togglePanel(force) {
  const show = force != null ? force : !panel.classList.contains('show');
  panel.classList.toggle('show', show);
  panel.style.display = show ? 'grid' : 'none';
}

function autoResize() {
  inputEl.style.height = 'auto';
  inputEl.style.height = Math.min(inputEl.scrollHeight, 120) + 'px';
}

// 图片压缩到最大 800px，JPEG 0.8
function compressImage(dataUrl, cb) {
  const img = new Image();
  img.onload = () => {
    const max = 800;
    let w = img.width;
    let h = img.height;
    if (Math.max(w, h) > max) {
      const ratio = max / Math.max(w, h);
      w = Math.round(w * ratio);
      h = Math.round(h * ratio);
    }
    const canvas = document.createElement('canvas');
    canvas.width = w;
    canvas.height = h;
    canvas.getContext('2d').drawImage(img, 0, 0, w, h);
    cb(canvas.toDataURL('image/jpeg', 0.8));
  };
  img.src = dataUrl;
}
