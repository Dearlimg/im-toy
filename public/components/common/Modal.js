// 通用组件：模态框（内容由调用方注入，返回关闭函数）

export function openModal(title, buildBody, actions) {
  const mask = document.createElement('div');
  mask.className = 'modal-mask show';

  const modal = document.createElement('div');
  modal.className = 'modal';

  const h = document.createElement('h3');
  h.textContent = title;
  modal.appendChild(h);

  if (buildBody) {
    modal.appendChild(buildBody());
  }

  if (actions && actions.length) {
    const row = document.createElement('div');
    row.className = 'modal-actions';
    actions.forEach(a => {
      const btn = document.createElement('button');
      btn.textContent = a.text;
      btn.className = a.primary ? 'btn-primary' : 'btn-cancel';
      btn.style.margin = '0';
      btn.style.width = 'auto';
      btn.addEventListener('click', () => {
        const result = a.onClick ? a.onClick() : null;
        if (result !== false) close();
      });
      row.appendChild(btn);
    });
    modal.appendChild(row);
  }

  mask.appendChild(modal);
  document.body.appendChild(mask);

  mask.addEventListener('click', (e) => {
    if (e.target === mask) close();
  });

  function close() {
    mask.remove();
  }

  return { mask, modal, close };
}
