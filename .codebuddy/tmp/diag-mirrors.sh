#!/bin/bash
echo "=== MIRROR REACHABILITY (fast, 4s timeout each) ==="
for m in docker.mirrors.ustc.edu.cn hub-mirror.c.163.com docker.m.daocloud.io mirror.baidubce.com dockerproxy.com mirror.iscas.ac.cn registry-1.docker.io; do
  code=$(curl -s -o /dev/null -w '%{http_code}' --connect-timeout 4 --max-time 6 https://$m/v2/ 2>/dev/null)
  echo "$m -> ${code:-FAIL}"
done
echo ""
echo "=== DOCKER DAEMON CONFIG ==="
cat /etc/docker/daemon.json 2>/dev/null | head -30
