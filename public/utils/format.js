// 通用工具函数（纯函数，无外部依赖）

// 防 XSS 的 HTML 转义
export function esc(s) {
  const d = document.createElement('div');
  d.textContent = s == null ? '' : String(s);
  return d.innerHTML;
}

// 从任意时间值解析为 Date
export function parseDate(t) {
  if (typeof t === 'number') return new Date(t);
  const n = parseInt(t, 10);
  const d = new Date(n);
  return isNaN(d.getTime()) ? new Date(t) : d;
}

// 本地化完整时间
export function fullTime(t) {
  const d = parseDate(t);
  if (isNaN(d.getTime())) return String(t);
  return d.toLocaleString();
}

// 列表时间：今天显示 HH:mm，否则 MM-DD
export function listTime(t) {
  const d = parseDate(t);
  if (isNaN(d.getTime())) return '';
  const now = new Date();
  const hh = String(d.getHours()).padStart(2, '0');
  const mm = String(d.getMinutes()).padStart(2, '0');
  const sameDay = d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth() && d.getDate() === now.getDate();
  if (sameDay) return hh + ':' + mm;
  return String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

// 未读角标显示（>99 显示 99+）
export function unreadLabel(n) {
  n = n || 0;
  return n > 99 ? '99+' : String(n);
}

// 消息预览：纯文本截断
export function preview(content, maxLen) {
  maxLen = maxLen || 20;
  if (content == null) return '';
  if (isImageUrl(content)) return '[图片]';
  const s = String(content).replace(/\n/g, ' ');
  return s.length > maxLen ? s.slice(0, maxLen) + '…' : s;
}

// 是否图片消息（http(s) 图片 URL 或 base64 图片）
export function isImageUrl(content) {
  if (!content || typeof content !== 'string') return false;
  return /^data:image\//i.test(content) ||
    /^https?:\/\/.+\.(png|jpe?g|gif|webp|bmp|svg)(\?.*)?$/i.test(content.trim());
}

// 头像首字
export function avatarChar(name) {
  const s = name == null ? '' : String(name).trim();
  return s ? s.charAt(0).toUpperCase() : '?';
}
