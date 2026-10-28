// Keep recipient receipts until the server confirms it persisted each one.
export function createDeliveryQueue(uid, sendReceipt, storage = localStorage) {
  const key = 'im_delivery_receipts:' + uid;
  let pending = new Map();
  try {
    const saved = JSON.parse(storage.getItem(key) || '[]');
    if (Array.isArray(saved)) {
      pending = new Map(saved.filter(item => Array.isArray(item) && item.length === 2));
    }
  } catch (e) {
    pending = new Map();
  }
  const persist = () => {
    try { storage.setItem(key, JSON.stringify([...pending])); } catch (e) { console.warn('delivery receipt queue not saved', e); }
  };
  const transmit = (id, sender) => sendReceipt({ type: 'delivery_receipt', msg_id: id, from_uid: sender });
  return {
    enqueue(msg) {
      if (!msg.msg_id || !msg.from_uid || String(msg.from_uid) === String(uid) || String(msg.to_uid) !== String(uid)) return;
      const id = String(msg.msg_id);
      if (pending.has(id)) return;
      pending.set(id, String(msg.from_uid));
      persist();
      transmit(id, String(msg.from_uid));
    },
    acknowledge(id) {
      if (pending.delete(String(id))) persist();
    },
    flush() {
      for (const [id, sender] of pending) transmit(id, sender);
    },
    size() { return pending.size; }
  };
}
