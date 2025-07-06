# User Online Status - Data Models

## Redis Data Structure

### Key Pattern
```
user:status:{user_id}
```

### Value Structure
```
"online" | "away" | "offline" | (nil for unknown)
```

### TTL (Time To Live)
```
online:  30 seconds
away:    10 minutes  
offline: 24 hours
unknown: no key (expired)
```

## Data Operations

### 1. Set User Online
```redis
SET user:status:123 "online" EX 30
```
- **Key**: `user:status:123`
- **Value**: `"online"`
- **TTL**: 30 seconds
- **Action**: User becomes online, auto-expire after 30s if no heartbeat

### 2. Set User Away
```redis
SET user:status:123 "away" EX 600
```
- **Key**: `user:status:123`
- **Value**: `"away"`
- **TTL**: 10 minutes (600 seconds)
- **Action**: User idle but still logged in

### 3. Set User Offline
```redis
SET user:status:123 "offline" EX 86400
```
- **Key**: `user:status:123`
- **Value**: `"offline"`
- **TTL**: 24 hours (longer TTL for offline status)
- **Action**: Explicitly set user offline

### 4. Get User Status
```redis
GET user:status:123
```
- **Returns**: `"online"` | `"away"` | `"offline"` | `nil`
- **Logic**: 
  - If key exists and value = "online" → User is online
  - If key exists and value = "away" → User is away (idle)
  - If key exists and value = "offline" → User is offline
  - If key doesn't exist → User status is unknown (expired)

### 5. Get Multiple Users Status
```redis
MGET user:status:123 user:status:456 user:status:789
```
- **Returns**: Array of values `["online", "away", "offline", nil]`
- **Logic**: Same as single get, but for multiple users

### 6. Refresh Status (Heartbeat)
```redis
EXPIRE user:status:123 30
```
- **Action**: Reset TTL to 30 seconds and transition away→online
- **Condition**: Works for both "online" and "away" status

## Redis Configuration

### Basic Setup
```conf
# Enable keyspace notifications for expired keys (optional)
notify-keyspace-events Ex

# Memory optimization
maxmemory-policy allkeys-lru
```

### Connection Settings
```go
redis.Options{
    Addr:         "redis:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 3,
    MaxRetries:   3,
}
```

## Data Flow Examples

### User Goes Online
```
1. Client calls SetUserStatus(user_id: "123", status: ONLINE)
2. Server executes: SET user:status:123 "online" EX 30
3. Redis stores: user:status:123 = "online" (expires in 30s)
```

### User Sends Heartbeat
```
1. Client calls SendHeartbeat(user_id: "123")
2. Server checks: GET user:status:123
3. If value = "online": EXPIRE user:status:123 30
4. TTL reset to 30 seconds
```

### User Goes Offline (Explicit)
```
1. Client calls SetUserStatus(user_id: "123", status: OFFLINE)
2. Server executes: SET user:status:123 "offline" EX 86400
3. Redis stores: user:status:123 = "offline" (expires in 24h)
```

### User Goes Offline (Auto)
```
1. No heartbeat received for 30 seconds
2. Redis key user:status:123 expires automatically
3. GET user:status:123 returns nil
4. Server interprets nil as "offline"
```

### Check User Status
```
1. Client calls GetUserStatus(user_id: "123")
2. Server executes: GET user:status:123
3. Result interpretation:
   - "online" → User is online
   - "offline" → User is offline
   - nil → User is offline (key expired)
```

## Performance Considerations

### Redis Optimization
- Use connection pooling
- Implement pipelining for bulk operations
- Monitor memory usage
- Use appropriate TTL values

### Scaling
- Consider Redis Cluster for high availability
- Implement read replicas for read-heavy workloads
- Use Redis Sentinel for automatic failover

## Error Handling

### Common Redis Errors
- **Connection timeout**: Retry with exponential backoff
- **Memory full**: Implement proper eviction policy
- **Key not found**: Treat as offline status
- **Network issues**: Use circuit breaker pattern 