# IM 系统 API 文档

## 一、接口说明

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| 基础 URL | `http://localhost:8080/api` |
| 认证方式 | Header 中携带 `Authorization: Bearer <token>` |
| 数据格式 | JSON |
| 编码格式 | UTF-8 |

### 1.2 通用响应结构

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 状态码，0=成功，其他=失败 |
| message | string | 提示信息 |
| data | object | 返回数据（可为 null） |

### 1.3 通用错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未授权 |
| 1003 | 权限不足 |
| 2001 | 用户不存在 |
| 2002 | 用户已存在 |
| 3001 | 群不存在 |
| 3002 | 群成员不存在 |
| 3003 | 已在群中 |
| 4001 | 好友不存在 |
| 4002 | 好友请求不存在 |
| 4003 | 已经是好友 |
| 5001 | 系统错误 |

---

## 二、认证接口（无需登录）

### 2.1 用户注册

**请求**

```
POST /api/user/register
Content-Type: application/json
```

**请求体**

```json
{
  "username": "string",    // 用户名，2-20字符，必填
  "password": "string",   // 密码，6-20字符，必填
  "nickname": "string"    // 昵称，2-20字符，必填
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "uid": "string",
    "username": "string",
    "nickname": "string",
    "avatar": "string"
  }
}
```

**示例**

```bash
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "zhangsan",
    "password": "123456",
    "nickname": "张三"
  }'
```

---

### 2.2 用户登录

**请求**

```
POST /api/user/login
Content-Type: application/json
```

**请求体**

```json
{
  "username": "string",  // 用户名，必填
  "password": "string"    // 密码，必填
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "uid": "string",
      "username": "string",
      "nickname": "string",
      "avatar": "string"
    }
  }
}
```

**示例**

```bash
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "zhangsan",
    "password": "123456"
  }'
```

---

## 三、好友接口（需登录）

### 3.1 搜索用户

**请求**

```
POST /api/friend/search
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**

```json
{
  "keyword": "string"  // 搜索关键词（用户名或昵称），必填
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "uid": "string",
      "username": "string",
      "nickname": "string",
      "avatar": "string",
      "status": 1
    }
  ]
}
```

| status | 说明 |
|--------|------|
| 1 | 不是好友 |
| 2 | 已是好友 |

---

### 3.2 添加好友

**请求**

```
POST /api/friend/add
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**

```json
{
  "friend_uid": "string",  // 好友UID，必填
  "remark": "string"       // 备注，可选
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 3.3 接受好友请求

**请求**

```
POST /api/friend/accept
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**

```json
{
  "friend_uid": "string"  // 申请人UID，必填
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 3.4 删除好友

**请求**

```
POST /api/friend/delete
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**

```json
{
  "friend_uid": "string"  // 好友UID，必填
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 3.5 获取好友列表

**请求**

```
GET /api/friend/list
Authorization: Bearer <token>
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "uid": "string",
      "friend_uid": "string",
      "username": "string",
      "nickname": "string",
      "avatar": "string",
      "remark": "string",
      "status": 2
    }
  ]
}
```

---

### 3.6 获取好友请求列表

**请求**

```
GET /api/friend/requests
Authorization: Bearer <token>
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "uid": "string",
      "friend_uid": "string",
      "username": "string",
      "nickname": "string",
      "avatar": "string",
      "remark": "string",
      "status": 1
    }
  ]
}
```

| status | 说明 |
|--------|------|
| 1 | 待接受 |
| 2 | 已是好友 |

---

## 四、群组接口（需登录）

### 4.1 创建群聊

**请求**

```
POST /api/group/create
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**

```json
{
  "name": "string",       // 群名称，2-50字符，必填
  "member_uids": ["string"]  // 成员UID数组，可选（不含创建者）
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "group_id": "string",
    "name": "string",
    "owner_uid": "string",
    "avatar": "string",
    "max_member": 200
  }
}
```

---

### 4.2 获取群列表

**请求**

```
GET /api/group/list
Authorization: Bearer <token>
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "group_id": "string",
      "name": "string",
      "owner_uid": "string",
      "avatar": "string",
      "max_member": 200
    }
  ]
}
```

---

### 4.3 获取群成员

**请求**

```
GET /api/group/members?group_id=xxx
Authorization: Bearer <token>
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "uid": "string",
      "username": "string",
      "nickname": "string",
      "avatar": "string",
      "role": 1
    }
  ]
}
```

| role | 说明 |
|------|------|
| 1 | 群主 |
| 2 | 管理员 |
| 3 | 普通成员 |

---

### 4.4 获取群消息历史

**请求**

```
GET /api/group/messages?group_id=xxx&limit=50
Authorization: Bearer <token>
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| limit | int | 否 | 返回数量，默认50 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "msg_id": "string",
      "group_id": "string",
      "from_uid": "string",
      "from_username": "string",
      "from_nickname": "string",
      "content": "string",
      "send_time": 1712304000000,
      "read_count": 5,
      "total_count": 10
    }
  ]
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| msg_id | string | 消息唯一ID |
| group_id | string | 群ID |
| from_uid | string | 发送者UID |
| from_username | string | 发送者用户名 |
| from_nickname | string | 发送者昵称 |
| content | string | 消息内容 |
| send_time | int64 | 发送时间（毫秒时间戳） |
| read_count | int64 | 已读人数 |
| total_count | int | 群成员总数 |

---

### 4.5 获取群未读消息

**请求**

```
GET /api/group/unread-messages?group_id=xxx&last_read_msg_id=xxx&limit=100
Authorization: Bearer <token>
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group_id | string | 是 | 群ID |
| last_read_msg_id | string | 否 | 最后已读消息ID，传入则只返回比该ID大的消息 |
| limit | int | 否 | 返回数量，默认50，最大100 |

**说明**

- 如果传入 `last_read_msg_id`，只返回比该ID大的消息（未读消息）
- 如果未读消息超过 limit，只返回最新的 limit 条
- 如果不传 `last_read_msg_id`，等同于 `/group/messages`

**响应**

同 4.4 获取群消息历史

---

## 五、私聊接口（需登录）

### 5.1 发送私信

**请求**

```
POST /api/private/send
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**

```json
{
  "to_uid": "string",   // 接收者UID，必填
  "content": "string"    // 消息内容，必填
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg_id": "string",
    "from_uid": "string",
    "to_uid": "string",
    "content": "string",
    "send_time": 1712304000000
  }
}
```

---

### 5.2 获取会话消息

**请求**

```
GET /api/private/conversation?friend_uid=xxx&limit=50&offset=0
Authorization: Bearer <token>
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| friend_uid | string | 是 | 好友UID |
| limit | int | 否 | 返回数量，默认50 |
| offset | int | 否 | 偏移量，默认0 |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "msg_id": "string",
      "from_uid": "string",
      "to_uid": "string",
      "content": "string",
      "type": 1,
      "send_time": 1712304000000,
      "read": true,
      "is_read": true
    }
  ]
}
```

---

### 5.3 获取所有未读消息数

**请求**

```
GET /api/private/unread-counts
Authorization: Bearer <token>
```

**说明**

返回当前用户与每个好友的未读消息数，用于上线时显示红点。

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "friend_uid_1": 3,
    "friend_uid_2": 5
  }
}
```

