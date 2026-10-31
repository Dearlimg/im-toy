import { test } from 'node:test';
import assert from 'node:assert/strict';
import { webcrypto } from 'node:crypto';

const storage = new Map();
globalThis.localStorage = {
  getItem: key => storage.get(key) ?? null,
  setItem: (key, value) => storage.set(key, String(value)),
  removeItem: key => storage.delete(key)
};
globalThis.window = { crypto: webcrypto, confirm: () => true };
Object.defineProperty(globalThis, 'crypto', { value: webcrypto, configurable: true });
const publicKeys = new Map();
let currentUser = '';
globalThis.fetch = async (_url, options) => {
  if (options.method === 'POST') {
    const { pubkey } = JSON.parse(options.body);
    publicKeys.set(currentUser, pubkey);
  }
  return { json: async () => ({ code: 0, data: { pubkey: publicKeys.get(currentUser) || '' } }) };
};

const e2ee = await import('../public/services/e2ee.js');

test('history-only initialization never creates or replaces device keys', async () => {
  currentUser = 'history-only-new-device';
  publicKeys.set(currentUser, 'existing-device-key');
  window.confirm = () => { throw new Error('must not prompt'); };
  assert.equal(await e2ee.ensureE2EE(currentUser, { readOnly: true }), false);
  assert.equal(publicKeys.get(currentUser), 'existing-device-key');
  assert.equal(localStorage.getItem('im_e2ee_keypair_v2:' + currentUser), null);
  window.confirm = () => true;
});

test('accounts encrypt and decrypt, and key changes require confirmation', async () => {
  currentUser = 'alice';
  assert.equal(await e2ee.ensureE2EE(currentUser), true);
  const aliceKey = publicKeys.get('alice');
  currentUser = 'bob';
  assert.equal(await e2ee.ensureE2EE(currentUser), true);
  const bobKey = publicKeys.get('bob');
  assert.notEqual(aliceKey, bobKey);

  currentUser = 'alice';
  await e2ee.ensureE2EE(currentUser);
  assert.equal(await e2ee.verifyPeerKey('alice', 'bob', bobKey), true);
  const ciphertext = await e2ee.encryptContent(bobKey, 'hello');
  assert.ok(ciphertext.startsWith('enc.'));
  currentUser = 'bob';
  await e2ee.ensureE2EE(currentUser);
  assert.equal(await e2ee.decryptContent(aliceKey, ciphertext), 'hello');

  localStorage.removeItem('im_e2ee_keypair_v2:bob');
  window.confirm = () => false;
  assert.equal(await e2ee.ensureE2EE('bob'), false);
  assert.equal(publicKeys.get('bob'), bobKey);
  assert.equal(await e2ee.verifyPeerKey('alice', 'bob', aliceKey), false);
});
