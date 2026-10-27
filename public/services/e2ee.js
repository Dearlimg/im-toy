// E2EE：ECDH P-256 密钥交换 + AES-GCM 端到端加密
// 密钥对在浏览器本地生成并持久化（localStorage），私钥永不出浏览器；
// 公钥注册到服务器用于好友间交换（公钥公开，服务器拿到无碍）。
// 消息格式：enc.<iv_b64url>.<ct_b64url>，服务器只存/转发"不透明字符串"。
// 注意：清空浏览器存储会丢失私钥，历史加密消息将无法解密（E2EE 固有特性）。
// 算法说明：ECDH P-256 是 Web Crypto 全浏览器支持的标准算法（Chrome 37+）；
// 曾用 X25519（Chrome 124+/Firefox 127+），旧密钥不兼容，故存储 key 版本为 v2。

import { getPubKey, registerPubKey } from './api.js';

const KEY_STORAGE = 'im_e2ee_keypair_v2';
let accountStorage = KEY_STORAGE;
const CURVE = { name: 'ECDH', namedCurve: 'P-256' };
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
  try { return JSON.parse(localStorage.getItem(accountStorage) || 'null'); } catch (e) { return null; }
}

async function generateKeyPair() {
  const kp = await crypto.subtle.generateKey(CURVE, true, ['deriveBits']);
  const rawPub = new Uint8Array(await crypto.subtle.exportKey('raw', kp.publicKey));
  const rawPriv = new Uint8Array(await crypto.subtle.exportKey('pkcs8', kp.privateKey));
  return { pubkey: b64url(rawPub), priv: b64url(rawPriv) };
}

// 生成（如无）并持久化密钥对，返回是否可用
export async function initE2EE() {
  if (!window.crypto || !crypto.subtle) return false;
  if (getKeyPair()) return true;
  try {
    localStorage.setItem(accountStorage, JSON.stringify(await generateKeyPair()));
    return true;
  } catch (e) {
    console.warn('E2EE not supported:', e);
    return false;
  }
}

// 每个账号单独保存设备密钥。新设备替换公钥需要明确确认。
export async function ensureE2EE(uid) {
  if (!uid) return false;
  accountStorage = KEY_STORAGE + ':' + uid;
  let serverKey;
  try {
    serverKey = (await getPubKey()).pubkey || '';
    const legacy = JSON.parse(localStorage.getItem(KEY_STORAGE) || 'null');
    if (!getKeyPair() && legacy && legacy.pubkey === serverKey) {
      localStorage.setItem(accountStorage, JSON.stringify(legacy));
    }
  } catch (e) {
    console.warn('could not check registered key:', e);
    return false;
  }
  const ok = await initE2EE();
  if (!ok) return false;
  const localKey = getKeyPair().pubkey;
  let replace = false;
  if (serverKey && serverKey !== localKey) {
    replace = window.confirm('此账号已有其他设备的加密密钥。本设备没有对应私钥。替换后旧设备及历史密文可能无法解密，好友需要重新核验指纹。确认替换？');
    if (!replace) return false;
  }
  try {
    await registerPubKey(localKey, replace);
  } catch (e) {
    console.warn('pubkey register failed:', e);
    return false;
  }
  return true;
}

async function deriveShared(peerPubB64url) {
  const kp = getKeyPair();
  if (!kp) return null;
  const peerKey = await crypto.subtle.importKey('raw', unb64url(peerPubB64url), CURVE, false, []);
  const privKey = await crypto.subtle.importKey('pkcs8', unb64url(kp.priv), CURVE, false, ['deriveBits']);
  const shared = await crypto.subtle.deriveBits({ name: 'ECDH', public: peerKey }, privKey, 256);
  return new Uint8Array(shared);
}

// 用对方公钥加密；失败返回 null，由调用方阻止发送。
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

// 128 位公钥指纹；需要通过独立可信渠道核对。
export async function fingerprint(pubkey) {
  if (!pubkey) return '';
  try {
    const digest = new Uint8Array(await crypto.subtle.digest('SHA-256', unb64url(pubkey)));
    let hex = '';
    for (let i = 0; i < 16; i++) hex += digest[i].toString(16).padStart(2, '0');
    return hex.toUpperCase().replace(/(.{4})/g, '$1-').replace(/-$/, '');
  } catch (e) {
    return '';
  }
}

// Pin only after the user confirms an out-of-band comparison; changed keys are never silently accepted.
export async function verifyPeerKey(uid, peerUID, pubkey) {
  const pinKey = 'im_e2ee_peer:' + uid + ':' + peerUID;
  const previous = localStorage.getItem(pinKey);
  if (previous === pubkey) return true;
  const peerFingerprint = await fingerprint(pubkey);
  const ownFingerprint = await fingerprint(getKeyPair()?.pubkey);
  if (!peerFingerprint || !ownFingerprint) return false;
  const change = previous ? '好友密钥发生变化，旧消息可能无法解密。' : '首次向该好友发送加密消息。';
  const accepted = window.confirm(change + '\n请通过其他可信渠道与好友核对指纹。\n好友：' + peerFingerprint + '\n自己：' + ownFingerprint + '\n\n双方指纹已核对一致？');
  if (accepted) localStorage.setItem(pinKey, pubkey);
  return accepted;
}
