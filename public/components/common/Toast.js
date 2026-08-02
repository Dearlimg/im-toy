// 通用组件：轻提示 Toast（自动消失）

let container = null;
let current = null;
let hideTimer = null;

function ensureContainer() {
  if (!container) {
    container = document.createElement('div');
    container.id = 'toast';
    document.body.appendChild(container);
  }
  return container;
}

export function Toast(text, ms) {
  const c = ensureContainer();
  c.textContent = text;
  c.style.opacity = '0';
  c.style.transition = 'opacity 0.2s, transform 0.2s';
  c.style.transform = 'translateX(-50%) translateY(-10px)';
  // 强制重排以触发过渡
  void c.offsetHeight;
  c.style.opacity = '1';
  c.style.transform = 'translateX(-50%) translateY(0)';

  if (hideTimer) clearTimeout(hideTimer);
  hideTimer = setTimeout(() => {
    c.style.opacity = '0';
    c.style.transform = 'translateX(-50%) translateY(-10px)';
  }, ms || 2000);
}
