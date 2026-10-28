// 聊天室性能压测：群聊（WS 广播）+ 私聊（HTTP 发送 / WS 发送）
//
// 运行（自签证书需忽略 TLS 校验，仅本进程生效）：
//   $env:NODE_TLS_REJECT_UNAUTHORIZED='0'; node tests/load_test.mjs
//
// 前置：目标服务可达。脚本自动注册压测账号并建群。
// 可用环境变量覆盖：BENCH_BASE、BENCH_GROUP_SENDERS、BENCH_GROUP_RATE、
//   BENCH_GROUP_DURATION、BENCH_PM_CONCURRENCY、BENCH_PM_DURATION

const BASE = process.env.BENCH_BASE || 'https://121.40.235.227';
const PREFIX = 'load_' + Date.now().toString(36).slice(-5); // 账号前缀，便于识别清理
const GROUP_SENDERS = +(process.env.BENCH_GROUP_SENDERS || 5);    // 群聊 WS 发送者数量
const GROUP_RATE = +(process.env.BENCH_GROUP_RATE || 4);          // 每个发送者每秒消息数
const GROUP_DURATION = +(process.env.BENCH_GROUP_DURATION || 20); // 群聊阶段秒数
const PM_CONCURRENCY = +(process.env.BENCH_PM_CONCURRENCY || 10); // 私聊 HTTP 并发
const PM_DURATION = +(process.env.BENCH_PM_DURATION || 20);       // 私聊阶段秒数
const PM_WS_COUNT = +(process.env.BENCH_PM_WS_COUNT || 6);        // 私聊 WS 发送条数（验证 ack 与限流路径）

// ---------- HTTP 工具 ----------
async function api(method, path, token, body) {
  const t0 = performance.now();
  try {
    const res = await fetch(BASE + path, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: 'Bearer ' + token } : {}),
      },
      body: body ? JSON.stringify(body) : undefined,
    });
    const rtt = performance.now() - t0;
    const json = await res.json().catch(() => ({}));
    return { rtt, ok: res.ok && (json.code === 0 || json.code === 200), json };
  } catch (e) {
    return { rtt: performance.now() - t0, ok: false, json: { message: 'network: ' + (e.cause?.code || e.cause?.message || e.message) } };
  }
}

function fakeCipher() {
  // 模拟 E2EE 密文（服务端透传不解析），约 120 字符
  return 'enc:' + Buffer.from('msg-' + Math.random().toString(36).repeat(15)).toString('base64').slice(0, 110);
}

let seqCounter = 0;
const msgId = (tag) => `${tag}-${process.pid}-${Date.now().toString(36)}-${(seqCounter++).toString(36)}`;

// 全局发送时刻表：让接收方连接也能按 client_msg_id 匹配出端到端延迟
const sentTimes = new Map();
const trackSend = (id) => { const t = performance.now(); sentTimes.set(id, t); return t; };

// ---------- WS 客户端 ----------
async function connectWS(token, label) {
  // 公网链路偶发丢 SYN 导致握手卡 10 秒，重试绕开
  for (let attempt = 1; ; attempt++) {
    try {
      return await connectWSOnce(token, label);
    } catch (e) {
      if (attempt >= 3) throw e;
      console.log(`  [${label}] 握手失败（${e.message}），第 ${attempt} 次重试`);
      await sleep(500);
    }
  }
}

function connectWSOnce(token, label) {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(`wss://121.40.235.227/room?room=default&token=${encodeURIComponent(token)}`);
    const t0 = performance.now();
    ws._pending = new Map(); // client_msg_id -> send timestamp
    ws._stats = { label, sent: 0, failed: 0, e2e: [] };
    ws.onopen = () => resolve(ws);
    ws.onerror = () => reject(new Error(`${label} ws error`));
    setTimeout(() => reject(new Error(`${label} ws connect timeout`)), 10000);
    ws.onmessage = (ev) => {
      String(ev.data).split('\n').filter(Boolean).forEach((line) => {
        let m; try { m = JSON.parse(line); } catch { return; }
        if ((m.type === 'group_message' || m.type === 'private_message') && m.client_msg_id) {
          const sent = ws._pending.get(m.client_msg_id) ?? sentTimes.get(m.client_msg_id);
          if (sent !== undefined) {
            ws._pending.delete(m.client_msg_id);
            sentTimes.delete(m.client_msg_id);
            ws._stats.e2e.push(performance.now() - sent);
          }
        } else if (m.type === 'send_failed') {
          ws._stats.failed++;
        }
      });
    };
    ws._t0 = t0;
  });
}

const sendWS = (ws, obj) => {
  ws.send(JSON.stringify(obj));
  const t = trackSend(obj.client_msg_id);
  ws._pending.set(obj.client_msg_id, t);
  ws._stats.sent++;
};

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// ---------- 统计 ----------
function pct(arr, p) {
  if (!arr.length) return NaN;
  const s = [...arr].sort((a, b) => a - b);
  return s[Math.min(s.length - 1, Math.floor((s.length * p) / 100))];
}
function report(name, stats) {
  const e2e = stats.e2e;
  const loss = stats.sent - e2e.length - 0; // 未在阶段结束时收回的视为在途
  console.log(`\n[${name}] 发送=${stats.sent} send_failed=${stats.failed} ` +
    `确认=${e2e.length} (在途/未匹配=${loss})`);
  if (e2e.length) {
    console.log(`  端到端延迟 ms: P50=${pct(e2e, 50).toFixed(0)} P90=${pct(e2e, 90).toFixed(0)} ` +
      `P95=${pct(e2e, 95).toFixed(0)} P99=${pct(e2e, 99).toFixed(0)} max=${Math.max(...e2e).toFixed(0)}`);
  }
}

