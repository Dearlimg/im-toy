// 通用组件：头像（首字 + 在线点）

import { avatarChar } from '../../utils/format.js';

export function Avatar(name, opts) {
  opts = opts || {};
  const div = document.createElement('div');
  div.className = 'avatar' + (opts.online ? ' online' : '');
  if (opts.size) div.style.width = div.style.height = opts.size + 'px';
  div.style.fontSize = opts.fontSize ? opts.fontSize + 'px' : '15px';
  div.textContent = avatarChar(name);
  return div;
}
