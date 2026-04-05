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

### 5.3 获取未读消息数

**请求**

```
GET /api/private/unread
Authorization: Bearer <token>
```

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
