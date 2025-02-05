<template>
  <div class="playbook-preview">
    <h3>Playbook 预览</h3>
    
    <div class="preview-content">
      <div class="preview-section">
        <h4>基本信息</h4>
        <div class="info-item">
          <strong>Playbook:</strong> {{ playbook }}
        </div>
        <div class="info-item">
          <strong>Inventory:</strong> {{ inventory }}
        </div>
      </div>

      <div class="preview-section">
        <h4>变量</h4>
        <div class="variables-editor">
          <div v-for="(value, key) in parsedVariables" :key="key" class="variable-item">
            <input 
              type="text" 
              :value="key"
              @input="updateVariableKey(key, $event.target.value)"
              class="variable-key"
            />
            <input 
              type="text" 
              :value="value"
              @input="updateVariableValue(key, $event.target.value)"
              class="variable-value"
            />
            <button @click="removeVariable(key)" class="btn-remove">×</button>
          </div>
          <button @click="addVariable" class="btn btn-secondary">添加变量</button>
        </div>
      </div>

      <div class="preview-section">
        <h4>预计影响的主机</h4>
        <div class="affected-hosts">
          <ul>
            <li v-for="host in affectedHosts" :key="host.id">
              {{ host.hostname }} ({{ host.ip }})
            </li>
          </ul>
        </div>
      </div>

      <div class="preview-actions">
        <button @click="runPlaybook" class="btn" :disabled="!isValid">
          执行 Playbook
        </button>
        <button @click="saveAsTemplate" class="btn btn-secondary">
          保存为模板
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PlaybookPreview',
  props: {
    playbook: {
      type: String,
      required: true
    },
    inventory: {
      type: String,
      required: true
    },
    variables: {
      type: String,
      default: '{}'
    }
  },
  data() {
    return {
      affectedHosts: [],
      parsedVariables: {},
      isValid: false
    }
  },
  watch: {
    variables: {
      immediate: true,
      handler(newVal) {
        try {
          this.parsedVariables = JSON.parse(newVal)
        } catch (e) {
          this.parsedVariables = {}
        }
      }
    }
  },
  methods: {
    async checkPlaybook() {
      try {
        const response = await fetch('http://localhost:8080/playbook/check', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            playbook: this.playbook,
            inventory: this.inventory,
            variables: this.parsedVariables
          })
        })
        const data = await response.json()
        this.isValid = data.valid
        this.affectedHosts = data.affected_hosts || []
      } catch (error) {
        console.error('Error checking playbook:', error)
        this.isValid = false
      }
    },
    updateVariableKey(oldKey, newKey) {
      const value = this.parsedVariables[oldKey]
      delete this.parsedVariables[oldKey]
      this.$set(this.parsedVariables, newKey, value)
      this.emitVariablesChange()
    },
    updateVariableValue(key, value) {
      this.$set(this.parsedVariables, key, value)
      this.emitVariablesChange()
    },
    removeVariable(key) {
      this.$delete(this.parsedVariables, key)
      this.emitVariablesChange()
    },
    addVariable() {
      this.$set(this.parsedVariables, `new_var_${Object.keys(this.parsedVariables).length}`, '')
      this.emitVariablesChange()
    },
    emitVariablesChange() {
      this.$emit('variables-change', JSON.stringify(this.parsedVariables))
    },
    runPlaybook() {
      this.$emit('run')
    },
    async saveAsTemplate() {
      this.$emit('save-template', {
        content: this.playbook,
        variables: Object.keys(this.parsedVariables)
      })
    }
  },
  mounted() {
    this.checkPlaybook()
  }
}
</script>

<style scoped>
.playbook-preview {
  margin-top: 20px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 4px;
}

.preview-section {
  margin-bottom: 20px;
}

.info-item {
  margin: 5px 0;
}

.variables-editor {
  margin-top: 10px;
}

.variable-item {
  display: flex;
  gap: 10px;
  margin-bottom: 5px;
}

.variable-key,
.variable-value {
  padding: 5px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.variable-key {
  width: 30%;
}

.variable-value {
  width: 60%;
}

.btn-remove {
  padding: 0 10px;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.affected-hosts {
  margin-top: 10px;
}

.affected-hosts ul {
  list-style: none;
  padding: 0;
}

.affected-hosts li {
  padding: 5px 0;
}

.preview-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style> 