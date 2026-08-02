// WebSocket 客户端：连接 / 心跳 / 自动重连 / 多行 JSON 解析 / 事件分发
//
// 对外 emit：connect/disconnect/private/group/readReceipt/groupReadReceipt

import { authStore } from '../store/auth.js';

let ws = null;
let heartbeatTimer = null;
let reconnectTimer = null;
let connecting = false;
let closedByUser = false;
const listeners = new Set();

function wsRoomURL(token) {
  const cfg = (typeof window.IM_CONFIG !== 'undefined' && window.IM_CONFIG) || {};
  let origin;
  if (cfg.API_ORIGIN) {
    try {
      const u = new URL(cfg.API_ORIGIN);
      origin = (u.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + u.host;
    } catch (e) {
      origin = '';
    }
  }
  if (!origin) {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    origin = proto + '//' + location.host;
  }
  return origin + '/room?room=default&token=' + encodeURIComponent(token);
}

function startHeartbeat() {
  stopHeartbeat();
  heartbeatTimer = setInterval(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'heartbeat', ts: Date.now() }));
    }
  }, 20000);
}

function stopHeartbeat() {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer);
    heartbeatTimer = null;
  }
}

function emit(type, payload) {
  listeners.forEach(fn => fn(type, payload));
}

export function onMessage(fn) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export function connect() {
  const token = authStore.get().token;
  if (!token) return;
  if (connecting) return;
  stopHeartbeat();
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (ws) {
    ws.onclose = null;
    ws.close();
    ws = null;
  }
  connecting = true;
  closedByUser = false;
  ws = new WebSocket(wsRoomURL(token));

  ws.onmessage = (ev) => {
    String(ev.data).split('\n').filter(Boolean).forEach((line) => {
      try {
        handleMessage(JSON.parse(line));
      } catch (e) {
        // 忽略无法解析的行
      }
    });
  };

  ws.onopen = () => {
    connecting = false;
    startHeartbeat();
    emit('connect');
  };

  ws.onclose = () => {
    connecting = false;
    stopHeartbeat();
    emit('disconnect');
    if (!closedByUser) {
      reconnectTimer = setTimeout(connect, 3000);
    }
  };

  ws.onerror = () => {
    connecting = false;
  };
}

export function disconnect() {
  closedByUser = true;
  stopHeartbeat();
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (ws) {
    ws.onclose = null;
    ws.close();
    ws = null;
  }
}

export function send(obj) {
  if (!ws || ws.readyState !== WebSocket.OPEN) return false;
  ws.send(JSON.stringify(obj));
  return true;
}

export function isOpen() {
  return !!(ws && ws.readyState === WebSocket.OPEN);
}

function handleMessage(msg) {
  switch (msg.type) {
    case 'private_message':
      emit('private', msg);
      break;
    case 'group_message':
      emit('group', msg);
      break;
    case 'read_receipt':
      emit('readReceipt', msg);
      break;
    case 'group_read_receipt':
      emit('groupReadReceipt', msg);
      break;
    default:
      emit('message', msg);
  }
}
