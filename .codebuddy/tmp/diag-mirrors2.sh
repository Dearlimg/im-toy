#!/bin/bash
echo "=== SECOND ROUND MIRROR TEST ==="
for m in 0vmzj3q6.mirror.aliyuncs.com vlgh0kqj.mirror.aliyuncs.com huecker.io dockerhub.timeweb.cloud noohub.ru docker.1panel.live docker.1ms.run docker.xuanyuan.me dockerpull.org; do
  code=$(curl -s -o /dev/null -w '%{http_code}' --connect-timeout 4 --max-time 6 https://$m/v2/ 2>/dev/null)
  echo "$m -> ${code:-FAIL}"
done
