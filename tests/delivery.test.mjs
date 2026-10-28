import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createDeliveryQueue } from '../public/services/delivery.js';
import { messageStatus } from '../public/services/message-status.js';
import { conversationsStore, openConversation, pushMessage, mergeMessages, applyDeliveryStatuses } from '../public/store/conversations.js';

test('server acknowledgement is distinct from recipient delivery and read', () => {
  assert.equal(messageStatus({}), 'pending');
  assert.equal(messageStatus({ acked: true }), 'sent');
  assert.equal(messageStatus({ acked: true, delivered: true }), 'delivered');
  assert.equal(messageStatus({ acked: true, is_read: true }), 'read');
  assert.equal(messageStatus({ failed: true }), 'failed');
});

test('delivery receipts survive reconnect until the server acknowledges them', () => {
  const values = new Map();
  const storage = { getItem: key => values.get(key) ?? null, setItem: (key, value) => values.set(key, value) };
  const outgoing = [];
  let queue = createDeliveryQueue('bob', msg => outgoing.push(msg), storage);
  queue.enqueue({ msg_id: 'm1', from_uid: 'alice', to_uid: 'bob', content: 'ciphertext' });
  queue.enqueue({ msg_id: 'm1', from_uid: 'alice', to_uid: 'bob' });
  queue.enqueue({ msg_id: 'm2', from_uid: 'mallory', to_uid: 'other' });
  assert.equal(queue.size(), 1);
  assert.equal(outgoing.length, 1);
  queue = createDeliveryQueue('bob', msg => outgoing.push(msg), storage);
  queue.flush();
  assert.deepEqual(outgoing[1], { type: 'delivery_receipt', msg_id: 'm1', from_uid: 'alice' });
  queue.acknowledge('m1');
  assert.equal(createDeliveryQueue('bob', () => {}, storage).size(), 0);
});

test('sync or delayed server updates cannot regress delivered/read status', () => {
  openConversation('friend', 'bob');
  pushMessage({ msg_id: 'm1', seq: 1, from_uid: 'alice', delivered: true, is_read: true });
  mergeMessages([{ msg_id: 'm1', seq: 1, delivered: false, is_read: false }]);
  applyDeliveryStatuses([{ msg_id: 'm1', delivered: false, is_read: false }]);
  assert.equal(messageStatus(conversationsStore.get().messages[0]), 'read');
});
