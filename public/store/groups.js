// 群组状态：列表、成员、当前群

import { createStore } from './index.js';

export const groupsStore = createStore({
  groups: [],          // [{ id, name, owner_id, created_at }]
  members: [],         // 当前群成员 [{ uid, username, nickname }]
  current: null        // 当前群 { id, name, owner_id }
});

export function setGroups(list) {
  groupsStore.set({ groups: list || [] });
}

export function setMembers(list) {
  groupsStore.set({ members: list || [] });
}

export function setCurrentGroup(group) {
  groupsStore.set({ current: group });
}

export function findGroup(id) {
  return groupsStore.get().groups.find(g => String(g.id) === String(id)) || null;
}
