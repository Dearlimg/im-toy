#!/bin/sh
set -e

# 替换 config.js 中的 API 地址
if [ -n "$BACKEND_URL" ]; then
    sed -i "s|http://backend:8080|$BACKEND_URL|g" /usr/share/nginx/html/config.js
fi

exec "$@"