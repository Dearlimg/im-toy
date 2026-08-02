// 通用组件：表情选择面板（常用 emoji 网格）

const EMOJIS = [
  '😀', '😁', '😂', '🤣', '😊', '😍', '😘', '😜',
  '🤪', '😎', '🤓', '🥳', '😭', '😅', '😳', '🥺',
  '😡', '😱', '🤔', '🙄', '😴', '🤗', '😇', '🙃',
  '👍', '👎', '👏', '🙏', '💪', '🤝', '✌️', '👌',
  '❤️', '💔', '💯', '🔥', '✨', '🎉', '🎂', '🌹',
  '😋', '🫡', '😷', '🤯', '💤', '🌟', '⚡', '🍀'
];

export function createEmojiPanel(onPick) {
  const panel = document.createElement('div');
  panel.id = 'emojiPanel';
  panel.className = 'emoji-panel';
  EMOJIS.forEach(e => {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.textContent = e;
    btn.addEventListener('click', () => onPick && onPick(e));
    panel.appendChild(btn);
  });
  return panel;
}

// 预定义常用表情（供业务组件直接使用）
export const EMOJI_LIST = EMOJIS;
