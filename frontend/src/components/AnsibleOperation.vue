<template>
  <div class="ansible-operation">
    <h2>Ansible 操作</h2>
    
    <form @submit.prevent="runAnsible" class="form">
      <div class="form-group">
        <label for="playbook">选择 Playbook 模板:</label>
        <select 
          v-model="selectedPlaybook" 
          id="playbook" 
          required 
          class="form-control"
        >
          <option value="">请选择 Playbook</option>
          <option 
            v-for="template in playbookTemplates" 
            :key="template.id" 
            :value="template"
          >
            {{ template.name }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label for="inventory">选择 Inventory 模板:</label>
        <select 
          v-model="selectedInventory" 
          id="inventory" 
          required 
          class="form-control"
        >
          <option value="">请选择 Inventory</option>
          <option 
            v-for="template in inventoryTemplates" 
            :key="template.id" 
            :value="template"
          >
            {{ template.name }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label for="variables">环境变量 (JSON):</label>
        <textarea 
          v-model="variables" 
          id="variables" 
          class="form-control"
          placeholder="{ &quot;key&quot;: &quot;value&quot; }"
        ></textarea>
      </div>

      <button type="submit" class="btn" :disabled="!selectedPlaybook || !selectedInventory">
        运行
      </button>
    </form>

    <div v-if="showExecutionStatus" class="execution-status">
      <div class="status-header">
        <h3>执行状态</h3>
        <span :class="['status-badge', status]">{{ getStatusText }}</span>
        <button @click="clearLogs" class="btn btn-sm btn-secondary">清除</button>
      </div>
      <div class="log-window">
        <div v-for="(log, index) in logs" :key="index" :class="['log-line', getLogType(log)]">
          {{ log }}
        </div>
      </div>
    </div>

    <div v-if="logs.length > 0" class="output">
      <h3>执行结果</h3>
      <pre>{{ lastOutput }}</pre>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AnsibleOperation',
  data() {
    return {
      playbookTemplates: [],
      inventoryTemplates: [],
      selectedPlaybook: '',
      selectedInventory: '',
      variables: '',
      loading: false,
      logs: [],
      status: 'running', // pending, running, complete, error
      eventSource: null
    }
  },
  computed: {
    getStatusText() {
      const statusMap = {
        pending: '等待中',
        running: '执行中',
        complete: '已完成',
        error: '执行失败'
      }
      return statusMap[this.status] || this.status
    },
    showExecutionStatus() {
      return this.loading || this.logs.length > 0
    },
    lastOutput() {
      return this.logs[this.logs.length - 1] || ''
    }
  },
  methods: {
    clearLogs() {
      this.logs = []
      this.status = 'pending'
      this.loading = false
    },
    async fetchTemplates() {
      try {
        // 获取 Playbook 模板
        const playbookResponse = await fetch('http://localhost:8080/templates?type=playbook', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json'
          }
        })
        this.playbookTemplates = await playbookResponse.json()

        // 获取 Inventory 模板
        const inventoryResponse = await fetch('http://localhost:8080/templates?type=inventory', {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json'
          }
        })
        this.inventoryTemplates = await inventoryResponse.json()
      } catch (error) {
        console.error('Error fetching templates:', error.message)
        this.$emit('error', '获取模板失败: ' + error.message)
      }
    },
    async runAnsible() {
      if (!this.selectedPlaybook || !this.selectedInventory) return

      this.loading = true
      this.logs = []
      this.status = 'running'

      // 关闭之前的 EventSource（如果存在）
      if (this.eventSource) {
        this.eventSource.close()
      }

      try {
        // 首先发送 POST 请求
        const response = await fetch('http://localhost:8080/run', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            playbook: this.selectedPlaybook.content,
            inventory: this.selectedInventory.content,
            variables: this.variables ? JSON.parse(this.variables) : {}
          })
        })
        
        // 读取响应流
        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        
        while (true) {
          const { value, done } = await reader.read()
          if (done) break
          
          const text = decoder.decode(value)
          const lines = text.split('\n')
          
          for (const line of lines) {
            if (line.startsWith('data: ')) {
              const log = line.slice(6)  // 移除 "data: " 前缀
              this.logs.push(log)
              
              // 自动滚动到最新的日志
              this.$nextTick(() => {
                const logWindow = this.$el.querySelector('.log-window')
                if (logWindow) {
                  logWindow.scrollTop = logWindow.scrollHeight
                }
              })
              
              // 检查是否完成
              if (log.includes('Command completed successfully')) {
                this.status = 'complete'
              } else if (log.includes('ERROR:')) {
                this.status = 'error'
              }
            }
          }
        }
      } catch (error) {
        this.status = 'error'
        this.logs.push(`Error: ${error.message}`)
      } finally {
        this.loading = false
      }
    },
    getLogType(log) {
      if (log.includes('ERROR:')) return 'error'
      if (log.includes('WARN:')) return 'warning'
      if (log.includes('SUCCESS:')) return 'success'
      return 'info'
    }
  },
  mounted() {
    this.fetchTemplates()
  }
}
</script>

<style scoped>
.ansible-operation {
  padding: 20px;
}

.form {
  max-width: 600px;
  margin: 0 auto;
}

.output {
  margin-top: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 4px;
}

.output pre {
  margin: 0;
  white-space: pre-wrap;
}

.execution-status {
  margin: 20px 0;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background: #f8f9fa;
  border-bottom: 1px solid #ddd;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9em;
}

.status-badge.running {
  background: #007bff;
  color: white;
}

.status-badge.complete {
  background: #28a745;
  color: white;
}

.status-badge.error {
  background: #dc3545;
  color: white;
}

.log-window {
  height: 300px;
  overflow-y: auto;
  padding: 10px;
  background: #1e1e1e;
  font-family: monospace;
  color: #d4d4d4;
}

.log-line {
  padding: 2px 0;
  white-space: pre-wrap;
}

.log-line.error {
  color: #f14c4c;
}

.log-line.warning {
  color: #ffd700;
}

.log-line.success {
  color: #3fb950;
}

.btn-sm {
  padding: 4px 8px;
  font-size: 0.875rem;
  border-radius: 3px;
}

.btn-secondary {
  background-color: #6c757d;
  color: white;
  border: none;
}

.btn-secondary:hover {
  background-color: #5a6268;
}
</style> 