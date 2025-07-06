# User Online Status - API Design

## gRPC Service Definition

### Protocol Buffer Schema
```protobuf
syntax = "proto3";

package users.status;

option go_package = "github.com/social-app/internal/pb/users/status";

// User Status Service
service UserStatusService {
  // Set user online status
  rpc SetUserStatus(SetUserStatusRequest) returns (SetUserStatusResponse);
  
  // Get single user status
  rpc GetUserStatus(GetUserStatusRequest) returns (GetUserStatusResponse);
  
  // Get multiple users status
  rpc GetMultipleUserStatus(GetMultipleUserStatusRequest) returns (GetMultipleUserStatusResponse);
  
  // Subscribe to user status changes (streaming)
  rpc SubscribeStatusUpdates(SubscribeStatusUpdatesRequest) returns (stream StatusUpdateEvent);
  
  // Send heartbeat to maintain online status
  rpc SendHeartbeat(SendHeartbeatRequest) returns (SendHeartbeatResponse);
}

// Enums
enum UserStatus {
  UNKNOWN = 0;
  ONLINE = 1;
  AWAY = 2;
  OFFLINE = 3;
}

// Request/Response Messages
message SetUserStatusRequest {
  string user_id = 1;
  UserStatus status = 2;
}

message SetUserStatusResponse {
  bool success = 1;
  string message = 2;
  UserStatusData data = 3;
}

message GetUserStatusRequest {
  string user_id = 1;
}

message GetUserStatusResponse {
  bool success = 1;
  UserStatusData data = 2;
}

message GetMultipleUserStatusRequest {
  repeated string user_ids = 1;
}

message GetMultipleUserStatusResponse {
  bool success = 1;
  repeated UserStatusData data = 2;
}

message SubscribeStatusUpdatesRequest {
  string user_id = 1;
  repeated string watch_user_ids = 2; // Users to watch for status changes
}

message StatusUpdateEvent {
  string user_id = 1;
  UserStatus status = 2;
  int64 timestamp = 3;
}

message SendHeartbeatRequest {
  string user_id = 1;
}

message SendHeartbeatResponse {
  bool success = 1;
  int64 next_heartbeat_in_seconds = 2;
}

// Data Models
message UserStatusData {
  string user_id = 1;
  UserStatus status = 2;
  int64 last_seen = 3; // Unix timestamp
  int64 timestamp = 4; // Status update timestamp
}
```

## API Endpoints

### 1. SetUserStatus
**Purpose**: Update user's online status

**Request**:
```json
{
  "user_id": "123",
  "status": "ONLINE"
}
```

**Response**:
```json
{
  "success": true,
  "message": "User status updated successfully",
  "data": {
    "user_id": "123",
    "status": "ONLINE",
    "last_seen": 1704067200,
    "timestamp": 1704067200
  }
}
```

### 2. GetUserStatus
**Purpose**: Get current status of a specific user

**Request**:
```json
{
  "user_id": "123"
}
```

**Response**:
```json
{
  "success": true,
  "data": {
    "user_id": "123",
    "status": "ONLINE",
    "last_seen": 1704067200,
    "timestamp": 1704067200
  }
}
```

### 3. GetMultipleUserStatus
**Purpose**: Get status of multiple users in one request

**Request**:
```json
{
  "user_ids": ["123", "456", "789"]
}
```

**Response**:
```json
{
  "success": true,
  "data": [
    {
      "user_id": "123",
      "status": "ONLINE",
      "last_seen": 1704067200,
      "timestamp": 1704067200
    },
    {
      "user_id": "456",
      "status": "AWAY",
      "last_seen": 1704067100,
      "timestamp": 1704067100
    },
    {
      "user_id": "789",
      "status": "OFFLINE",
      "last_seen": 1704067050,
      "timestamp": 1704067050
    }
  ]
}
```

### 4. SubscribeStatusUpdates (Streaming)
**Purpose**: Real-time status updates via gRPC streaming

**Request**:
```json
{
  "user_id": "123",
  "watch_user_ids": ["456", "789"]
}
```

**Stream Response**:
```json
{
  "user_id": "456",
  "status": "AWAY",
  "timestamp": 1704067200
}
```

### 5. SendHeartbeat
**Purpose**: Maintain user online status

**Request**:
```json
{
  "user_id": "123"
}
```

**Response**:
```json
{
  "success": true,
  "next_heartbeat_in_seconds": 15
}
```

## Error Handling

### Error Codes
```protobuf
enum ErrorCode {
  UNKNOWN_ERROR = 0;
  USER_NOT_FOUND = 1;
  INVALID_STATUS = 2;
  REDIS_ERROR = 3;
  UNAUTHORIZED = 4;
  RATE_LIMITED = 5;
}

message ErrorResponse {
  ErrorCode code = 1;
  string message = 2;
  map<string, string> details = 3;
}
```

### Common Error Responses
- `USER_NOT_FOUND`: User ID doesn't exist
- `INVALID_STATUS`: Invalid status value
- `REDIS_ERROR`: Redis connection/operation failed
- `UNAUTHORIZED`: User not authorized
- `RATE_LIMITED`: Too many requests

## Performance Considerations

### gRPC Optimizations
- Use connection pooling
- Implement proper timeout handling
- Use compression for large responses
- Implement circuit breaker pattern

### Streaming Best Practices
- Limit concurrent streaming connections
- Implement heartbeat for stream health
- Handle stream reconnection gracefully
- Use proper backpressure handling

## Security

### Authentication
- JWT token validation for all requests
- User authorization for status updates
- Rate limiting per user/IP

### Data Protection
- No sensitive data in status messages
- Audit logging for status changes
- Input validation and sanitization 