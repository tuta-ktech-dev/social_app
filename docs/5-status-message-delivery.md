# 5-Status Message Delivery Logic (Discord Model)

## Status States Overview

### 1. **Online** 🟢
- **Behavior**: User is actively using the app
- **Visibility**: Shows as "Online" to others
- **TTL**: 30 seconds (requires heartbeat)
- **Delivery**: Real-time via WebSocket/gRPC only

### 2. **Away** 🟡
- **Behavior**: User idle for 5+ minutes
- **Visibility**: Shows as "Away" to others
- **TTL**: 10 minutes
- **Delivery**: Hybrid (WebSocket + Push Notification)

### 3. **Offline** ⚫
- **Behavior**: User disconnected or idle for 30+ minutes
- **Visibility**: Shows as "Offline" to others
- **TTL**: 24 hours
- **Delivery**: Push Notification only

### 4. **Invisible** 👻
- **Behavior**: User is online but wants to appear offline
- **Visibility**: Shows as "Offline" to others (stealth mode)
- **TTL**: 30 seconds (same as online)
- **Delivery**: Real-time via WebSocket/gRPC only (no push)

### 5. **Do Not Disturb (DND)** 🔴
- **Behavior**: User is online but wants limited notifications
- **Visibility**: Shows as "Online 🔴" to others
- **TTL**: 30 seconds (same as online)
- **Delivery**: Real-time + filtered push notifications

## Status Transition Logic

### Automatic Transitions
```
User Activity → Online (30s TTL)
↓
5 min idle → Away (10min TTL)
↓
30 min idle → Offline (24h TTL)
↓
Key expires → Unknown
```

### Manual Overrides
```
User sets Invisible → Appears offline but receives real-time
User sets DND → Appears online with 🔴 icon, limited notifications
User sets any status → Overrides automatic detection
```

### Connection-Based Transitions
```
WebSocket Connect + Activity → Online
WebSocket Connect + No Activity → Away
WebSocket Disconnect → Offline
```

## Message Delivery Matrix

| Receiver Status | WebSocket | Push Notification | Notification Type | Reason |
|-----------------|-----------|-------------------|-------------------|---------|
| **Online** 🟢 | ✅ | ❌ | Real-time only | User is actively using app |
| **Away** 🟡 | ✅ | ✅ | Both methods | Ensure delivery (app may be backgrounded) |
| **Offline** ⚫ | ❌ | ✅ | Push only | User disconnected |
| **Invisible** 👻 | ✅ | ❌ | Real-time only | Maintain stealth mode |
| **DND** 🔴 | ✅ | ⚠️ | Real-time + filtered push | Limited notifications |

## Notification Priority System

### High Priority (Always Deliver)
- Direct messages from friends
- @mentions in groups
- Friend requests
- System alerts

### Medium Priority (Respect DND)
- Group messages
- Server notifications
- Activity updates

### Low Priority (Minimal)
- Presence updates
- Typing indicators
- Read receipts

## DND Notification Rules

### Always Notify (High Priority)
```go
func shouldNotifyDND(messageType string, relationship string) bool {
    return messageType == "direct_message" ||
           messageType == "mention" ||
           messageType == "friend_request" ||
           relationship == "friend"
}
```

### Never Notify (Low Priority)
```go
func shouldSuppressDND(messageType string) bool {
    return messageType == "presence_update" ||
           messageType == "typing_indicator" ||
           messageType == "read_receipt"
}
```

## Implementation Logic

### Message Delivery Service
```go
type MessageDeliveryService struct {
    userStatusService *UserStatusService
    webSocketManager  *WebSocketManager
    pushService       *PushNotificationService
    notificationPrefs *NotificationPreferences
}

func (s *MessageDeliveryService) SendMessage(senderID, receiverID, message string, messageType string) error {
    // 1. Get receiver's actual status (not public status)
    status, err := s.userStatusService.GetUserStatus(receiverID)
    if err != nil {
        return err
    }
    
    // 2. Get notification preferences
    prefs, err := s.notificationPrefs.GetUserPreferences(receiverID)
    if err != nil {
        return err
    }
    
    // 3. Determine delivery method based on status
    switch status.Status {
    case domain.StatusOnline:
        return s.deliverRealTime(receiverID, message)
        
    case domain.StatusAway:
        return s.deliverHybrid(receiverID, message, messageType)
        
    case domain.StatusOffline:
        return s.deliverPushOnly(receiverID, message, messageType)
        
    case domain.StatusInvisible:
        return s.deliverStealth(receiverID, message)
        
    case domain.StatusDND:
        return s.deliverDND(receiverID, message, messageType, prefs)
        
    default:
        return s.deliverPushOnly(receiverID, message, messageType)
    }
}
```

