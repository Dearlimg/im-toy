# 前端目录架构（public/）

原生 ES Modules 分层架构，无构建工具，浏览器直接加载。

## 一、目录结构

```
public/
├── index.html            # 页面骨架 + 3 个 CSS + config.js + 入口模块
├── config.js             # 运行时配置（部署时由环境注入，不入库）
├── config.example.js     # 配置模板/说明
├── assets/               # 静态资源层：图片、图标、字体等
├── styles/               # 样式资源层
│   ├── base.css          # 变量 / reset / 基础元素
│   ├── layout.css        # 页面骨架布局（双栏、响应式）
│   └── components.css    # 组件样式（气泡、按钮、弹窗、输入区）
├── pages/                # 页面层
│   └── chat.js           # 主聊天页编排入口
├── components/           # 组件层
│   ├── common/           # 通用组件（与业务无关，可复用）
│   │   ├── Avatar.js
│   │   ├── Toast.js
│   │   ├── Modal.js
│   │   ├── EmojiPicker.js
│   │   └── MessageBubble.js
│   └── business/         # 业务组件（依赖具体页面/数据）
│       ├── AuthForm.js   # 登录/注册表单
│       ├── Sidebar.js    # 左侧列表（好友/请求/群）
│       ├── ChatWindow.js # 聊天消息区 + 标题栏
│       └── ChatInput.js  # 输入区（表情/图片/发送）
├── store/                # 状态管理层（自研 createStore 订阅机制）
│   ├── index.js          # createStore(initialState) => {get,set,subscribe}
│   ├── auth.js           # 登录态 + session（localStorage）
│   ├── friends.js        # 好友/好友请求/当前好友
│   ├── groups.js         # 群列表/成员/当前群
│   ├── conversations.js  # 当前会话消息（含乐观消息去重）
│   └── unread.js         # 未读数（key: friend:<uid> / group:<gid>）
├── services/             # 服务层（网络 IO，唯一触碰后端的地方）
│   ├── http.js           # fetch 封装 + Bearer token + {code,message,data}
│   ├── api.js            # 业务 API 定义（注册/登录/好友/群/消息/未读）
│   └── ws.js             # WebSocket 客户端（心跳/重连/事件分发）
└── utils/                # 工具函数层（纯函数，无副作用）
    ├── format.js         # esc/时间格式化/预览/头像字符等
    ├── validator.js      # 用户名/昵称/密码/群名/消息校验
    └── dom.js            # $id / $ / $$ / el 快捷选择
```

## 二、各层职责与依赖方向

依赖只能自上而下，禁止反向引用：

```
pages（编排）
  ├──> business components（业务组件）
  │     └──> common components（通用组件）
  ├──> store（状态管理）
  ├──> services（网络）
  └──> utils（工具）
```

| 层 | 职责 | 允许依赖 |
|----|------|----------|
| `pages/` | 初始化、事件注册、会话切换、WS 事件分发 | 全部 |
| `components/business/` | 页面局部业务逻辑（列表、聊天窗口、输入） | common / store / services / utils |
| `components/common/` | 无业务语义的复用组件 | store 或 utils（避免 services） |
| `store/` | 全局状态 + 变更通知 | 仅自身（`index.js`） |
| `services/` | 与后端交互的唯一入口 | store（取 token） |
| `utils/` | 纯函数 | 无 |

约束约定：
- `components/common/` 不依赖 `components/business/` 与 `pages/`。
- `store/` 之间允许按需互相引用（如 `unread` 独立，`auth` 被 services 引用），但不应引入 DOM。
- `utils/` 必须是纯函数，不得触发渲染或网络。

## 三、数据流方向

单向数据流：`事件 → 页面/组件 → store（set）→ 订阅者 → 重渲染`

```
                    ┌──────────────┐
 用户操作/WS消息 ──> │ pages/chat.js │
                    └──────┬───────┘
            ┌──────────────┼──────────────────┐
            ▼              ▼                  ▼
    services(HTTP/WS)   store.set()      business 组件
        │                （状态变更）         │
        └── 返回数据 ────>│                  │
                         ▼                  ▼
                  store.subscribe()   直接更新 DOM（局部）
                         │
                         ▼
                  重渲染（Sidebar / ChatWindow）
```

关键路径：

1. **发送消息**：`ChatInput 回车 → chat.js handleSend → ws.send(私聊/群消息) → pushMessage 乐观上屏 → renderAll`
2. **接收消息**：`ws.js 解析 → emit('private'|'group') → chat.js → 当前会话 pushMessage/renderAll，否则 increment 未读数 → unreadStore 订阅 → document.title + Sidebar 角标`
3. **会话切换**：`Sidebar 点击 → chat.js selectFriend/selectGroup → openConversation + clear 未读 + loadHistory(services) → conversationsStore.set → renderAll`
4. **登录**：`AuthForm → api.login → saveSession(authStore) → chat.js enterApp → connect(ws) + loadAll + loadUnread`
5. **已读回执**：`收到私聊 → chat.js send({type:'read_receipt'}) → 对方 ws 收到 → updateMessageRead 更新气泡状态`

## 四、前后端契约要点

- HTTP 统一响应 `{code:0, message, data}`，非 0 抛 `ApiError`。
- 好友字段用 `friend_uid`，群字段用 `group_id`，好友请求/搜索结果用 `uid`。
- WS 地址：`ws://<host>/room?room=default&token=<token>`，心跳 20s，断线 3s 自动重连。
- 未读接口返回 `{ friend_uid: count }`。
