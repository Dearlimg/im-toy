// 业务组件：侧边栏（好友 / 请求 / 群聊列表 + 未读角标）
// 依赖：store、services、Avatar、Toast、Modal

import { Avatar } from '../common/Avatar.js';
import { Toast } from '../common/Toast.js';
import { openModal } from '../common/Modal.js';
import { friendsStore } from '../../store/friends.js';
import { groupsStore } from '../../store/groups.js';
import { getCount } from '../../store/unread.js';
import { listTime, preview, unreadLabel, esc } from '../../utils/format.js';
import {
  friendList, friendRequests, friendAccept, friendDelete, groupList, groupCreate
} from '../../services/api.js';

const listArea = document.getElementById('listArea');
let currentTab = 'friends';

let onSelectFriend = null;
let onSelectGroup = null;

export function setHandlers({ selectFriend, selectGroup }) {
  onSelectFriend = selectFriend;
  onSelectGroup = selectGroup;
}

export function switchSide(tab) {
  currentTab = tab;
  document.querySelectorAll('.side-tabs button').forEach(b =>
    b.classList.toggle('active', b.dataset.tab === tab));
  render();
}

export async function loadAll() {
  try {
    const [f, r, g] = await Promise.all([friendList(), friendRequests(), groupList()]);
    friendsStore.set({ friends: f || [], requests: r || [] });
    groupsStore.set({ groups: g || [] });
    render();
  } catch (e) {
    Toast('加载数据失败: ' + e.message);
  }
}

export function render() {
  listArea.innerHTML = '';
  if (currentTab === 'friends') renderFriends();
  else if (currentTab === 'requests') renderRequests();
  else renderGroups();
}

function renderFriends() {
  const friends = friendsStore.get().friends;
  if (!friends.length) {
    listArea.innerHTML = '<div class="empty-tip">暂无好友<br>点击右上角放大镜添加好友</div>';
    return;
  }
  const sorted = friends.slice().sort((a, b) =>
    (getCount('friend', b.friend_uid) || 0) - (getCount('friend', a.friend_uid) || 0));
  sorted.forEach(f => {
    const item = document.createElement('div');
    item.className = 'conv-item';
    const cur = friendsStore.get().current;
    if (cur && cur.kind === 'friend' && String(cur.friend_uid) === String(f.friend_uid)) {
      item.classList.add('active');
    }

    item.appendChild(Avatar(f.nickname || f.username, { online: true, size: 44, fontSize: 17 }));

    const info = document.createElement('div');
    info.className = 'conv-info';
    info.innerHTML = '<div class="conv-name">' + esc(f.nickname || f.username) + '</div>';
    item.appendChild(info);

    const right = document.createElement('div');
    right.className = 'conv-right';
    const cnt = getCount('friend', f.friend_uid);
    if (cnt) {
      const badge = document.createElement('div');
      badge.className = 'unread-badge';
      badge.textContent = unreadLabel(cnt);
      right.appendChild(badge);
    }
    item.appendChild(right);

    item.addEventListener('click', () => onSelectFriend && onSelectFriend(f));
    listArea.appendChild(item);
  });
}

function renderRequests() {
  const requests = friendsStore.get().requests;
  if (!requests.length) {
    listArea.innerHTML = '<div class="empty-tip">暂无好友请求</div>';
    return;
  }
  requests.forEach(r => {
    const item = document.createElement('div');
    item.className = 'conv-item';
    item.appendChild(Avatar(r.nickname || r.username));

    const info = document.createElement('div');
    info.className = 'conv-info';
    info.innerHTML = '<div class="conv-name">' + esc(r.nickname || r.username) + '</div>'
      + '<div class="conv-preview">请求加好友</div>';
    item.appendChild(info);

    const right = document.createElement('div');
    right.className = 'conv-right';
    const btn = document.createElement('button');
    btn.className = 'mini-btn';
    btn.textContent = '接受';
    btn.addEventListener('click', async (e) => {
      e.stopPropagation();
      try {
        await friendAccept(r.uid);
        Toast('已接受');
        await loadAll();
      } catch (err) {
        Toast(err.message);
      }
    });
    right.appendChild(btn);
    item.appendChild(right);

    item.addEventListener('click', () => openRequestDetail(r));
    listArea.appendChild(item);
  });
}

