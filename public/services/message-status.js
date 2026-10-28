// A server save is distinct from confirmation by the recipient's client.
export function messageStatus(msg) {
  if (msg.failed) return 'failed';
  if (msg.read_count != null) return 'group';
  if (msg.is_read) return 'read';
  if (msg.delivered) return 'delivered';
  if (msg.acked || msg.msg_id) return 'sent';
  return 'pending';
}
