// E2EE：X25519 密钥交换 + AES-GCM 端到端加密
// 密钥对在浏览器本地生成并持久化（localStorage），私钥永不出浏览器；
// 公钥注册到服务器用于好友间交换（公钥公开，服务器拿到无碍）。
// 消息格式：enc.<iv_b64url>.<ct_b64url>，服务器只存/转发"不透明字符串"。
// 注意：清空浏览器存储会丢失私钥，历史加密消息将无法解密（E2EE 固有特性）。

import { registerPubKey } from './api.js';

const KEY_STORAGE = 'im_e2ee_keypair_v1';
const encoder = new TextEncoder();
const decoder = new TextDecoder();

function b64url(bytes) {
  let bin = '';
  bytes.forEach(b => { bin += String.fromCharCode(b); });
  return btoa(bin).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function unb64url(s) {
  let str = s.replace(/-/g, '+').replace(/_/g, '/');
  while (str.length % 4) str += '=';
  const bin = atob(str);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  return bytes;
}

export function isEncrypted(content) {
  return typeof content === 'string' && content.indexOf('enc.') === 0;
}

function getKeyPair() {
  try { return JSON.parse(localStorage.getItem(KEY_STORAGE) || 'null'); } catch (e) { return null; }
}

async function generateKeyPair() {
  const kp = await crypto.subtle.generateKey({ name: 'X25519' }, true, ['deriveBits']);
  const rawPub = new Uint8Array(await crypto.subtle.exportKey('raw', kp.publicKey));
  const rawPriv = new Uint8Array(await crypto.subtle.exportKey('pkcs8', kp.privateKey));
  return { pubkey: b64url(rawPub), priv: b64url(rawPriv) };
}

// 生成（如无）并持久化密钥对，返回是否可用
export async function initE2EE() {
  if (!window.crypto || !crypto.subtle) return false;
  if (getKeyPair()) return true;
  try {
    localStorage.setItem(KEY_STORAGE, JSON.stringify(await generateKeyPair()));
    return true;
  } catch (e) {
    console.warn('E2EE not supported:', e);
    return false;
  }
}

// 确保密钥对存在并注册公钥到服务器；浏览器不支持 E2EE 时返回 false（降级明文）
export async function ensureE2EE() {
  const ok = await initE2EE();
  if (!ok) return false;
  try {
    await registerPubKey(getKeyPair().pubkey);
  } catch (e) {
    console.warn('pubkey register failed:', e);
  }
  return true;
}

async function deriveShared(peerPubB64url) {
  const kp = getKeyPair();
  if (!kp) return null;
  const peerKey = await crypto.subtle.importKey('raw', unb64url(peerPubB64url), { name: 'X25519' }, false, ['deriveBits']);
  const privKey = await crypto.subtle.importKey('pkcs8', unb64url(kp.priv), { name: 'X25519' }, false, ['deriveBits']);
  const shared = await crypto.subtle.deriveBits({ name: 'X25519', public: peerKey }, privKey, 256);
  return new Uint8Array(shared);
}

// 用对方公钥加密；失败返回 null（调用方降级明文）
export async function encryptContent(peerPub, text) {
  try {
    const shared = await deriveShared(peerPub);
    if (!shared) return null;
    const key = await crypto.subtle.importKey('raw', shared, { name: 'AES-GCM' }, false, ['encrypt']);
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const ct = new Uint8Array(await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, encoder.encode(text)));
    return 'enc.' + b64url(iv) + '.' + b64url(ct);
  } catch (e) {
    console.warn('encrypt failed:', e);
    return null;
  }
}

// 用对方公钥解密 enc. 格式内容；失败返回 null
export async function decryptContent(peerPub, content) {
  try {
    const parts = content.split('.');
    if (parts.length !== 3 || parts[0] !== 'enc') return null;
    const shared = await deriveShared(peerPub);
    if (!shared) return null;
    const key = await crypto.subtle.importKey('raw', shared, { name: 'AES-GCM' }, false, ['decrypt']);
    const plain = await crypto.subtle.decrypt({ name: 'AES-GCM', iv: unb64url(parts[1]) }, key, unb64url(parts[2]));
    return decoder.decode(plain);
  } catch (e) {
    console.warn('decrypt failed:', e);
    return null;
  }
}

// 公钥指纹（SHA-256 前 4 字节 → 8 位 hex，如 A1B2-C3D4），用于双方人工比对防中间人
export async function fingerprint(pubkey) {
  if (!pubkey) return '';
  try {
    const digest = new Uint8Array(await crypto.subtle.digest('SHA-256', unb64url(pubkey)));
    let hex = '';
    for (let i = 0; i < 4; i++) hex += digest[i].toString(16).padStart(2, '0');
    return hex.toUpperCase().replace(/(.{4})/g, '$1-').replace(/-$/, '');
  } catch (e) {
    return '';
  }
}
