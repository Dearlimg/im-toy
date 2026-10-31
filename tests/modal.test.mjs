import { test } from 'node:test';
import assert from 'node:assert/strict';
import { openModal } from '../public/components/common/Modal.js';

class Element {
  children = [];
  handlers = {};
  disabled = false;
  removed = false;
  style = {};
  appendChild(child) { this.children.push(child); }
  addEventListener(type, callback) { this.handlers[type] = callback; }
  remove() { this.removed = true; }
}
globalThis.document = { createElement: () => new Element(), body: new Element() };

test('modal waits for submission, blocks double clicks, and retains validation failures', async () => {
  let resolve;
  let calls = 0;
  const result = new Promise(r => { resolve = r; });
  const dialog = openModal('create', null, [{ text: 'submit', onClick: () => { calls++; return result; } }]);
  const button = dialog.modal.children[1].children[0];
  const pending = button.handlers.click();
  await button.handlers.click();
  assert.equal(calls, 1);
  assert.equal(dialog.mask.removed, false);
  resolve(false);
  await pending;
  assert.equal(dialog.mask.removed, false);
  assert.equal(button.disabled, false);
});

test('modal closes only after successful async submission or explicit close', async () => {
  const dialog = openModal('create', null, [{ text: 'submit', onClick: async () => true }]);
  await dialog.modal.children[1].children[0].handlers.click();
  assert.equal(dialog.mask.removed, true);
  const cancel = openModal('create', null, [{ text: 'cancel', onClick: () => true }]);
  await cancel.modal.children[1].children[0].handlers.click();
  assert.equal(cancel.mask.removed, true);
});
