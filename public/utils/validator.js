// 表单校验工具（纯函数，无外部依赖）

// 校验用户名：3-32 位字母数字下划线
export function validUsername(u) {
  return typeof u === 'string' && /^[a-zA-Z0-9_]{3,32}$/.test(u);
}

// 校验昵称：非空且不超过 32 字
export function validNickname(n) {
  return typeof n === 'string' && n.trim().length >= 1 && n.trim().length <= 32;
}

// 校验密码：6-64 位
export function validPassword(p) {
  return typeof p === 'string' && p.length >= 6 && p.length <= 64;
}

// 校验群名：2-50 字
export function validGroupName(n) {
  return typeof n === 'string' && n.trim().length >= 2 && n.trim().length <= 50;
}

// 校验消息非空
export function validMessage(m) {
  return typeof m === 'string' && m.trim().length > 0;
}
