// HTTP 请求封装：统一带 token、按 {code, message, data} 约定解析

import { authStore } from '../store/auth.js';

const base = (typeof window.IM_CONFIG !== 'undefined' && window.IM_CONFIG.API_ORIGIN)
  ? String(window.IM_CONFIG.API_ORIGIN).replace(/\/$/, '')
  : '';

export class ApiError extends Error {
  constructor(message) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request(path, method, body) {
  const headers = { 'Content-Type': 'application/json' };
  const token = authStore.get().token;
  if (token) headers.Authorization = 'Bearer ' + token;

  const res = await fetch(base + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined
  });

  let j = null;
  try {
    j = await res.json();
  } catch (e) {
    j = null;
  }

  if (!j || j.code !== 0) {
    throw new ApiError((j && (j.message || j.error)) || ('HTTP ' + res.status));
  }
  return j.data;
}

export const http = {
  get: (path) => request(path, 'GET'),
  post: (path, body) => request(path, 'POST', body || {}),
  put: (path, body) => request(path, 'PUT', body || {}),
  delete: (path, body) => request(path, 'DELETE', body || {})
};
