#!/bin/bash
set -e
COMPOSE=/root/workspace/docker-compose.yaml

echo "=== 1. Backup ==="
cp $COMPOSE ${COMPOSE}.bak.$(date +%s)
echo "backup done"

echo ""
echo "=== 2. Stop & remove old bitnami kafka container ==="
docker rm -f kafka 2>/dev/null || true

echo ""
echo "=== 3. Update compose file ==="
# Write new compose with apache/kafka replacing bitnami
cat > $COMPOSE <<'EOF'
version: '3'

#  这个是写给软协go组的基础软件的docker-compose,把这个复制到服务器的一个熟悉的文件夹,然后执行:docker compose up -d (前提是安装了docker)
#  端口映射,文件挂载,时区都搞了,如果配置redis配置文件的时候有问题,到时候你们会自己修改的
#  需要密码的都是: sta_go
#  如果kafka执行失败,大概率是因为读写权限的问题,执行这个:
#  chmod -R 777 components

services:
  redis:
    image: redis:latest
    ports:
      - 6379:6379
    container_name: redis
    volumes:
      - ./components/redis/data:/data
      - ./components/redis/config/redis.conf:/usr/local/redis/config/redis.conf
    environment:
      TZ: Asia/Shanghai
    restart: always
    command: redis-server --requirepass sta_go --appendonly yes

  mysql:
    image: mysql:latest
    ports:
      - 3307:3306
    container_name: mysql
    volumes:
      - ./components/mysql/data:/var/lib/mysql
      - /etc/localtime:/etc/localtime
    environment:
      MYSQL_ROOT_PASSWORD: sta_go
      TZ: Asia/Shanghai
    restart: always

  # Kafka 4.x KRaft 模式(不需要 Zookeeper)
  kafka:
    image: apache/kafka:latest
    container_name: kafka
    ports:
      - 9092:9092
    volumes:
      - ./components/kafka/data:/tmp/kafka-logs
      - /etc/localtime:/etc/localtime
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://121.40.235.227:9092
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
      KAFKA_HEAP_OPTS: "-Xmx256M -Xms128M"
    restart: always

  # pgvector 向量数据库 PG16 + pgvector0.8+
  pgvector:
    image: pgvector/pgvector:pg16
    container_name: pgvector
    ports:
      - "5432:5432"
    volumes:
      - ./components/pgvector-data:/var/lib/postgresql/data
      - /etc/localtime:/etc/localtime
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: sta_go
      POSTGRES_DB: rag_db
      TZ: Asia/Shanghai
    shm_size: 4g
    restart: always
EOF

echo "compose updated"
cat $COMPOSE | grep -A4 'kafka:'
