import { test } from 'node:test';
import assert from 'node:assert/strict';
import { summariesStore, updateSummary, getSummary, summaryText } from '../public/store/summaries.js';
import { unreadStore, increment, clear, totalCount, getCount } from '../public/store/unread.js';

test('latest summaries remain isolated by conversation and reject older history', () => {
  summariesStore.set({ latest: {} });
  updateSummary('group', 'one', { msg_id: 'new', seq: 5, content: '最新消息', send_time: 50 });
  updateSummary('group', 'one', { msg_id: 'old', seq: 4, content: '旧消息', send_time: 60 });
  updateSummary('friend', 'one', { content: '私聊', send_time: 10 });
  assert.equal(getSummary('group', 'one').content, '最新消息');
  assert.equal(getSummary('friend', 'one').content, '私聊');
  updateSummary('group', 'one', { msg_id: 'next', seq: 6, content: '下一条', send_time: 70 });
  assert.equal(getSummary('group', 'one').content, '下一条');
});

test('summaries show image labels and sender names without exposing historical ciphertext', () => {
  assert.equal(summaryText('friend', { content: 'data:image/png;base64,aaa' }), '[图片]');
  assert.equal(summaryText('group', { from_username: 'user1', content: '你好\n朋友' }), 'user1：你好 朋友');
  assert.equal(summaryText('friend', { content: 'enc.secret' }), '[历史消息]');
  assert.equal(summaryText('group', null), '暂无消息');
  updateSummary('friend', 'encrypted', { msg_id: 'e1', content: '解密后的消息', send_time: 1 });
  updateSummary('friend', 'encrypted', { msg_id: 'e1', content: 'enc.secret', send_time: 1 });
  assert.equal(getSummary('friend', 'encrypted').content, '解密后的消息');
});

test('navigation totals aggregate each type and opening one conversation preserves other unread counts', () => {
  unreadStore.set({ counts: {} });
  increment('group', 'one');
  increment('group', 'one');
  increment('group', 'two');
  increment('friend', 'one');
  assert.equal(totalCount('group'), 3);
  assert.equal(totalCount('friend'), 1);
  assert.equal(totalCount(), 4);
  clear('group', 'one');
  assert.equal(getCount('group', 'one'), 0);
  assert.equal(totalCount('group'), 1);
  assert.equal(totalCount('friend'), 1);
});
