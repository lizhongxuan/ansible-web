<template>
  <div class="playbook-editor">
    <div class="left-panel">
      <div class="editor-header">
        <h3>任务类型</h3>
        <div class="search-box">
          <input v-model="searchQuery" placeholder="搜索模块..." @input="filterModules">
        </div>
      </div>
      <div class="task-types">
        <div 
          v-for="type in filteredTaskTypes"
          :key="type.id"
          class="task-type"
          draggable="true"
          @dragstart="dragStart($event, type)"
        >
          <i :class="type.icon"></i>
          <div class="task-type-info">
            <span class="task-type-name">{{ type.name }}</span>
            <span class="task-type-desc">{{ type.description }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <div class="center-panel">
      <div class="editor-header">
        <div class="playbook-info">
          <h2>Playbook 可视化编辑器</h2>
          <div class="playbook-controls">
            <input v-model="playbookName" placeholder="Playbook 名称" class="playbook-name">
            <select v-model="playbookHosts" class="playbook-hosts">
              <option value="all">所有主机</option>
              <option value="localhost">本地主机</option>
              <option value="web">Web 服务器</option>
              <option value="db">数据库服务器</option>
            </select>
          </div>
        </div>
        <div class="workflow-actions">
          <button @click="previewYaml" class="btn">预览 YAML</button>
          <button @click="savePlaybook" class="btn btn-primary">保存 Playbook</button>
        </div>
      </div>
      <div class="workflow-area" @dragover.prevent @drop="dropTask($event)">
        <div v-if="tasks.length === 0" class="empty-workflow">
          <i class="fas fa-arrow-left"></i>
          <p>从左侧拖拽任务到这里开始构建工作流</p>
        </div>
        <div 
          v-for="(task, index) in tasks" 
          :key="index"
          class="task-node"
          :class="{ selected: selectedTask === task }"
          @click="selectTask(task)"
        >
          <div class="task-header">
            <div class="task-index">{{ index + 1 }}</div>
            <span class="task-name">{{ task.name }}</span>
            <div class="task-actions">
              <button @click.stop="duplicateTask(index)" title="复制任务">
                <i class="fas fa-copy"></i>
              </button>
              <button @click.stop="moveTask(index, -1)" :disabled="index === 0">↑</button>
              <button @click.stop="moveTask(index, 1)" :disabled="index === tasks.length - 1">↓</button>
              <button @click.stop="removeTask(index)" class="delete">×</button>
            </div>
          </div>
          <div class="task-content">
            <div class="task-module">模块: {{ task.module }}</div>
            <div class="task-summary">
              <span v-if="Object.keys(task.params).length">
                参数: {{ Object.keys(task.params).length }}
              </span>
              <span v-if="Object.keys(task.variables).length">
                变量: {{ Object.keys(task.variables).length }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="right-panel">
      <div v-if="selectedTask" class="task-editor">
        <div class="editor-header">
          <h3>编辑任务</h3>
          <button @click="closeEditor" class="btn-close">×</button>
        </div>
        <div class="form-group">
          <label>任务名称:</label>
          <input v-model="selectedTask.name" type="text">
        </div>
        <div class="form-group">
          <label>模块:</label>
          <select v-model="selectedTask.module" @change="updateModuleParams">
            <option v-for="module in availableModules" :key="module" :value="module">
              {{ module }}
            </option>
          </select>
          <small class="module-help">{{ getModuleHelp(selectedTask.module) }}</small>
        </div>
        <div class="params-editor">
          <h4>参数:</h4>
          <div class="param-templates" v-if="moduleParamTemplates.length">
            <button 
              v-for="template in moduleParamTemplates"
              :key="template.name"
              @click="applyParamTemplate(template)"
              class="param-template-btn"
            >
              {{ template.name }}
            </button>
          </div>
          <div v-for="(value, key) in selectedTask.params" :key="key" class="param-item">
            <input v-model="paramKeys[key]" placeholder="参数名" @input="updateParamKey(key)">
            <input v-model="selectedTask.params[key]" placeholder="参数值">
            <button @click="removeParam(key)" class="delete">×</button>
          </div>
          <button @click="addParam" class="btn">添加参数</button>
        </div>
      </div>
      <div v-else class="empty-editor">
        <p>选择一个任务来编辑</p>
      </div>
    </div>

    <!-- YAML 预览对话框 -->
    <div v-if="showYamlPreview" class="yaml-preview-modal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>YAML 预览</h3>
          <button @click="showYamlPreview = false" class="btn-close">×</button>
        </div>
        <pre class="yaml-content">{{ yamlContent }}</pre>
        <div class="modal-footer">
          <button @click="copyYaml" class="btn">复制</button>
          <button @click="showYamlPreview = false" class="btn">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PlaybookEditor',
  data() {
    return {
      playbookName: 'New Playbook',
      playbookHosts: 'all',
      searchQuery: '',
      showYamlPreview: false,
      yamlContent: '',
      taskTypes: [
        { 
          id: 1, 
          name: '命令执行', 
          module: 'command',
          icon: 'fas fa-terminal',
          description: '执行 Shell 命令'
        },
        { 
          id: 2, 
          name: '文件操作', 
          module: 'file',
          icon: 'fas fa-file',
          description: '管理文件和目录'
        },
        { id: 3, name: '包管理', module: 'package' },
        { id: 4, name: '服务管理', module: 'service' },
        { id: 5, name: '系统信息', module: 'setup' }
      ],
      tasks: [],
      selectedTask: null,
      paramKeys: {},
      variableKeys: {},
      availableModules: [
        'command', 'shell', 'file', 'copy', 'template',
        'package', 'service', 'setup', 'debug', 'git'
      ],
      moduleHelp: {
        command: '执行 Shell 命令，支持管道和重定向',
        file: '创建、修改、删除文件和目录',
        // ... 其他模块帮助信息
      },
      moduleParamTemplates: {
        command: [
          { name: '带超时的命令', params: { cmd: '', timeout: 30 } },
          { name: '后台运行', params: { cmd: '', async: 1 } }
        ],
        file: [
          { name: '创建目录', params: { path: '', state: 'directory' } },
          { name: '删除文件', params: { path: '', state: 'absent' } }
        ]
        // ... 其他模块的参数模板
      }
    }
  },
  computed: {
    filteredTaskTypes() {
      if (!this.searchQuery) return this.taskTypes
      const query = this.searchQuery.toLowerCase()
      return this.taskTypes.filter(type => 
        type.name.toLowerCase().includes(query) ||
        type.description.toLowerCase().includes(query)
      )
    }
  },
  methods: {
    dragStart(event, type) {
      event.dataTransfer.setData('taskType', JSON.stringify(type))
    },
    dropTask(event) {
      const type = JSON.parse(event.dataTransfer.getData('taskType'))
      this.tasks.push({
        name: `新建${type.name}`,
        module: type.module,
        params: {},
        variables: {}
      })
    },
    selectTask(task) {
      this.selectedTask = task
      this.paramKeys = { ...task.params }
      this.variableKeys = { ...task.variables }
    },
    moveTask(index, direction) {
      const newIndex = index + direction
      if (newIndex >= 0 && newIndex < this.tasks.length) {
        const task = this.tasks.splice(index, 1)[0]
        this.tasks.splice(newIndex, 0, task)
      }
    },
    removeTask(index) {
      this.tasks.splice(index, 1)
      if (this.selectedTask === this.tasks[index]) {
        this.selectedTask = null
      }
    },
    addParam() {
      const key = `param_${Object.keys(this.selectedTask.params).length + 1}`
      this.$set(this.selectedTask.params, key, '')
      this.$set(this.paramKeys, key, key)
    },
    removeParam(key) {
      this.$delete(this.selectedTask.params, key)
      this.$delete(this.paramKeys, key)
    },
    updateParamKey(oldKey) {
      const newKey = this.paramKeys[oldKey]
      if (newKey !== oldKey) {
        const value = this.selectedTask.params[oldKey]
        this.$delete(this.selectedTask.params, oldKey)
        this.$set(this.selectedTask.params, newKey, value)
      }
    },
    addVariable() {
      const key = `var_${Object.keys(this.selectedTask.variables).length + 1}`
      this.$set(this.selectedTask.variables, key, '')
      this.$set(this.variableKeys, key, key)
    },
    removeVariable(key) {
      this.$delete(this.selectedTask.variables, key)
      this.$delete(this.variableKeys, key)
    },
    updateVariableKey(oldKey) {
      const newKey = this.variableKeys[oldKey]
      if (newKey !== oldKey) {
        const value = this.selectedTask.variables[oldKey]
        this.$delete(this.selectedTask.variables, oldKey)
        this.$set(this.selectedTask.variables, newKey, value)
      }
    },
    savePlaybook() {
      const playbook = {
        name: 'Generated Playbook',
        hosts: 'all',
        tasks: this.tasks.map(task => ({
          name: task.name,
          [task.module]: task.params,
          vars: task.variables
        }))
      }
      
      const yamlContent = `---
- name: ${playbook.name}
- hosts: ${playbook.hosts}
- tasks:
${this.tasksToYaml(playbook.tasks)}`

      // 发送到后端保存
      this.saveTemplate(yamlContent)
    },
    tasksToYaml(tasks) {
      return tasks.map(task => `    - name: ${task.name}
      ${task.module}:
${Object.entries(task.params).map(([key, value]) => `        ${key}: ${value}`).join('\n')}
      vars:
${Object.entries(task.vars).map(([key, value]) => `        ${key}: ${value}`).join('\n')}`
      ).join('\n')
    },
    async saveTemplate(content) {
      try {
        const template = {
          name: 'Generated Playbook',
          description: '通过可视化编辑器生成的 Playbook',
          content: content,
          type: 'playbook',
          variables: []
        }
        
        const response = await fetch('http://localhost:8080/templates/add', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(template)
        })
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        alert('Playbook 保存成功！')
      } catch (error) {
        console.error('Error saving template:', error)
        alert('保存失败: ' + error.message)
      }
    },
    duplicateTask(index) {
      const task = JSON.parse(JSON.stringify(this.tasks[index]))
      task.name = `${task.name} (复制)`
      this.tasks.splice(index + 1, 0, task)
    },
    closeEditor() {
      this.selectedTask = null
    },
    getModuleHelp(module) {
      return this.moduleHelp[module] || ''
    },
    updateModuleParams(event) {
      const module = event.target.value
      this.selectedTask.params = {}
      this.paramKeys = {}
    },
    applyParamTemplate(template) {
      this.selectedTask.params = { ...template.params }
      this.paramKeys = { ...template.params }
    },
    previewYaml() {
      this.yamlContent = this.generateYaml()
      this.showYamlPreview = true
    },
    copyYaml() {
      navigator.clipboard.writeText(this.yamlContent)
        .then(() => alert('YAML 已复制到剪贴板'))
        .catch(err => console.error('复制失败:', err))
    },
    generateYaml() {
      const playbook = {
        name: this.playbookName,
        hosts: this.playbookHosts,
        tasks: this.tasks.map(task => ({
          name: task.name,
          [task.module]: task.params,
          vars: task.variables
        }))
      }
      
      return `---
- name: ${playbook.name}
  hosts: ${playbook.hosts}
  tasks:
${this.tasksToYaml(playbook.tasks)}`
    },
    filterModules() {
      // 过滤逻辑已经在 computed 属性中实现
    }
  }
}
</script>

