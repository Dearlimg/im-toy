// 前端运行时配置示例
//
// 使用方式：复制为 config.js，按环境修改 API_ORIGIN。
//
// 容器部署（Docker Compose / nginx 代理）：
//   API_ORIGIN 留空字符串 ""，前后端同域，HTTP/WS 自动走当前 host。
//
// 本地开发（前后端分离，Go 直连）：
//   API_ORIGIN: 'http://127.0.0.1:8080'
window.IM_CONFIG = {
  API_ORIGIN: ''
};