// ---------- 主流程 ----------
const users = []; // {username, uid, token}

async function prepare() {
  console.log(`== 准备阶段：账号前缀 ${PREFIX} ==`);
  const names = ['a', 'b', ...Array.from({ length: GROUP_SENDERS }, (_, i) => 'g' + i)];
  for (const suffix of names) {
    const username = PREFIX + suffix;
    const reg = await api('POST', '/api/user/register', null, {
      username, password: 'bench123', nickname: '压测' + suffix,
    });
    if (!reg.ok && !/已存在|exist|taken/i.test(reg.json.message || '')) {
      throw new Error(`register ${username}: ${JSON.stringify(reg.json)}`);
    }
    const login = await api('POST', '/api/user/login', null, { username, password: 'bench123' });
    if (!login.ok) throw new Error(`login ${username}: ${JSON.stringify(login.json)}`);
    users[suffix] = { username, uid: login.json.data.user.uid, token: login.json.data.token };
  }
  const g0 = users['g0'];
  const create = await api('POST', '/api/group/create', g0.token, {
    name: '压测群' + PREFIX,
    member_uids: Array.from({ length: GROUP_SENDERS - 1 }, (_, i) => users['g' + (i + 1)].uid),
  });
  if (!create.ok) throw new Error('create group: ' + JSON.stringify(create.json));
  users.groupID = create.json.data.group_id;
  console.log(`群已创建: ${users.groupID}，成员 ${GROUP_SENDERS} 人`);
}

async function phaseGroup() {
  console.log(`\n== 阶段一：群聊 WS，${GROUP_SENDERS} 人 × ${GROUP_RATE}/s × ${GROUP_DURATION}s（同一群会话，观察行锁串行化）==`);
  const conns = [];
  for (let i = 0; i < GROUP_SENDERS; i++) {
    conns.push(await connectWS(users['g' + i].token, 'g' + i));
  }
  console.log(`${conns.length} 个 WS 连接就绪`);

  const stopAt = Date.now() + GROUP_DURATION * 1000;
  const interval = 1000 / GROUP_RATE;
  const senders = conns.map((ws, i) => (async () => {
    while (Date.now() < stopAt) {
      sendWS(ws, {
        type: 'group_message', group_id: users.groupID,
        client_msg_id: msgId('g' + i), content: fakeCipher(),
      });
      await sleep(interval);
    }
  })());
  await Promise.all(senders);
  await sleep(2500); // 等待在途广播
  const total = conns.reduce((s, w) => s + w._stats.sent, 0);
  const e2e = conns.flatMap((w) => w._stats.e2e);
  const failed = conns.reduce((s, w) => s + w._stats.failed, 0);
  report(`群聊合计（发送者视角）`, { sent: total, failed, e2e });
  const throughput = (total / GROUP_DURATION).toFixed(1);
  console.log(`  目标吞吐 ${GROUP_SENDERS * GROUP_RATE}/s，实际提交 ${throughput}/s`);
  for (const ws of conns) ws.close();
}

async function phasePrivateHTTP() {
  console.log(`\n== 阶段二：私聊 HTTP，A→B，并发 ${PM_CONCURRENCY} × ${PM_DURATION}s（无服务端限流）==`);
  const a = users['a'], b = users['b'];
  const bws = await connectWS(b.token, 'b');
  const rtt = [], httpFail = [];
  let sent = 0;
  const stopAt = Date.now() + PM_DURATION * 1000;
  const worker = async (id) => {
    let backoff = 0;
    while (Date.now() < stopAt) {
      if (backoff > 0) { await sleep(backoff); backoff = 0; }
      const body = { to_uid: b.uid, content: fakeCipher(), client_msg_id: msgId('h' + id) };
      trackSend(body.client_msg_id);
      const r = await api('POST', '/api/private/send', a.token, body);
      if (r.ok) { rtt.push(r.rtt); sent++; }
      else {
        httpFail.push(r.json.message || JSON.stringify(r.json));
        backoff = 150; // 失败后退避，避免错误风暴
      }
    }
  };
  await Promise.all(Array.from({ length: PM_CONCURRENCY }, (_, i) => worker(i)));
  await sleep(2500);
  report(`私聊 HTTP 发送确认`, { sent, failed: httpFail.length, e2e: rtt });
  report(`私聊 HTTP → B 推送`, bws._stats);
  if (httpFail.length) console.log('  失败样例:', httpFail[0]);
  bws.close();
}

async function phasePrivateWS() {
  console.log(`\n== 阶段三：私聊 WS，A→B，共 ${PM_WS_COUNT} 条 @2/s（验证 ack/限流路径，上限 60 条/窗口）==`);
  const a = users['a'], b = users['b'];
  const aws = await connectWS(a.token, 'a');
  const bws = await connectWS(b.token, 'b');
  await sleep(300);
  for (let i = 0; i < PM_WS_COUNT; i++) {
    sendWS(aws, { type: 'private_message', to_uid: b.uid, client_msg_id: msgId('w' + i), content: fakeCipher() });
    await sleep(500);
  }
  await sleep(2500);
  report('私聊 WS（A 视角 ack）', aws._stats);
  report('私聊 WS → B 推送', bws._stats);
  aws.close(); bws.close();
}

try {
  await prepare();
  await phaseGroup();
  await sleep(1000);
  await phasePrivateHTTP();
  await sleep(1000);
  await phasePrivateWS();
  console.log('\n== 完成。压测账号前缀：' + PREFIX + '（如需清理按前缀删除）==');
} catch (e) {
  console.error('压测失败:', e.message);
  process.exit(1);
}
