// 极简响应式状态管理：createStore + 订阅

let uid = 0;

export function createStore(initialState) {
  let state = initialState || {};
  const listeners = new Set();

  function get() {
    return state;
  }

  function set(partial) {
    const next = typeof partial === 'function' ? partial(state) : partial;
    state = Object.assign({}, state, next);
    listeners.forEach(fn => fn(state));
    return state;
  }

  function subscribe(fn) {
    listeners.add(fn);
    return () => listeners.delete(fn);
  }

  return { get, set, subscribe, _id: ++uid };
}
