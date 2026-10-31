import { test } from 'node:test';
import assert from 'node:assert/strict';

globalThis.window = {};
class Element {
  children = []; style = {}; dataset = {}; className = '';
  get innerHTML() { return this.textContent || ''; }
  set innerHTML(value) { this.textContent = value; }
  appendChild(child) { this.children.push(child); }
  addEventListener() {}
}
globalThis.document = { createElement: () => new Element() };
const { renderMessageBubble } = await import('../public/components/common/MessageBubble.js');

test('unknown group presence is not rendered as offline', () => {
  const msg = { from_username: 'user1', content: 'hello', send_time: Date.now() };
  const avatar = online => renderMessageBubble(msg, false, online).row.children[0];
  assert.equal(avatar(undefined).className, 'avatar');
  assert.equal(avatar(true).className, 'avatar online');
  assert.equal(avatar(false).className, 'avatar offline');
  assert.equal(renderMessageBubble(msg, true, false).row.children[0].className, 'avatar');
});