### Delivery Methods
```go
// Real-time delivery for online users
func (s *MessageDeliveryService) deliverRealTime(receiverID, message string) error {
    return s.webSocketManager.SendMessage(receiverID, message)
}

// Hybrid delivery for away users
func (s *MessageDeliveryService) deliverHybrid(receiverID, message, messageType string) error {
    // Try WebSocket first (non-blocking)
    go s.webSocketManager.SendMessage(receiverID, message)
    
    // Always send push notification for away users
    return s.pushService.SendNotification(receiverID, message, messageType)
}

// Push-only delivery for offline users
func (s *MessageDeliveryService) deliverPushOnly(receiverID, message, messageType string) error {
    return s.pushService.SendNotification(receiverID, message, messageType)
}

// Stealth delivery for invisible users
func (s *MessageDeliveryService) deliverStealth(receiverID, message string) error {
    // Only WebSocket, no push notifications to maintain invisibility
    return s.webSocketManager.SendMessage(receiverID, message)
}

// Filtered delivery for DND users
func (s *MessageDeliveryService) deliverDND(receiverID, message, messageType string, prefs *NotificationPreferences) error {
    // Always send via WebSocket
    err := s.webSocketManager.SendMessage(receiverID, message)
    
    // Send push notification only for high priority messages
    if s.isHighPriority(messageType, prefs) {
        go s.pushService.SendNotification(receiverID, message, messageType)
    }
    
    return err
}
```

### Priority Checking
```go
func (s *MessageDeliveryService) isHighPriority(messageType string, prefs *NotificationPreferences) bool {
    switch messageType {
    case "direct_message":
        return prefs.AllowDirectMessages
    case "mention":
        return prefs.AllowMentions
    case "friend_request":
        return true // Always high priority
    case "group_message":
        return false // Respect DND
    default:
        return false
    }
}
```

## Status Visibility Logic

### Public Status (What Others See)
```go
func (s *UserStatusService) GetPublicUserStatus(userID string) (*domain.UserStatus, error) {
    actualStatus, err := s.GetUserStatus(userID)
    if err != nil {
        return nil, err
    }
    
    // Handle invisible mode
    publicStatus := actualStatus.Status
    if actualStatus.Status == domain.StatusInvisible {
        publicStatus = domain.StatusOffline
    }
    
    return &domain.UserStatus{
        UserID:       userID,
        Status:       publicStatus,
        ActualStatus: actualStatus.Status, // For internal use
        Timestamp:    time.Now(),
    }, nil
}
```

### Batch Status Queries
```go
func (s *UserStatusService) GetMultiplePublicUserStatus(userIDs []string) (map[string]*domain.UserStatus, error) {
    statuses := make(map[string]*domain.UserStatus)
    
    for _, userID := range userIDs {
        status, err := s.GetPublicUserStatus(userID)
        if err != nil {
            continue
        }
        statuses[userID] = status
    }
    
    return statuses, nil
}
```

## Smart Batching & Rate Limiting

### Notification Batching
```go
type NotificationBatcher struct {
    pending map[string][]*Message
    timers  map[string]*time.Timer
    delay   time.Duration
}

func (nb *NotificationBatcher) AddMessage(userID string, message *Message) {
    nb.pending[userID] = append(nb.pending[userID], message)
    
    // Reset timer for batching
    if timer, exists := nb.timers[userID]; exists {
        timer.Stop()
    }
    
    nb.timers[userID] = time.AfterFunc(nb.delay, func() {
        nb.flushMessages(userID)
    })
}

func (nb *NotificationBatcher) flushMessages(userID string) {
    messages := nb.pending[userID]
    if len(messages) == 0 {
        return
    }
    
    if len(messages) == 1 {
        // Single message
        nb.sendNotification(userID, messages[0])
    } else {
        // Batched notification
        nb.sendBatchedNotification(userID, messages)
    }
    
    // Clear pending messages
    delete(nb.pending, userID)
    delete(nb.timers, userID)
}
```

## API Endpoints

### Status Management
```
POST   /api/v1/users/:id/status              # Set status (online/away/offline/invisible/dnd)
GET    /api/v1/users/:id/status              # Get own status
GET    /api/v1/users/:id/status/public       # Get public status (visible to others)
PUT    /api/v1/users/:id/status/away         # Set away
PUT    /api/v1/users/:id/status/offline      # Set offline
PUT    /api/v1/users/:id/status/invisible    # Set invisible
PUT    /api/v1/users/:id/status/dnd          # Set do not disturb
POST   /api/v1/users/:id/heartbeat           # Send heartbeat
GET    /api/v1/users/status                  # Get multiple users status
```

### Testing Commands
```bash
# Set user to invisible
curl -X PUT http://localhost:8080/api/v1/users/user_123/status/invisible

# Set user to DND
curl -X PUT http://localhost:8080/api/v1/users/user_123/status/dnd

# Get public status (what others see)
curl http://localhost:8080/api/v1/users/user_123/status/public

# Set status via JSON
curl -X POST http://localhost:8080/api/v1/users/user_123/status \
  -H "Content-Type: application/json" \
  -d '{"status": "invisible"}'
```

## Performance Considerations

### Connection Management
- Track active WebSocket connections
- Implement connection pooling
- Handle reconnection logic
- Graceful degradation when WebSocket fails

### Redis Optimization
- Use Redis pipelines for batch operations
- Implement proper TTL management
- Monitor memory usage
- Use appropriate data structures (Sets for friend lists)

### Push Notification Optimization
- Batch notifications when possible
- Use priority queues for different message types
- Implement retry logic with exponential backoff
- Track delivery success rates

This 5-status system provides comprehensive user presence management while maintaining performance and user experience across both mobile and web platforms. 