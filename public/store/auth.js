// 会话状态：token、当前用户、在线状态

import { createStore } from './index.js';

export const authStore = createStore({
  token: '',
  user: null,      // { uid, username, nickname }
  online: false
});

const TOKEN_KEY = 'im_token';
const USER_KEY = 'im_user';

export function loadSession() {
  const token = localStorage.getItem(TOKEN_KEY);
  const user = localStorage.getItem(USER_KEY);
  if (token) {
    authStore.set({ token, user: user ? JSON.parse(user) : null });
  }
  return authStore.get();
}

export function saveSession(token, user) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
  authStore.set({ token, user, online: true });
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  authStore.set({ token: '', user: null, online: false });
}

export function setOnline(v) {
  authStore.set({ online: v });
}