function renderGroups() {
  const createBtn = document.createElement('button');
  createBtn.className = 'create-group-btn';
  createBtn.textContent = '+ 创建群聊';
  createBtn.addEventListener('click', openCreateGroup);
  listArea.appendChild(createBtn);

  const groups = groupsStore.get().groups;
  if (!groups.length) {
    const tip = document.createElement('div');
    tip.className = 'empty-tip';
    tip.textContent = '暂无群聊\n点击上方按钮创建';
    listArea.appendChild(tip);
    return;
  }
  groups.forEach(g => {
    const item = document.createElement('div');
    item.className = 'conv-item';
    const cur = groupsStore.get().current;
    if (cur && cur.kind === 'group' && String(cur.group_id) === String(g.group_id)) {
      item.classList.add('active');
    }

    item.appendChild(Avatar(g.name, { size: 44, fontSize: 17 }));

    const info = document.createElement('div');
    info.className = 'conv-info';
    info.innerHTML = '<div class="conv-name">' + esc(g.name) + '</div>';
    item.appendChild(info);

    const right = document.createElement('div');
    right.className = 'conv-right';
    const cnt = getCount('group', g.group_id);
    if (cnt) {
      const badge = document.createElement('div');
      badge.className = 'unread-badge';
      badge.textContent = unreadLabel(cnt);
      right.appendChild(badge);
    }
    item.appendChild(right);

    item.addEventListener('click', () => onSelectGroup && onSelectGroup(g));
    listArea.appendChild(item);
  });
}

function openCreateGroup() {
  const friends = friendsStore.get().friends;
  if (!friends.length) {
    Toast('请先添加好友再建群');
    return;
  }
  const checked = new Set();
  const body = () => {
    const wrap = document.createElement('div');
    const input = document.createElement('input');
    input.id = 'groupNameInput';
    input.placeholder = '群名称（2-50 字）';
    input.style.cssText = 'width:100%;padding:10px 12px;border:1px solid #e3e3e3;border-radius:8px;font-size:14px;margin-bottom:12px;';
    wrap.appendChild(input);

    const list = document.createElement('div');
    list.style.maxHeight = '200px';
    list.style.overflowY = 'auto';
    friends.forEach(f => {
      const label = document.createElement('label');
      label.className = 'member-check';
      const cb = document.createElement('input');
      cb.type = 'checkbox';
      cb.value = f.friend_uid;
      cb.addEventListener('change', () => {
        if (cb.checked) checked.add(f.uid);
        else checked.delete(f.uid);
      });
      label.appendChild(cb);
      label.appendChild(Avatar(f.nickname || f.username, { size: 28, fontSize: 12 }));
      const span = document.createElement('span');
      span.textContent = f.nickname || f.username;
      label.appendChild(span);
      list.appendChild(label);
    });
    wrap.appendChild(list);
    return wrap;
  };

  openModal('创建群聊', body, [
    { text: '取消', onClick: () => false },
    {
      text: '创建', primary: true,
      onClick: async () => {
        const name = document.getElementById('groupNameInput').value.trim();
        if (!name || name.length < 2 || name.length > 50) {
          Toast('群名称需 2-50 字');
          return false;
        }
        try {
          await groupCreate(name, Array.from(checked));
          Toast('创建成功');
          await loadAll();
          switchSide('groups');
        } catch (e) {
          Toast(e.message);
          return false;
        }
      }
    }
  ]);
}

function openRequestDetail(r) {
  openModal('好友请求', () => {
    const d = document.createElement('div');
    d.style.cssText = 'text-align:center;color:#666;font-size:14px;line-height:2;';
    d.innerHTML = '用户 <b>' + esc(r.nickname || r.username) + '</b> 请求添加你为好友';
    return d;
  }, [
    { text: '关闭', onClick: () => false },
    {
      text: '接受', primary: true,
      onClick: async () => {
        try {
          await friendAccept(r.uid);
          Toast('已接受');
          await loadAll();
        } catch (e) {
          Toast(e.message);
          return false;
        }
      }
    }
  ]);
}
