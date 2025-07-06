# Redis Debug and Troubleshooting Guide

## Overview

This document provides comprehensive guidance on debugging and troubleshooting Redis in the social app project, including viewing logs, checking connections, and useful commands.

## 1. View Logs

### Docker Compose Logs
```bash
# View all logs
make logs

# View Redis service logs
docker-compose logs redis

# View logs in real-time
docker-compose logs -f redis

# View app logs
docker-compose logs app
docker-compose logs -f app

# View logs with timestamp
docker-compose logs -t redis
```

### Redis Server Logs
```bash
# Connect to Redis container
docker-compose exec redis bash

# View Redis log file (if exists)
tail -f /var/log/redis/redis-server.log

# Or view logs via Redis CLI
redis-cli
> CONFIG GET logfile
> CONFIG GET loglevel
```

## 2. Redis Connection Debug

### Check Redis Connection
```bash
# Test connection from host
redis-cli -h localhost -p 6379 ping

# Test connection from app container
docker-compose exec app redis-cli -h redis -p 6379 ping

# Check Redis info
redis-cli -h localhost -p 6379 info
```

### Environment Variables Check
```bash
# Check env vars in app container
docker-compose exec app env | grep REDIS

# Expected values:
# REDIS_HOST=redis
# REDIS_PORT=6379
# REDIS_PASSWORD=
# REDIS_DB=0
```

## 3. Redis Commands Debug

### Basic Commands
```bash
# Connect to Redis
redis-cli -h localhost -p 6379

# List all keys
KEYS *

# List keys with pattern
KEYS user:status:*

# Get key value
GET user:status:123

# Get key with TTL
TTL user:status:123

# Get all info about key
TYPE user:status:123
TTL user:status:123
GET user:status:123
```

### User Status Debug Commands
```bash
# Set test user status
SET user:status:test_user "online"
EXPIRE user:status:test_user 30

# Check status
GET user:status:test_user
TTL user:status:test_user

# Monitor Redis commands real-time
MONITOR

# Clear all data (CAREFUL!)
FLUSHDB
```

## 4. Application Debug

### HTTP API Testing
```bash
# Test set user status
curl -X POST http://localhost:8080/api/v1/users/123/status \
  -H "Content-Type: application/json" \
  -d '{"status": "online"}'

# Test get user status
curl http://localhost:8080/api/v1/users/123/status

# Test heartbeat
curl -X POST http://localhost:8080/api/v1/users/123/heartbeat

# Test set away
curl -X PUT http://localhost:8080/api/v1/users/123/status/away

# Test bulk get
curl "http://localhost:8080/api/v1/users/status?user_ids=123,456,789"
```

### Go Application Logs
```bash
# View app logs with filter
docker-compose logs app | grep "ERROR"
docker-compose logs app | grep "Redis"
docker-compose logs app | grep "status"

# View logs real-time with filter
docker-compose logs -f app | grep -i "error\|redis\|status"
```

## 5. Common Issues & Solutions

### Issue 1: Redis Connection Failed
```
Error: dial tcp: lookup redis on X.X.X.X:53: no such host
```
**Solution:**
- Check docker-compose network
- Verify Redis service name in docker-compose.yml
- Check REDIS_HOST environment variable

### Issue 2: Redis Auth Failed
```
Error: NOAUTH Authentication required
```
**Solution:**
- Check REDIS_PASSWORD environment variable
- Check redis.conf requirepass setting

### Issue 3: Key Not Found
```
Error: redis: nil (key not found)
```
**Debug Steps:**
```bash
# Check if key exists
redis-cli EXISTS user:status:123

# Check TTL
redis-cli TTL user:status:123

# List all keys
redis-cli KEYS user:status:*
```

### Issue 4: TTL Issues
```bash
# Check current TTL
TTL user:status:123

# Set TTL manually for testing
EXPIRE user:status:123 30

# Check TTL again
TTL user:status:123
```

## 6. Debug Scripts

### Create Debug Script
```bash
# Tạo file debug script
cat > debug_redis.sh << 'EOF'
#!/bin/bash
echo "=== Redis Debug Script ==="
echo "1. Checking Redis connection..."
redis-cli -h localhost -p 6379 ping

echo "2. Checking Redis info..."
redis-cli -h localhost -p 6379 info server | head -10

echo "3. Checking user status keys..."
redis-cli -h localhost -p 6379 KEYS "user:status:*"

echo "4. Checking Redis memory usage..."
redis-cli -h localhost -p 6379 info memory | grep used_memory_human

echo "5. Checking connected clients..."
redis-cli -h localhost -p 6379 info clients | grep connected_clients
EOF

chmod +x debug_redis.sh
```

### Run Debug Script
```bash
./debug_redis.sh
```

## 7. Performance Monitoring

### Redis Performance Commands
```bash
# Monitor Redis performance
redis-cli --latency -h localhost -p 6379

# Monitor memory usage
redis-cli info memory | grep used_memory

# Monitor connected clients
redis-cli info clients

# Monitor keyspace
redis-cli info keyspace
```

### Application Performance
```bash
# Monitor HTTP requests
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/v1/users/123/status

# Create curl-format.txt
cat > curl-format.txt << 'EOF'
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF
```

