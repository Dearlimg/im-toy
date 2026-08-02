#!/bin/bash
set -e

echo "=== 1. Backup old daemon.json ==="
cp /etc/docker/daemon.json /etc/docker/daemon.json.bak.$(date +%s) 2>/dev/null || true

echo "=== 2. Write new daemon.json (only working mirrors) ==="
cat > /etc/docker/daemon.json <<'EOF'
{
    "registry-mirrors": [
        "https://0vmzj3q6.mirror.aliyuncs.com",
        "https://vlgh0kqj.mirror.aliyuncs.com",
        "https://docker.1panel.live",
        "https://docker.1ms.run",
        "https://docker.xuanyuan.me",
        "https://docker.m.daocloud.io"
    ]
}
EOF
cat /etc/docker/daemon.json

echo ""
echo "=== 3. Restart docker ==="
systemctl restart docker
sleep 5
docker info 2>/dev/null | grep -A8 'Registry Mirrors'

echo ""
echo "=== 4. Verify: pull hello-world (small test) ==="
timeout 90 docker pull hello-world:latest 2>&1 | tail -8
