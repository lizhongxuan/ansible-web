<template>
  <div class="notification-center">
    <div class="notification-toggle" @click="toggleNotifications">
      <i class="notification-icon">🔔</i>
      <span v-if="unreadCount" class="notification-badge">{{ unreadCount }}</span>
    </div>

    <div v-if="showNotifications" class="notification-panel">
      <div class="notification-header">
        <h3>通知</h3>
        <button @click="markAllAsRead" class="btn btn-secondary">全部标记为已读</button>
      </div>

      <div class="notification-list">
        <div 
          v-for="notification in notifications" 
          :key="notification.id"
          :class="['notification-item', notification.type, { 'read': notification.read }]"
          @click="markAsRead(notification)"
        >
          <div class="notification-content">
            <div class="notification-message">{{ notification.message }}</div>
            <div class="notification-time">
              {{ new Date(notification.created_at).toLocaleString() }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NotificationCenter',
  data() {
    return {
      notifications: [],
      showNotifications: false,
      polling: null
    }
  },
  computed: {
    unreadCount() {
      return this.notifications.filter(n => !n.read).length
    }
  },
  methods: {
    async fetchNotifications() {
      try {
        const response = await fetch('http://localhost:8080/notifications')
        const data = await response.json()
        this.notifications = data
      } catch (error) {
        console.error('Error fetching notifications:', error)
      }
    },
    async markAsRead(notification) {
      if (notification.read) return

      try {
        const response = await fetch('http://localhost:8080/notifications/read', {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ id: notification.id })
        })

        if (response.ok) {
          await this.fetchNotifications()
        }
      } catch (error) {
        console.error('Error marking notification as read:', error)
      }
    },
    async markAllAsRead() {
      const unreadNotifications = this.notifications.filter(n => !n.read)
      for (const notification of unreadNotifications) {
        await this.markAsRead(notification)
      }
    },
    toggleNotifications() {
      this.showNotifications = !this.showNotifications
    },
    startPolling() {
      this.polling = setInterval(() => {
        this.fetchNotifications()
      }, 5000) // 每5秒更新一次
    },
    stopPolling() {
      if (this.polling) {
        clearInterval(this.polling)
        this.polling = null
      }
    }
  },
  mounted() {
    this.fetchNotifications()
    this.startPolling()
  },
  beforeDestroy() {
    this.stopPolling()
  }
}
</script>

<style scoped>
.notification-center {
  position: relative;
}

.notification-toggle {
  position: fixed;
  top: 20px;
  right: 20px;
  cursor: pointer;
  z-index: 1000;
}

.notification-icon {
  font-size: 24px;
}

.notification-badge {
  position: absolute;
  top: -8px;
  right: -8px;
  background: #dc3545;
  color: white;
  border-radius: 50%;
  padding: 2px 6px;
  font-size: 12px;
}

.notification-panel {
  position: fixed;
  top: 60px;
  right: 20px;
  width: 300px;
  max-height: 400px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  z-index: 1000;
  overflow-y: auto;
}

.notification-header {
  padding: 10px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.notification-list {
  padding: 10px;
}

.notification-item {
  padding: 10px;
  margin-bottom: 10px;
  border-radius: 4px;
  cursor: pointer;
}

.notification-item.read {
  opacity: 0.7;
}

.notification-item.info {
  background: #d1ecf1;
  color: #0c5460;
}

.notification-item.warning {
  background: #fff3cd;
  color: #856404;
}

.notification-item.error {
  background: #f8d7da;
  color: #721c24;
}

.notification-item.success {
  background: #d4edda;
  color: #155724;
}

.notification-time {
  font-size: 0.8em;
  color: #666;
  margin-top: 5px;
}
</style> 