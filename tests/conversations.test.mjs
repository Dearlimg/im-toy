import { test } from 'node:test';
import assert from 'node:assert/strict';
import { conversationsStore, openConversation, pushMessage, mergeMessages } from '../public/store/conversations.js';

test('acknowledgement replaces only the matching optimistic message', () => {
  openConversation('friend', 'bob');
  pushMessage({ from_uid: 'alice', client_msg_id: 'a', content: 'same', send_time: 1 });
  pushMessage({ from_uid: 'alice', client_msg_id: 'b', content: 'same', send_time: 2 });
  pushMessage({ from_uid: 'alice', client_msg_id: 'a', msg_id: 'm1', seq: 1, content: 'same', send_time: 3 });
  assert.equal(conversationsStore.get().messages.length, 2);
  assert.ok(conversationsStore.get().messages.some(m => m.client_msg_id === 'b' && !m.msg_id));
});

test('sync orders by sequence and later acknowledgement clears pending duplicate', () => {
  openConversation('friend', 'bob');
  pushMessage({ from_uid: 'alice', client_msg_id: 'a', content: 'pending', send_time: 20 });
  mergeMessages([{ msg_id: 'm2', seq: 2, send_time: 1 }, { msg_id: 'm1', seq: 1, send_time: 2 }]);
  assert.deepEqual(conversationsStore.get().messages.slice(0, 2).map(m => m.msg_id), ['m1', 'm2']);
  pushMessage({ msg_id: 'm2', seq: 2, client_msg_id: 'a', send_time: 1 });
  assert.equal(conversationsStore.get().messages.length, 2);
});