<style scoped>
.playbook-editor {
  display: grid;
  grid-template-columns: 250px 1fr 300px;
  gap: 20px;
  height: 100vh;
  overflow: hidden;
}

.left-panel,
.center-panel,
.right-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.left-panel {
  padding: 15px;
  background: #f8f9fa;
}

.center-panel {
  display: flex;
  flex-direction: column;
}

.right-panel {
  padding: 15px;
  background: #f8f9fa;
}

.editor-header {
  padding: 15px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #eee;
}

.playbook-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.playbook-controls {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.workflow-area {
  flex: 1;
  background: #f8f9fa;
  padding: 20px;
  overflow-y: auto;
}

.task-types {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow-y: auto;
  margin-top: 15px;
}

.task-type {
  padding: 10px;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: move;
  display: flex;
  align-items: center;
  gap: 10px;
}

.task-type-info {
  display: flex;
  flex-direction: column;
}

.task-type-desc {
  font-size: 0.8em;
  color: #666;
}

.task-node {
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-bottom: 10px;
  padding: 10px;
}

.task-node.selected {
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.task-actions {
  display: flex;
  gap: 5px;
}

.task-content {
  font-size: 0.9em;
  color: #666;
}

.task-editor {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 4px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
}

.params-editor,
.variables-editor {
  margin-top: 15px;
}

.param-item,
.variable-item {
  display: flex;
  gap: 5px;
  margin-bottom: 5px;
}

.delete {
  color: #dc3545;
  border: none;
  background: none;
  cursor: pointer;
}

.btn {
  background: #007bff;
  color: white;
  border: none;
  padding: 8px 15px;
  border-radius: 4px;
  cursor: pointer;
}

.btn:hover {
  background: #0056b3;
}

input,
select {
  padding: 5px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.empty-workflow {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #666;
  border: 2px dashed #ddd;
  border-radius: 4px;
}

.task-index {
  width: 24px;
  height: 24px;
  background: #007bff;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9em;
}

.yaml-preview-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 4px;
  width: 80%;
  max-width: 800px;
  max-height: 80vh;
  overflow-y: auto;
}
</style> 