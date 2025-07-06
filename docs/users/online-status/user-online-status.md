# User Online Status Feature

## Overview
This feature allows the system to track and determine whether users are currently online or offline. It provides real-time status updates and can be used for various social features like showing active users, message delivery status, etc.

## Technical Design

### Architecture
- **Redis**: Used as the primary storage for online status with TTL (Time To Live)
- **gRPC**: For real-time status updates and API calls
- **Protocol Buffers**: For message serialization

### Data Structure
```
Key: "user:status:{user_id}"
Value: "online" | "offline"
TTL: 30 seconds (configurable)
```

## gRPC Methods

### 1. SetUserStatus
```protobuf
rpc SetUserStatus(SetUserStatusRequest) returns (SetUserStatusResponse);
```

**Request:**
```protobuf
message SetUserStatusRequest {
  string user_id = 1;
  UserStatus status = 2;  // ONLINE or OFFLINE
}
```

**Response:**
```protobuf
message SetUserStatusResponse {
  bool success = 1;
  string message = 2;
  UserStatusData data = 3;
}
```

### 2. GetUserStatus
```protobuf
rpc GetUserStatus(GetUserStatusRequest) returns (GetUserStatusResponse);
```

**Request:**
```protobuf
message GetUserStatusRequest {
  string user_id = 1;
}
```

**Response:**
```protobuf
message GetUserStatusResponse {
  bool success = 1;
  UserStatusData data = 2;
}
```

### 3. GetMultipleUserStatus
```protobuf
rpc GetMultipleUserStatus(GetMultipleUserStatusRequest) returns (GetMultipleUserStatusResponse);
```

**Request:**
```protobuf
message GetMultipleUserStatusRequest {
  repeated string user_ids = 1;
}
```

**Response:**
```protobuf
message GetMultipleUserStatusResponse {
  bool success = 1;
  repeated UserStatusData data = 2;
}
```

## gRPC Streaming

### Subscribe to Status Updates
```protobuf
rpc SubscribeStatusUpdates(SubscribeStatusUpdatesRequest) returns (stream StatusUpdateEvent);
```

**Request:**
```protobuf
message SubscribeStatusUpdatesRequest {
  string user_id = 1;
  repeated string watch_user_ids = 2;  // Users to watch
}
```

**Stream Response:**
```protobuf
message StatusUpdateEvent {
  string user_id = 1;
  UserStatus status = 2;  // ONLINE or OFFLINE
  int64 timestamp = 3;
}
```

### Heartbeat
```protobuf
rpc SendHeartbeat(SendHeartbeatRequest) returns (SendHeartbeatResponse);
```

**Request:**
```protobuf
message SendHeartbeatRequest {
  string user_id = 1;
}
```

**Response:**
```protobuf
message SendHeartbeatResponse {
  bool success = 1;
  int64 next_heartbeat_in_seconds = 2;
}
```

## Implementation Flow

### 1. User Goes Online
1. Client calls SetUserStatus(user_id, ONLINE) via gRPC
2. Server sets Redis key `user:status:{user_id}` = "online" with TTL 30s
3. Server broadcasts status change to subscribed clients via gRPC streaming
4. Client receives confirmation

### 2. User Stays Online (Heartbeat)
1. Client calls SendHeartbeat(user_id) every 15 seconds via gRPC
2. Server refreshes Redis TTL to 30 seconds
3. If no heartbeat received, Redis key expires → user becomes offline

### 3. User Goes Offline
1. Client calls SetUserStatus(user_id, OFFLINE) or disconnects
2. Server sets Redis key `user:status:{user_id}` = "offline"
3. Server broadcasts offline status to subscribed clients

### 4. Status Query
1. Client calls GetUserStatus(user_id) via gRPC
2. Server checks Redis for key existence and value
3. If key exists and value = "online" → online
4. If key doesn't exist or value = "offline" → offline

## Configuration

### Environment Variables
```bash
REDIS_URL=redis://localhost:6379
USER_STATUS_TTL=30s
HEARTBEAT_INTERVAL=15s
```

### Redis Configuration
```conf
# Enable keyspace notifications for expired keys
notify-keyspace-events Ex
```

## Error Handling

### Common Error Responses
```json
{
  "success": false,
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with ID 123 not found"
  }
}
```

### Error Codes
- `USER_NOT_FOUND`: User ID doesn't exist
- `INVALID_STATUS`: Status value is not "online" or "offline"
- `REDIS_ERROR`: Redis connection or operation failed
- `UNAUTHORIZED`: User not authorized to update status

## Performance Considerations

### Redis Optimization
- Use Redis pipelining for bulk status checks
- Implement connection pooling
- Monitor Redis memory usage

### Scaling
- Consider Redis Cluster for high availability
- Implement status caching at application level
- Use pub/sub for real-time notifications

## Testing Strategy

### Unit Tests
- Status update logic
- Redis operations
- WebSocket message handling

### Integration Tests
- End-to-end status flow
- WebSocket connection handling
- Redis TTL behavior

### Load Tests
- Concurrent user status updates
- Bulk status queries
- WebSocket connection limits

## Future Enhancements
- Last seen timestamp tracking
- User activity levels (active, idle, away)
- Presence in specific channels/rooms
- Integration with push notifications 