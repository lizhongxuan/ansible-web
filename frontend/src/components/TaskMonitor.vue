<template>
  <div class="task-monitor">
    <h2>任务监控</h2>
    
    <div class="task-list">
      <div v-for="task in tasks" :key="task.id" class="task-item">
        <div class="task-header">
          <h3>任务 #{{ task.id }}</h3>
          <span :class="['status-badge', task.status]">{{ getStatusText(task.status) }}</span>
        </div>
        
        <div class="task-details">
          <div>Playbook: {{ task.playbook }}</div>
          <div>Inventory: {{ task.inventory }}</div>
          <div>开始时间: {{ new Date(task.start_time).toLocaleString() }}</div>
          <div v-if="task.end_time">
            结束时间: {{ new Date(task.end_time).toLocaleString() }}
          </div>
        </div>

        <div class="progress-bar" v-if="task.status === 'running'">
          <div class="progress" :style="{ width: task.progress + '%' }"></div>
        </div>

        <div class="task-logs" v-if="selectedTaskId === task.id">
          <h4>任务日志</h4>
          <div v-for="log in taskLogs" :key="log.id" :class="['log-entry', log.level]">
            <span class="log-time">{{ new Date(log.timestamp).toLocaleString() }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>

        <div class="task-actions">
          <button 
            @click="toggleLogs(task.id)" 
            class="btn btn-secondary"
          >
            {{ selectedTaskId === task.id ? '隐藏日志' : '显示日志' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TaskMonitor',
  data() {
    return {
      tasks: [],
      selectedTaskId: null,
      taskLogs: [],
      polling: null
    }
  },
  methods: {
    getStatusText(status) {
      const statusMap = {
        pending: '等待中',
        running: '运行中',
        complete: '已完成',
        failed: '失败'
      }
      return statusMap[status] || status
    },
    async fetchTasks() {
      try {
        const response = await fetch('http://localhost:8080/tasks')
        const data = await response.json()
        this.tasks = data
      } catch (error) {
        console.error('Error fetching tasks:', error)
      }
    },
    async fetchTaskLogs(taskId) {
      try {
        const response = await fetch(`http://localhost:8080/tasks/logs?task_id=${taskId}`)
        const data = await response.json()
        this.taskLogs = data
      } catch (error) {
        console.error('Error fetching task logs:', error)
      }
    },
    async toggleLogs(taskId) {
      if (this.selectedTaskId === taskId) {
        this.selectedTaskId = null
        this.taskLogs = []
      } else {
        this.selectedTaskId = taskId
        await this.fetchTaskLogs(taskId)
      }
    },
    startPolling() {
      this.polling = setInterval(async () => {
        await this.fetchTasks()
        if (this.selectedTaskId) {
          await this.fetchTaskLogs(this.selectedTaskId)
        }
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
    this.fetchTasks()
    this.startPolling()
  },
  beforeDestroy() {
    this.stopPolling()
  }
}
</script>

<style scoped>
.task-monitor {
  margin-top: 20px;
}

.task-item {
  background: #f8f9fa;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9em;
}

.status-badge.pending {
  background: #ffd700;
  color: #000;
}

.status-badge.running {
  background: #007bff;
  color: #fff;
}

.status-badge.complete {
  background: #28a745;
  color: #fff;
}

.status-badge.failed {
  background: #dc3545;
  color: #fff;
}

.task-details {
  margin-bottom: 10px;
}

.progress-bar {
  background: #e9ecef;
  border-radius: 4px;
  height: 20px;
  margin: 10px 0;
  overflow: hidden;
}

.progress {
  background: #007bff;
  height: 100%;
  transition: width 0.3s ease;
}

.task-logs {
  margin-top: 10px;
  padding: 10px;
  background: #fff;
  border-radius: 4px;
}

.log-entry {
  padding: 5px;
  margin: 2px 0;
  font-family: monospace;
}

.log-entry.info {
  color: #0c5460;
  background-color: #d1ecf1;
}

.log-entry.warning {
  color: #856404;
  background-color: #fff3cd;
}

.log-entry.error {
  color: #721c24;
  background-color: #f8d7da;
}

.log-time {
  margin-right: 10px;
  font-size: 0.9em;
}

.task-actions {
  margin-top: 10px;
}
</style> 