// 前端全局配置示例
//
// 使用方式：复制本文件为 config.js，按实际环境修改 API_ORIGIN 即可。
// config.js 通常由部署脚本/环境注入生成，不入库。
//
// API_ORIGIN 同时用于 HTTP API 和 WebSocket：
//   - HTTP:  <API_ORIGIN>/api/...
//   - WS:    ws(s)://<host>/room?room=default&token=...
window.IM_CONFIG = {
  // 本地开发：
  // API_ORIGIN: 'http://127.0.0.1:8080'
  // 容器/生产环境（前后端同域时留空字符串即可，走 location.host）：
  API_ORIGIN: 'http://backend:8080'
};
