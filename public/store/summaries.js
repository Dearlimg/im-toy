import { createStore } from './index.js';
import { preview } from '../utils/format.js';

export const summariesStore = createStore({ latest: {} });

export function updateSummary(type, id, message) {
  if (!message) return;
  const key = type + ':' + id;
  const current = summariesStore.get().latest[key];
  if (current) {
    const same = message.msg_id && message.msg_id === current.msg_id;
    if (same && !String(current.content).startsWith('enc.') && String(message.content).startsWith('enc.')) return;
    const older = message.seq && current.seq
      ? Number(message.seq) < Number(current.seq)
      : Number(message.send_time) < Number(current.send_time);
    if (!same && older) return;
  }
  summariesStore.set({ latest: { ...summariesStore.get().latest, [key]: message } });
}

export function getSummary(type, id) {
  return summariesStore.get().latest[type + ':' + id];
}

export function summaryText(type, message) {
  if (!message) return '暂无消息';
  const content = String(message.content || '');
  const text = content.startsWith('enc.') ? '[历史消息]' : preview(content, 60);
  const sender = message.from_nickname || message.from_username;
  return type === 'group' && sender ? sender + '：' + text : text;
}