**示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user123": 2,
    "user456": 7
  }
}
```

---

### 5.4 标记消息已读

**请求**

```
POST /api/private/mark-read?from_uid=xxx
Authorization: Bearer <token>
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| from_uid | string | 是 | 发送者UID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 六、消息已读接口（需登录）

### 6.1 上报已读状态

**请求**

```
POST /api/message/report-read
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**

```json
{
  "room_id": "string",         // 房间ID（私聊为对方UID，群聊为群ID）
  "last_read_msg_id": "string" // 最后已读消息ID
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 6.2 获取未读消息数

**请求**

```
GET /api/message/unread-count?room_id=xxx
Authorization: Bearer <token>
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| room_id | string | 是 | 房间ID |
| last_msg_id | string | 否 | 最后一条消息ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "count": 5
  }
}
```

---

### 6.3 获取房间已读状态

**请求**

```
GET /api/message/read-status?room_id=xxx
Authorization: Bearer <token>
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| room_id | string | 是 | 房间ID |

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "last_read_msg_id": "string",
    "last_read_time": 1712304000000
  }
}
```

---

## 七、WebSocket 接口

### 7.1 连接地址

```
ws://localhost:8080/ws?room=global&token=<token>
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| room | string | 是 | 固定值 `global` |
| token | string | 是 | 登录获取的 JWT token |

**连接示例**

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?room=global&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...');
```

---

### 7.2 通用消息结构

**发送消息**

```json
{
  "type": "string"  // 消息类型，必填
}
```

**接收消息**

```json
{
  "type": "string"  // 消息类型
}
```

---

### 7.3 发送私聊消息

**发送**

```json
{
  "type": "private_message",
  "client_msg_id": "string",
  "to_uid": "string",
  "content": "string",
  "timestamp": 1712304000000
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 固定值 `private_message` |
| client_msg_id | string | 必填 | 客户端生成的消息ID（用于幂等），格式：`Date.now() + Math.random().toString(36).substr(2, 9)` |
| to_uid | string | 是 | 接收者UID |
| content | string | 是 | 消息内容 |
| timestamp | int64 | 否 | 客户端时间戳 |

**响应（发送者收到）**

```json
{
  "type": "private_message",
  "msg_id": "string",
  "from_uid": "string",
  "from_username": "string",
  "from_nickname": "string",
  "to_uid": "string",
  "content": "string",
  "send_time": 1712304000000
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| msg_id | string | 服务端生成的唯一消息ID |
| from_uid | string | 发送者UID |
| from_username | string | 发送者用户名 |
| from_nickname | string | 发送者昵称 |
| to_uid | string | 接收者UID |
| content | string | 消息内容 |
| send_time | int64 | 发送时间（毫秒时间戳） |

**示例**

```javascript
// 发送
ws.send(JSON.stringify({
  type: 'private_message',
  client_msg_id: Date.now() + Math.random().toString(36).substr(2, 9),
  to_uid: 'user123',
  content: '你好',
  timestamp: Date.now()
}));

// 接收（发送者自己也会收到）
ws.onmessage = function(event) {
  const msg = JSON.parse(event.data);
  if (msg.type === 'private_message') {
    console.log('收到私聊:', msg.content);
  }
};
```

---

### 7.4 发送群聊消息

**发送**

```json
{
  "type": "group_message",
  "client_msg_id": "string",
  "group_id": "string",
  "content": "string",
  "timestamp": 1712304000000
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 固定值 `group_message` |
| client_msg_id | string | 必填 | 客户端生成的消息ID（用于幂等） |
| group_id | string | 是 | 群ID |
| content | string | 是 | 消息内容 |
| timestamp | int64 | 否 | 客户端时间戳 |

**响应（发送者收到）**

```json
{
  "type": "group_message",
  "msg_id": "string",
  "group_id": "string",
  "from_uid": "string",
  "from_username": "string",
  "from_nickname": "string",
  "content": "string",
  "send_time": 1712304000000,
  "read_count": 0,
  "total_count": 10
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| msg_id | string | 服务端生成的唯一消息ID |
| group_id | string | 群ID |
| from_uid | string | 发送者UID |
| from_username | string | 发送者用户名 |
| from_nickname | string | 发送者昵称 |
| content | string | 消息内容 |
| send_time | int64 | 发送时间（毫秒时间戳） |
| read_count | int64 | 已读人数 |
| total_count | int | 群成员总数 |

**示例**

```javascript
// 发送
ws.send(JSON.stringify({
  type: 'group_message',
  client_msg_id: Date.now() + Math.random().toString(36).substr(2, 9),
  group_id: 'group123',
  content: '大家好',
  timestamp: Date.now()
}));
```

---

### 7.5 已读回执

#### 私聊已读回执

**发送**

```json
{
  "type": "read_receipt",
  "msg_id": "string",
  "room_id": "string"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 固定值 `read_receipt` |
| msg_id | string | 是 | 已读消息ID |
| room_id | string | 是 | 房间ID（私聊为对方UID） |

**示例**

```javascript
ws.send(JSON.stringify({
  type: 'read_receipt',
  msg_id: 'msg123',
  room_id: 'user456'
}));
```

---

#### 群聊已读回执

**发送**

```json
{
  "type": "group_read_receipt",
  "group_id": "string",
  "msg_ids": ["string"]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 固定值 `group_read_receipt` |
| group_id | string | 是 | 群ID |
| msg_ids | array | 是 | 已读消息ID数组 |

**响应（发送者收到已读更新）**

```json
{
  "type": "group_read_receipt",
  "msg_id": "string",
  "read_count": 5,
  "total_count": 10
}
```

**示例**

```javascript
// 发送群聊已读
ws.send(JSON.stringify({
  type: 'group_read_receipt',
  group_id: 'group123',
  msg_ids: ['msg1', 'msg2', 'msg3']
}));

// 同时给自己发送已读回执
ws.send(JSON.stringify({
  type: 'read_receipt',
  msg_id: 'msg3',
  room_id: 'group123'
}));
```

---

### 7.6 接收服务器推送

#### 私聊消息推送

```json
{
  "type": "private_message",
  "msg_id": "string",
  "from_uid": "string",
  "from_username": "string",
  "from_nickname": "string",
  "to_uid": "string",
  "content": "string",
  "send_time": 1712304000000
}
```

#### 群聊消息推送

```json
{
  "type": "group_message",
  "msg_id": "string",
  "group_id": "string",
  "from_uid": "string",
  "from_username": "string",
  "from_nickname": "string",
  "content": "string",
  "send_time": 1712304000000,
  "read_count": 0,
  "total_count": 10
}
```

#### 私聊已读回执推送

```json
{
  "type": "read_receipt",
  "room_id": "string",
  "reader_uid": "string",
  "msg_id": "string",
  "read_time": 1712304000000
}
```

#### 群聊已读状态更新

```json
{
  "type": "group_read_receipt",
  "msg_id": "string",
  "read_count": 5,
  "total_count": 10
}
```

#### 心跳响应

```json
{
  "type": "heartbeat_ack",
  "ts": 1712304000000
}
```

---

### 7.7 错误类型

| type | 说明 |
|------|------|
| error | 错误消息 |
| duplicate_message | 重复消息（幂等拦截） |

**错误消息格式**

```json
{
  "type": "error",
  "code": "string",
  "message": "string"
}
```

---

### 7.8 断线重连

WebSocket 断开后，前端应该：

1. 检测断线（`onclose` 或 `onerror` 事件）
2. 等待一小段时间（如 1-3 秒）
3. 使用相同的 token 重新连接
4. 重新订阅消息

```javascript
let reconnectTimer = null;

ws.onclose = function() {
  console.log('WebSocket 断开');
  // 延迟重连
  reconnectTimer = setTimeout(() => {
    reconnect();
  }, 3000);
};

function reconnect() {
  const ws = new WebSocket('ws://localhost:8080/ws?room=global&token=' + token);
  // 设置回调...
}
```

---

## 八、数据类型参考

### 8.1 用户信息

```typescript
interface User {
  uid: string;       // 用户唯一ID（雪花算法生成）
  username: string;  // 用户名
  nickname: string;  // 昵称
  avatar: string;    // 头像URL
}
```

### 8.2 好友信息

```typescript
interface Friend {
  uid: string;       // 当前用户UID
  friend_uid: string; // 好友UID
  username: string;  // 好友用户名
  nickname: string;  // 好友昵称
  avatar: string;    // 好友头像
  remark: string;    // 备注
  status: number;    // 状态：1=待接受，2=已是好友
}
```

### 8.3 群组信息

```typescript
interface Group {
  group_id: string;  // 群唯一ID
  name: string;      // 群名称
  owner_uid: string; // 群主UID
  avatar: string;   // 群头像
  max_member: number; // 最大成员数（默认200）
}
```

### 8.4 群成员信息

```typescript
interface GroupMember {
  uid: string;       // 成员UID
  username: string;  // 成员用户名
  nickname: string;  // 成员昵称
  avatar: string;    // 成员头像
  role: number;      // 角色：1=群主，2=管理员，3=普通成员
}
```

### 8.5 私聊消息

```typescript
interface PrivateMessage {
  msg_id: string;    // 消息唯一ID
  from_uid: string;  // 发送者UID
  to_uid: string;    // 接收者UID
  content: string;   // 消息内容
  type: number;      // 消息类型：1=私聊
  send_time: number; // 发送时间（毫秒时间戳）
  read: boolean;     // 是否已读
  is_read: boolean;  // 是否已读（兼容字段）
}
```

### 8.6 群聊消息

```typescript
interface GroupMessage {
  msg_id: string;        // 消息唯一ID
  group_id: string;      // 群ID
  from_uid: string;      // 发送者UID
  from_username: string; // 发送者用户名
  from_nickname: string; // 发送者昵称
  content: string;       // 消息内容
  send_time: number;      // 发送时间（毫秒时间戳）
  read_count: number;    // 已读人数
  total_count: number;   // 群成员总数
}
```

---

## 九、前端代码示例

### 9.1 WebSocket 连接和消息处理

```javascript
class IMClient {
  constructor(token) {
    this.token = token;
    this.ws = null;
    this reconnectTimer = null;
  }

  connect() {
    const wsUrl = `ws://localhost:8080/ws?room=global&token=${this.token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket 已连接');
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      this.handleMessage(msg);
    };

    this.ws.onclose = () => {
      console.log('WebSocket 断开');
      this.scheduleReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket 错误:', error);
    };
  }

  handleMessage(msg) {
    switch (msg.type) {
      case 'private_message':
        console.log('收到私聊:', msg.from_nickname, msg.content);
        break;
      case 'group_message':
        console.log('收到群消息:', msg.from_nickname, msg.content);
        break;
      case 'read_receipt':
        console.log('消息已读:', msg.msg_id);
        break;
      case 'group_read_receipt':
        console.log('群消息已读更新:', msg.msg_id, msg.read_count, '/', msg.total_count);
        break;
      case 'heartbeat_ack':
        // 心跳响应
        break;
      case 'error':
        console.error('错误:', msg.code, msg.message);
        break;
    }
  }

  sendPrivateMessage(toUid, content) {
    const clientMsgId = Date.now() + Math.random().toString(36).substr(2, 9);
    this.ws.send(JSON.stringify({
      type: 'private_message',
      client_msg_id: clientMsgId,
      to_uid: toUid,
      content: content,
      timestamp: Date.now()
    }));
    return clientMsgId;
  }

  sendGroupMessage(groupId, content) {
    const clientMsgId = Date.now() + Math.random().toString(36).substr(2, 9);
    this.ws.send(JSON.stringify({
      type: 'group_message',
      client_msg_id: clientMsgId,
      group_id: groupId,
      content: content,
      timestamp: Date.now()
    }));
    return clientMsgId;
  }

  sendReadReceipt(msgId, roomId) {
    this.ws.send(JSON.stringify({
      type: 'read_receipt',
      msg_id: msgId,
      room_id: roomId
    }));
  }

  sendGroupReadReceipt(groupId, msgIds) {
    // 发送群聊已读回执
    this.ws.send(JSON.stringify({
      type: 'group_read_receipt',
      group_id: groupId,
      msg_ids: msgIds
    }));

    // 同时给自己发送已读回执
    if (msgIds.length > 0) {
      this.ws.send(JSON.stringify({
        type: 'read_receipt',
        msg_id: msgIds[msgIds.length - 1],
        room_id: groupId
      }));
    }
  }

  startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({
          type: 'heartbeat',
          ts: Date.now()
        }));
      }
    }, 30000); // 每30秒发送一次心跳
  }

  scheduleReconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    this.reconnectTimer = setTimeout(() => {
      console.log('准备重连...');
      this.connect();
    }, 3000);
  }

  disconnect() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
    }
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    if (this.ws) {
      this.ws.close();
    }
  }
}
```

---

## 十、注意事项

1. **幂等性**：发送消息时请使用唯一的 `client_msg_id`，重试时使用相同的 ID 可以避免重复发送
2. **心跳**：建议每 30 秒发送一次心跳保持连接
3. **断线重连**：WebSocket 断开后会自动重连，请处理好消息顺序
4. **Token 过期**：如果 token 过期，需要重新登录获取新 token
5. **消息存储**：私聊消息通过 HTTP API 获取历史，群聊消息同理


源代码：
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>IM 简易客户端</title>
  <style>
    body { font-family: sans-serif; margin: 0; padding: 8px; background: #f0f0f0; }
    h1 { font-size: 1.2rem; }
    input, button, textarea { font-size: 14px; padding: 6px; margin: 2px; }
    .hidden { display: none !important; }
    .row { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
    .panel { background: #fff; border: 1px solid #ccc; padding: 10px; margin-bottom: 10px; }
    #chatPage { display: flex; flex-direction: column; height: calc(100vh - 24px); max-width: 1100px; margin: 0 auto; }
    #mainRow { display: flex; flex: 1; min-height: 0; gap: 8px; }
    #sidebar { width: 280px; min-width: 200px; border: 1px solid #ccc; background: #fff; display: flex; flex-direction: column; }
    #sidebar .tabs { display: flex; flex-wrap: wrap; border-bottom: 1px solid #ddd; }
    #sidebar .tabs button { flex: 1; min-width: 60px; padding: 8px; border: none; background: #eee; cursor: pointer; }
    #sidebar .tabs button.active { background: #ddd; font-weight: bold; }
    #listArea { flex: 1; overflow: auto; padding: 6px; font-size: 13px; }
    #chatArea { flex: 1; display: flex; flex-direction: column; min-width: 0; border: 1px solid #ccc; background: #fff; }
    #chatTitle { padding: 8px; border-bottom: 1px solid #ddd; font-weight: bold; }
    #chatMessages { flex: 1; overflow: auto; padding: 8px; background: #fafafa; font-size: 14px; }
    .msg { margin: 6px 0; padding: 6px; border-left: 3px solid #999; background: #fff; }
    .msg.self { border-left-color: #07c; text-align: right; }
    .msg .meta { font-size: 12px; color: #666; }
    #chatInputRow { display: flex; padding: 8px; border-top: 1px solid #ddd; gap: 6px; }
    #chatInputRow input { flex: 1; }
    .friend-item, .group-item { padding: 8px; border: 1px solid #eee; margin: 4px 0; cursor: pointer; position: relative; }
    .friend-item.active, .group-item.active { background: #e3f2fd; }
    .item-actions { float: right; }
    .item-actions button { font-size: 12px; padding: 2px 6px; }
    .unread-badge {
      position: absolute;
      right: 8px;
      top: 50%;
      transform: translateY(-50%);
      background: linear-gradient(135deg, #ff6b6b, #ff4d4f);
      color: white;
      font-size: 11px;
      padding: 2px 7px;
      border-radius: 10px;
      min-width: 18px;
      text-align: center;
      box-shadow: 0 2px 4px rgba(255, 77, 79, 0.3);
      animation: pulse 1.5s ease-in-out infinite;
    }
    @keyframes pulse {
      0%, 100% { transform: translateY(-50%) scale(1); }
      50% { transform: translateY(-50%) scale(1.1); }
    }
    .msg .read-status {
      font-size: 11px;
      color: #999;
      margin-top: 2px;
      display: flex;
      align-items: center;
      gap: 3px;
    }
    .msg .read-status.unread {
      color: #1890ff;
      font-weight: 500;
    }
    .msg .read-status::before {
      content: '';
      display: inline-block;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #d9d9d9;
    }
    .msg .read-status.unread::before {
      background: #1890ff;
      box-shadow: 0 0 6px rgba(24, 144, 255, 0.5);
    }
    #authPage { max-width: 400px; margin: 40px auto; }
    .modal { position: fixed; inset: 0; background: rgba(0,0,0,.4); display: none; align-items: center; justify-content: center; z-index: 100; }
    .modal.show { display: flex; }
    .modal-inner { background: #fff; padding: 16px; border: 1px solid #333; max-width: 90vw; max-height: 90vh; overflow: auto; min-width: 280px; }
    #apiHint { font-size: 12px; color: #666; margin-top: 8px; }
  </style>
  <script src="config.js"></script>
</head>
<body>

  <div id="authPage">
    <div class="panel">
      <h1>登录 / 注册</h1>
      <div class="row">
        <button type="button" id="tabLogin" onclick="showAuth('login')">登录</button>
        <button type="button" id="tabReg" onclick="showAuth('register')">注册</button>
      </div>
      <div id="loginForm">
        <p>用户名 <input id="loginUsername" autocomplete="username"></p>
        <p>密码 <input id="loginPassword" type="password" autocomplete="current-password"></p>
        <p><button type="button" onclick="login()">登录</button></p>
      </div>
      <div id="registerForm" class="hidden">
        <p>用户名 <input id="regUsername" autocomplete="username"></p>
        <p>昵称 <input id="regNickname"></p>
        <p>密码 <input id="regPassword" type="password" autocomplete="new-password"></p>
        <p><button type="button" onclick="register()">注册</button></p>
      </div>
      <p id="authMsg" style="color:red;"></p>
      <p id="apiHint"></p>
    </div>
  </div>

  <div id="chatPage" class="hidden">
    <div class="panel row" style="justify-content: space-between;">
      <span id="meInfo">未登录</span>
      <button type="button" onclick="logout()">退出</button>
    </div>
    <div id="mainRow">
      <div id="sidebar">
        <div class="tabs">
          <button type="button" class="active" data-tab="friends" onclick="switchSide('friends')">好友</button>
          <button type="button" data-tab="requests" onclick="switchSide('requests')">请求</button>
          <button type="button" data-tab="groups" onclick="switchSide('groups')">群聊</button>
        </div>
        <div class="panel" style="margin:0;border-left:0;border-right:0;">
          <div class="row">
            <input id="searchInput" placeholder="搜索用户 keyword" style="flex:1" onkeydown="if(event.key==='Enter')searchUser()">
            <button type="button" onclick="searchUser()">搜</button>
          </div>
          <button type="button" onclick="openAddFriendModal()" style="margin-top:6px;width:100%">添加好友（弹窗搜索）</button>
        </div>
        <div id="listArea"></div>
      </div>
      <div id="chatArea">
        <div id="chatTitle">请选择左侧好友或群聊</div>
        <div id="chatMessages"></div>
        <div id="chatInputRow">
          <input id="messageInput" placeholder="消息..." onkeydown="if(event.key==='Enter')sendMessage()">
          <button type="button" onclick="sendMessage()">发送</button>
        </div>
      </div>
    </div>
  </div>

  <div id="addFriendModal" class="modal">
    <div class="modal-inner">
      <h3>添加好友</h3>
      <p><input id="addFriendKeyword" placeholder="用户名/昵称 keyword" style="width:100%"></p>
      <p><button type="button" onclick="searchUserForAdd()">搜索</button></p>
      <div id="searchResults"></div>
      <p><button type="button" onclick="closeAddFriendModal()">关闭</button></p>
    </div>
  </div>

  <div id="createGroupModal" class="modal">
    <div class="modal-inner">
      <h3>创建群聊</h3>
      <p>群名称 <input id="createGroupName" maxlength="50" style="width:100%"></p>
      <p>勾选好友为成员：</p>
      <div id="createGroupMemberList"></div>
      <p class="row">
        <button type="button" onclick="closeCreateGroupModal()">取消</button>
        <button type="button" onclick="submitCreateGroup()">创建</button>
      </p>
    </div>
  </div>

  <script>
    (function () {
      var cfg = window.IM_CONFIG || {};
      var hint = document.getElementById('apiHint');
      if (hint) {
        if (cfg.API_ORIGIN) {
          hint.textContent = 'API: ' + cfg.API_ORIGIN + '（独立前端模式）';
        } else {
          hint.textContent = 'API: 同域 ' + location.origin + '（可在 config.js 设置 API_ORIGIN）';
        }
      }
    })();

    var API = (function () {
      var cfg = window.IM_CONFIG || {};
      if (cfg.API_ORIGIN) {
        return String(cfg.API_ORIGIN).replace(/\/$/, '') + '/api';
      }
      return location.origin + '/api';
    })();

    function wsRoomURL(tok) {
      var cfg = window.IM_CONFIG || {};
      var origin;
      if (cfg.API_ORIGIN) {
        try {
          var u = new URL(cfg.API_ORIGIN);
          origin = (u.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + u.host;
        } catch (e) {
          origin = '';
        }
      }
      if (!origin) {
        var proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
        origin = proto + '//' + location.host;
      }
      return origin + '/room?room=default&token=' + encodeURIComponent(tok);
    }

    let token = null;
    let currentUser = null;
    let friendsCache = [];
    let groupsCache = [];
    let currentChatFriend = null;
    let currentChatGroup = null;
    let ws = null;
    let roomReadStatus = {}; // 记录每个房间的已读状态
    let unreadCounts = {}; // 记录每个房间的未读数

    function showAuth(which) {
      document.getElementById('loginForm').classList.toggle('hidden', which !== 'login');
      document.getElementById('registerForm').classList.toggle('hidden', which !== 'register');
      document.getElementById('tabLogin').classList.toggle('active', which === 'login');
      document.getElementById('tabReg').classList.toggle('active', which === 'register');
    }

    function setAuthMsg(t) {
      document.getElementById('authMsg').textContent = t || '';
    }

    function parseLoginPayload(data) {
      if (!data || typeof data !== 'object') return null;
      if (data.token && data.user) return { token: data.token, user: data.user };
      return null;
    }

    async function register() {
      setAuthMsg('');
      const username = document.getElementById('regUsername').value.trim();
      const nickname = document.getElementById('regNickname').value.trim();
      const password = document.getElementById('regPassword').value;
      try {
        const res = await fetch(API + '/user/register', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, nickname, password })
        });
        const j = await res.json();
        if (j.code === 0) {
          const p = parseLoginPayload(j.data);
          if (!p) {
            setAuthMsg('注册成功，请登录');
            showAuth('login');
            return;
          }
          applySession(p.token, p.user);
          enterChat();
          return;
        }
        setAuthMsg(j.message || '注册失败');
      } catch (e) {
        setAuthMsg('网络错误');
      }
    }

    async function login() {
      setAuthMsg('');
      const username = document.getElementById('loginUsername').value.trim();
      const password = document.getElementById('loginPassword').value;
      try {
        const res = await fetch(API + '/user/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password })
        });
        const j = await res.json();
        if (j.code === 0) {
          const p = parseLoginPayload(j.data);
          if (!p) {
            setAuthMsg('响应格式错误');
            return;
          }
          applySession(p.token, p.user);
          enterChat();
          return;
        }
        setAuthMsg(j.message || '登录失败');
      } catch (e) {
        setAuthMsg('网络错误');
      }
    }

    function applySession(t, user) {
      token = t;
      currentUser = user;
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
    }

    function logout() {
      token = null;
      currentUser = null;
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      if (ws) { try { ws.close(); } catch (e) {} }
      ws = null;
      document.getElementById('chatPage').classList.add('hidden');
      document.getElementById('authPage').classList.remove('hidden');
    }

    function enterChat() {
      document.getElementById('authPage').classList.add('hidden');
      document.getElementById('chatPage').classList.remove('hidden');
      document.getElementById('meInfo').textContent =
        (currentUser.nickname || currentUser.username) + ' | uid: ' + currentUser.uid;
      switchSide('friends');
      loadFriends();
      loadFriendRequests();
      loadGroups();
      connectWS();
    }

    function authHeader() {
      return { 'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json' };
    }

    function switchSide(tab) {
      document.querySelectorAll('#sidebar .tabs button[data-tab]').forEach(b => {
        b.classList.toggle('active', b.getAttribute('data-tab') === tab);
      });
      window._sideTab = tab;
      if (tab === 'friends') loadFriends();
      if (tab === 'requests') loadFriendRequests();
      if (tab === 'groups') loadGroups();
    }

    async function loadFriends() {
      const el = document.getElementById('listArea');
      if (window._sideTab !== 'friends') return;
      try {
        const res = await fetch(API + '/friend/list', { headers: authHeader() });
        const j = await res.json();
        if (j.code !== 0) {
          el.textContent = j.message || '加载失败';
          return;
        }
        friendsCache = j.data || [];
        el.innerHTML = friendsCache.length ? friendsCache.map(f => {
          const unread = unreadCounts[f.friend_uid] || 0;
          const unreadBadge = unread > 0 ? `<span class="unread-badge">${unread > 99 ? '99+' : unread}</span>` : '';
          return `
          <div class="friend-item ${currentChatFriend && currentChatFriend.friend_uid === f.friend_uid ? 'active' : ''}" data-fid="${f.friend_uid}" onclick="selectFriend('${f.friend_uid}')">
            ${esc(f.nickname || f.username)} (${esc(f.username)})
            <span class="item-actions">
              <button type="button" onclick="event.stopPropagation();deleteFriend('${f.friend_uid}')">删</button>
            </span>
            ${unreadBadge}
          </div>
        `}).join('') : '<div>暂无好友</div>';
      } catch (e) {
        el.textContent = '网络错误';
      }
    }

    async function deleteFriend(friendUid) {
      if (!confirm('删除好友 ' + friendUid + ' ?')) return;
      try {
        const res = await fetch(API + '/friend/delete', {
          method: 'POST',
          headers: authHeader(),
          body: JSON.stringify({ friend_uid: friendUid })
        });
        const j = await res.json();
        if (j.code === 0) {
          if (currentChatFriend && currentChatFriend.friend_uid === friendUid) {
            currentChatFriend = null;
            document.getElementById('chatTitle').textContent = '请选择左侧好友或群聊';
            document.getElementById('chatMessages').innerHTML = '';
          }
          loadFriends();
        } else alert(j.message || '删除失败');
      } catch (e) {
        alert('网络错误');
      }
    }

    async function loadFriendRequests() {
      const el = document.getElementById('listArea');
      if (window._sideTab !== 'requests') return;
      try {
        const res = await fetch(API + '/friend/requests', { headers: authHeader() });
        const j = await res.json();
        if (j.code !== 0) {
          el.textContent = j.message || '加载失败';
          return;
        }
        const list = j.data || [];
        el.innerHTML = list.length ? list.map(r => `
          <div class="friend-item">
            ${esc(r.nickname || r.username)} — 请求加你为友
            <button type="button" onclick="acceptFriend('${r.uid}')">接受</button>
          </div>
        `).join('') : '<div>暂无请求</div>';
      } catch (e) {
        el.textContent = '网络错误';
      }
    }

    async function acceptFriend(requesterUid) {
      try {
        const res = await fetch(API + '/friend/accept', {
          method: 'POST',
          headers: authHeader(),
          body: JSON.stringify({ friend_uid: requesterUid })
        });
        const j = await res.json();
        if (j.code === 0) {
          loadFriendRequests();
          loadFriends();
        } else alert(j.message || '失败');
      } catch (e) {
        alert('网络错误');
      }
    }

    async function searchUser() {
      const keyword = document.getElementById('searchInput').value.trim();
      if (!keyword) return;
      await searchUserApi(keyword, false);
    }

    async function searchUserForAdd() {
      const keyword = document.getElementById('addFriendKeyword').value.trim();
      if (!keyword) return;
      await searchUserApi(keyword, true);
    }

    async function searchUserApi(keyword, inModal) {
      try {
        const res = await fetch(API + '/friend/search', {
          method: 'POST',
          headers: authHeader(),
          body: JSON.stringify({ keyword: keyword })
        });
        const j = await res.json();
        if (j.code !== 0) {
          if (inModal) document.getElementById('searchResults').textContent = j.message || '搜索失败';
          else document.getElementById('listArea').innerHTML = '<div>' + (j.message || '搜索失败') + '</div>';
          return;
        }
        const users = j.data || [];
        const rows = users.filter(u => u.uid !== currentUser.uid).map(u => `
          <div class="row" style="justify-content:space-between;margin:4px 0;">
            <span>${esc(u.nickname)} @${esc(u.username)}</span>
            <button type="button" onclick="addFriend('${u.uid}')">添加</button>
          </div>
        `).join('');
        if (inModal) {
          document.getElementById('searchResults').innerHTML = rows || '<div>无结果</div>';
        } else {
          document.getElementById('listArea').innerHTML = '<div><b>搜索结果</b></div>' + (rows || '<div>无结果</div>');
        }
      } catch (e) {
        if (inModal) document.getElementById('searchResults').textContent = '网络错误';
      }
    }

    async function addFriend(friendUid) {
      try {
        const res = await fetch(API + '/friend/add', {
          method: 'POST',
          headers: authHeader(),
          body: JSON.stringify({ friend_uid: friendUid })
        });
        const j = await res.json();
        alert(j.code === 0 ? '已发送请求' : (j.message || '失败'));
      } catch (e) {
        alert('网络错误');
      }
    }

    function openAddFriendModal() {
      document.getElementById('addFriendModal').classList.add('show');
      document.getElementById('searchResults').innerHTML = '';
    }
    function closeAddFriendModal() {
      document.getElementById('addFriendModal').classList.remove('show');
    }

    function reportRead(roomId, lastMsgId, fromUid) {
      if (!ws || ws.readyState !== WebSocket.OPEN || !lastMsgId) return;
      ws.send(JSON.stringify({
        type: 'read_receipt',
        msg_id: lastMsgId,
        room_id: roomId,
        from_uid: fromUid
      }));
      unreadCounts[roomId] = 0;
      loadFriends();
    }

    async function getRoomReadStatus(roomId) {
      try {
        const res = await fetch(API + '/message/read-status?room_id=' + encodeURIComponent(roomId), {
          headers: authHeader()
        });
        const j = await res.json();
        if (j.code === 0 && j.data) {
          roomReadStatus[roomId] = j.data;
          return j.data;
        }
      } catch (e) {
        console.error('获取已读状态失败', e);
      }
      return null;
    }

    function selectFriend(friendUid) {
      currentChatGroup = null;
      const f = friendsCache.find(x => x.friend_uid === friendUid);
      const name = f ? (f.nickname || f.username) : friendUid;
      currentChatFriend = { friend_uid: friendUid, nickname: name };
      document.getElementById('chatTitle').textContent = '私聊: ' + name;
      document.querySelectorAll('.friend-item').forEach(el => {
        el.classList.toggle('active', el.getAttribute('data-fid') === friendUid);
      });
      document.querySelectorAll('.group-item').forEach(el => el.classList.remove('active'));
      loadPrivateHistory(friendUid);
    }

    async function loadPrivateHistory(friendUid) {
      const box = document.getElementById('chatMessages');
      try {
        const [readStatus, res] = await Promise.all([
          getRoomReadStatus(friendUid),
          fetch(API + '/private/conversation?friend_uid=' + encodeURIComponent(friendUid), {
            headers: { 'Authorization': 'Bearer ' + token }
          })
        ]);
        const j = await res.json();
        if (j.code !== 0) {
          box.textContent = j.message || '加载失败';
          return;
        }
        const msgs = j.data || [];
        
        if (msgs.length > 0) {
          const lastMsg = msgs[msgs.length - 1];
          reportRead(friendUid, lastMsg.msg_id);
        }

        box.innerHTML = msgs.map(m => {
          const isSelf = m.from_uid === currentUser.uid;
          const who = isSelf ? '我' : (currentChatFriend.nickname || '对方');
          const isRead = !readStatus || !readStatus.last_read_msg_id || m.msg_id <= readStatus.last_read_msg_id;
          return renderMsgLine(isSelf, m.content, m.send_time, who, m.is_read || isRead, m.msg_id);
        }).join('');
        box.scrollTop = box.scrollHeight;
      } catch (e) {
        box.textContent = '网络错误';
      }
    }

    async function loadGroups() {
      const el = document.getElementById('listArea');
      if (window._sideTab !== 'groups') return;
      try {
        const res = await fetch(API + '/group/list', { headers: { 'Authorization': 'Bearer ' + token } });
        const j = await res.json();
        if (j.code !== 0) {
          el.textContent = j.message || '加载失败';
          return;
        }
        groupsCache = j.data || [];
        const rows = groupsCache.map(g => {
          const unread = unreadCounts[g.group_id] || 0;
          const unreadBadge = unread > 0 ? `<span class="unread-badge">${unread > 99 ? '99+' : unread}</span>` : '';
          return `
          <div class="group-item ${currentChatGroup && currentChatGroup.group_id === g.group_id ? 'active' : ''}" data-gid="${g.group_id}" onclick="selectGroup('${g.group_id}')">
            ${esc(g.name || g.group_id)}
            <button type="button" onclick="event.stopPropagation();viewGroupMembers('${g.group_id}')">成员</button>
            ${unreadBadge}
          </div>
        `}).join('');
        el.innerHTML = '<div style="margin-bottom:8px;"><button type="button" onclick="openCreateGroupModal()">+ 创建群</button></div>' +
          (rows || '<div>暂无群</div>');
      } catch (e) {
        el.textContent = '网络错误';
      }
    }

    async function viewGroupMembers(groupId) {
      try {
        const res = await fetch(API + '/group/members?group_id=' + encodeURIComponent(groupId), {
          headers: { 'Authorization': 'Bearer ' + token }
        });
        const j = await res.json();
        if (j.code !== 0) {
          alert(j.message || '失败');
          return;
        }
        const list = (j.data || []).map(m => m.nickname + '(' + m.username + ')').join('\n');
        alert('群成员:\n' + list);
      } catch (e) {
        alert('网络错误');
      }
    }

    function selectGroup(groupId) {
      currentChatFriend = null;
      const g = groupsCache.find(x => x.group_id === groupId);
      const name = g && g.name ? g.name : groupId;
      currentChatGroup = { group_id: groupId, name: name };
      document.getElementById('chatTitle').textContent = '群聊: ' + name + ' | 成员按钮可看列表';
      document.querySelectorAll('.group-item').forEach(el => {
        el.classList.toggle('active', el.getAttribute('data-gid') === groupId);
      });
      document.querySelectorAll('.friend-item').forEach(el => el.classList.remove('active'));
      loadGroupHistory(groupId);
    }

    async function loadGroupHistory(groupId) {
      const box = document.getElementById('chatMessages');
      try {
        const res = await fetch(API + '/group/messages?group_id=' + encodeURIComponent(groupId) + '&limit=50', {
          headers: { 'Authorization': 'Bearer ' + token }
        });
        const j = await res.json();
        if (j.code !== 0) {
          box.textContent = j.message || '加载失败';
          return;
        }
        const msgs = j.data || [];
        
        if (msgs.length > 0) {
          const unreadMsgIds = msgs.filter(m => m.from_uid !== currentUser.uid).map(m => m.msg_id);
          if (unreadMsgIds.length > 0) {
            reportGroupRead(groupId, unreadMsgIds);
          }
        }

        box.innerHTML = msgs.map(m => {
          const self = m.from_uid === currentUser.uid;
          const who = self ? '我' : (m.from_nickname || m.from_username || m.from_uid);
          return renderMsgLine(self, m.content, m.send_time, who, false, m.msg_id, m.read_count, m.total_count);
        }).join('');
        box.scrollTop = box.scrollHeight;
      } catch (e) {
        box.textContent = '网络错误';
      }
    }

    function reportGroupRead(groupId, msgIds) {
      if (!ws || ws.readyState !== WebSocket.OPEN || msgIds.length === 0) return;
      ws.send(JSON.stringify({
        type: 'group_read_receipt',
        group_id: groupId,
        msg_ids: msgIds
      }));
      ws.send(JSON.stringify({
        type: 'read_receipt',
        msg_id: msgIds[msgIds.length - 1],
        room_id: groupId
      }));
      unreadCounts[groupId] = 0;
      loadGroups();
    }

    function openCreateGroupModal() {
      const box = document.getElementById('createGroupMemberList');
      document.getElementById('createGroupName').value = '';
      if (!friendsCache.length) {
        box.innerHTML = '<div>无好友，无法选人</div>';
      } else {
        box.innerHTML = friendsCache.map(f => `
          <label class="row"><input type="checkbox" class="cg-m" value="${f.friend_uid}"> ${esc(f.nickname || f.username)}</label>
        `).join('');
      }
      document.getElementById('createGroupModal').classList.add('show');
    }
    function closeCreateGroupModal() {
      document.getElementById('createGroupModal').classList.remove('show');
    }

    async function submitCreateGroup() {
      const name = document.getElementById('createGroupName').value.trim();
      if (name.length < 2) {
        alert('群名至少2字');
        return;
      }
      const member_uids = Array.from(document.querySelectorAll('.cg-m:checked')).map(c => c.value);
      try {
        const res = await fetch(API + '/group/create', {
          method: 'POST',
          headers: authHeader(),
          body: JSON.stringify({ name: name, member_uids: member_uids })
        });
        const j = await res.json();
        if (j.code === 0) {
          closeCreateGroupModal();
          switchSide('groups');
          loadGroups();
        } else alert(j.message || '失败');
      } catch (e) {
        alert('网络错误');
      }
    }

    function renderMsgLine(self, content, sendTime, whoLabel, isRead, msgId, readCount, totalCount) {
      const t = typeof sendTime === 'number' ? sendTime : parseInt(sendTime, 10);
      const d = new Date(t);
      const timeStr = isNaN(d.getTime()) ? String(sendTime) : d.toLocaleString();
      let readStatusHtml = '';
      if (self) {
        if (readCount !== undefined && totalCount !== undefined) {
          readStatusHtml = `<div class="read-status">${readCount}人已读 / ${totalCount}人</div>`;
        } else {
          readStatusHtml = `<div class="read-status ${isRead ? '' : 'unread'}">${isRead ? '已读' : '未读'}</div>`;
        }
      }
      return `<div class="msg ${self ? 'self' : ''}" data-msg-id="${esc(msgId || '')}"><div class="meta">${esc(whoLabel)} · ${esc(timeStr)}</div><div>${esc(content)}</div>${readStatusHtml}</div>`;
    }

    function esc(s) {
      const d = document.createElement('div');
      d.textContent = s == null ? '' : String(s);
      return d.innerHTML;
    }

    let heartbeatTimer = null;

    function startHeartbeat() {
      stopHeartbeat();
      heartbeatTimer = setInterval(function() {
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({type: 'heartbeat', ts: Date.now()}));
        }
      }, 20000);
    }

    function stopHeartbeat() {
      if (heartbeatTimer) {
        clearInterval(heartbeatTimer);
        heartbeatTimer = null;
      }
    }

    let wsReconnectTimer = null;
    let wsConnecting = false;

    function connectWS() {
      if (!token) return;
      if (wsConnecting) return;
      stopHeartbeat();
      if (wsReconnectTimer) {
        clearTimeout(wsReconnectTimer);
        wsReconnectTimer = null;
      }
      if (ws) {
        ws.onclose = null;
        ws.close();
        ws = null;
      }
      wsConnecting = true;
      const url = wsRoomURL(token);
      ws = new WebSocket(url);
      ws.onmessage = function(ev) {
        String(ev.data).split('\n').filter(Boolean).forEach(function(line) {
          try {
            onWsMessage(JSON.parse(line));
          } catch (e) {}
        });
      };
      ws.onopen = function() {
        wsConnecting = false;
        startHeartbeat();
      };
      ws.onclose = function() {
        wsConnecting = false;
        stopHeartbeat();
        wsReconnectTimer = setTimeout(connectWS, 3000);
      };
      ws.onerror = function(err) {
        wsConnecting = false;
        console.error('WebSocket error:', err);
      };
    }

    function onWsMessage(msg) {
      if (msg.type === 'read_receipt') {
        updateMessageReadStatus(msg.msg_id, true);
        return;
      }
      if (msg.type === 'group_read_receipt') {
        updateGroupMessageReadStatus(msg.msg_id, msg.read_count, msg.total_count);
        return;
      }
      if (msg.type === 'private_message') {
        const isSelf = msg.from_uid === currentUser.uid;
        if (isSelf) {
          if (currentChatFriend && currentChatFriend.friend_uid === msg.to_uid) {
            appendChatLine(true, msg.content, msg.send_time, '我', false, msg.msg_id);
          }
          return;
        }
        if (currentChatFriend && currentChatFriend.friend_uid === msg.from_uid) {
          appendChatLine(false, msg.content, msg.send_time, msg.from_nickname || msg.from_username || msg.from_uid, true, msg.msg_id);
          reportRead(msg.from_uid, msg.msg_id, msg.from_uid);
        } else {
          unreadCounts[msg.from_uid] = (unreadCounts[msg.from_uid] || 0) + 1;
          loadFriends();
          alert('新私聊: ' + (msg.from_nickname || msg.from_username));
        }
        return;
      }
      if (msg.type === 'group_message') {
        const self = msg.from_uid === currentUser.uid;
        if (self) {
          if (currentChatGroup && currentChatGroup.group_id === msg.group_id) {
            appendChatLine(true, msg.content, msg.send_time, '我', false, msg.msg_id, 1, msg.total_count || 0);
            reportGroupRead(msg.group_id, [msg.msg_id]);
          }
          return;
        }
        if (currentChatGroup && currentChatGroup.group_id === msg.group_id) {
          const who = msg.from_nickname || msg.from_username || msg.from_uid;
          appendChatLine(false, msg.content, msg.send_time, who, true, msg.msg_id);
          reportGroupRead(msg.group_id, [msg.msg_id]);
        } else {
          unreadCounts[msg.group_id] = (unreadCounts[msg.group_id] || 0) + 1;
          alert('群消息: ' + (msg.group_id || ''));
        }
      }
    }

    function updateMessageReadStatus(msgId, isRead) {
      const allMsgs = document.querySelectorAll('.msg[data-msg-id]');
      Array.from(allMsgs).forEach((msgEl, i) => {
        const currentMsgId = msgEl.getAttribute('data-msg-id');
        if (currentMsgId && currentMsgId <= msgId) {
          const readStatusEl = msgEl.querySelector('.read-status');
          if (readStatusEl) {
            msgEl.style.transition = 'all 0.3s ease';
            msgEl.style.backgroundColor = '#f0f9ff';
            setTimeout(() => {
              msgEl.style.backgroundColor = '';
            }, 500 + i * 50);
            readStatusEl.textContent = isRead ? '已读' : '未读';
            readStatusEl.classList.toggle('unread', !isRead);
          }
        }
      });
    }

    function updateGroupMessageReadStatus(msgId, readCount, totalCount) {
      const allMsgs = document.querySelectorAll('.msg[data-msg-id]');
      for (let i = 0; i < allMsgs.length; i++) {
        const msgEl = allMsgs[i];
        const currentMsgId = msgEl.getAttribute('data-msg-id');
        if (currentMsgId === msgId) {
          const readStatusEl = msgEl.querySelector('.read-status');
          if (readStatusEl) {
            msgEl.style.transition = 'all 0.3s ease';
            msgEl.style.backgroundColor = '#f0f9ff';
            setTimeout(() => {
              msgEl.style.backgroundColor = '';
            }, 500);
            readStatusEl.textContent = `${readCount}人已读 / ${totalCount}人`;
            readStatusEl.classList.remove('unread');
          }
          break;
        }
      }
    }

    function appendChatLine(self, content, sendTime, whoLabel, isRead, msgId, readCount, totalCount) {
      const box = document.getElementById('chatMessages');
      box.insertAdjacentHTML('beforeend', renderMsgLine(self, content, sendTime, whoLabel, isRead, msgId, readCount, totalCount));
      box.scrollTop = box.scrollHeight;
    }

    function sendMessage() {
      if (!currentChatFriend && !currentChatGroup) {
        alert('先选好友或群');
        return;
      }
      const input = document.getElementById('messageInput');
      const content = input.value.trim();
      if (!content) return;
      if (!ws || ws.readyState !== WebSocket.OPEN) {
        alert('WebSocket 未连接，请稍后重试');
        return;
      }
      if (currentChatGroup) {
        ws.send(JSON.stringify({
          type: 'group_message',
          client_msg_id: Date.now() + Math.random().toString(36).substr(2, 9),
          group_id: currentChatGroup.group_id,
          content: content,
          timestamp: Date.now()
        }));
        input.value = '';
        return;
      }
      ws.send(JSON.stringify({
        type: 'private_message',
        client_msg_id: Date.now() + Math.random().toString(36).substr(2, 9),
        to_uid: currentChatFriend.friend_uid,
        content: content,
        timestamp: Date.now()
      }));
      input.value = '';
    }

    window.onload = function() {
      const t = localStorage.getItem('token');
      const u = localStorage.getItem('user');
      if (t && u) {
        try {
          token = t;
          currentUser = JSON.parse(u);
          enterChat();
        } catch (e) {
          logout();
        }
      }
    };
  </script>
</body>
</html>
