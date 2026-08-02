#!/bin/bash
echo "=== 1. TEST MIRRORS REACHABILITY ==="
for m in docker.mirrors.ustc.edu.cn hub-mirror.c.163.com docker.m.daocloud.io mirror.baidubce.com dockerproxy.com mirror.iscas.ac.cn huecker.io dockerhub.timeweb.cloud noohub.ru registry-1.docker.io; do
  code=$(curl -sI -o /dev/null -w '%{http_code}' --connect-timeout 5 https://$m/v2/ 2>/dev/null)
  echo "$m -> ${code:-FAIL}"
done

echo ""
echo "=== 2. GENERAL INTERNET CHECK ==="
curl -sI -o /dev/null -w 'baidu.com: %{http_code} (%{time_total}s)\n' --connect-timeout 5 https://www.baidu.com 2>&1
curl -sI -o /dev/null -w 'google.com: %{http_code} (%{time_total}s)\n' --connect-timeout 5 https://www.google.com 2>&1

echo ""
echo "=== 3. DOCKER PULL TEST (apache/kafka) ==="
timeout 90 docker pull apache/kafka:latest 2>&1 | tail -20

echo ""
echo "=== 4. DOCKER PULL TEST (small image, direct from hub) ==="
timeout 60 docker pull hello-world:latest 2>&1 | tail -10
