<template>
  <div class="role-manager">
    <h2>角色管理</h2>
    
    <!-- 添加角色表单 -->
    <form @submit.prevent="addRole" class="form">
      <div class="form-group">
        <label for="name">角色名称:</label>
        <input 
          type="text" 
          v-model="newRole.name" 
          id="name" 
          required 
          class="form-control"
        />
      </div>
      
      <div class="form-group">
        <label for="description">描述:</label>
        <textarea 
          v-model="newRole.description" 
          id="description" 
          class="form-control"
        ></textarea>
      </div>
      
      <!-- 任务配置 -->
      <div class="form-group">
        <label>任务:</label>
        <div class="tasks-editor">
          <div v-for="(task, index) in newRole.tasks" :key="index" class="task-item">
            <input 
              type="text" 
              v-model="task.name"
              placeholder="任务名称"
              class="form-control"
            />
            <textarea 
              v-model="task.content"
              placeholder="任务内容 (YAML)"
              class="form-control code-editor"
            ></textarea>
            <button @click="removeTask(index)" type="button" class="btn-remove">删除任务</button>
          </div>
          <button @click="addTask" type="button" class="btn btn-secondary">添加任务</button>
        </div>
      </div>
      
      <!-- 变量配置 -->
      <div class="form-group">
        <label>默认变量:</label>
        <div class="variables-editor">
          <div v-for="(value, key) in newRole.defaults" :key="key" class="variable-item">
            <input 
              type="text" 
              :value="key"
              @input="updateDefaultKey(key, $event.target.value)"
              placeholder="变量名"
              class="form-control"
            />
            <input 
              type="text" 
              :value="value"
              @input="updateDefaultValue(key, $event.target.value)"
              placeholder="默认值"
              class="form-control"
            />
            <button @click="removeDefault(key)" type="button" class="btn-remove">×</button>
          </div>
          <button @click="addDefault" type="button" class="btn btn-secondary">添加默认变量</button>
        </div>
      </div>
      
      <!-- 依赖配置 -->
      <div class="form-group">
        <label>依赖角色:</label>
        <div class="dependencies-editor">
          <div v-for="(dep, index) in newRole.dependencies" :key="index" class="dependency-item">
            <input 
              type="text" 
              v-model="newRole.dependencies[index]"
              placeholder="依赖角色名称"
              class="form-control"
            />
            <button @click="removeDependency(index)" type="button" class="btn-remove">×</button>
          </div>
          <button @click="addDependency" type="button" class="btn btn-secondary">添加依赖</button>
        </div>
      </div>
      
      <button type="submit" class="btn">保存角色</button>
    </form>

    <!-- 角色列表 -->
    <div class="roles-list">
      <h3>已创建的角色</h3>
      <div class="roles">
        <div v-for="role in roles" :key="role.id" class="role-item">
          <div class="role-header">
            <h4>{{ role.name }}</h4>
            <div class="role-actions">
              <button @click="editRole(role)" class="btn btn-secondary">编辑</button>
              <button @click="exportRole(role)" class="btn btn-secondary">导出</button>
              <button @click="deleteRole(role.id)" class="btn btn-danger">删除</button>
            </div>
          </div>
          <p class="role-description">{{ role.description }}</p>
          
          <div class="role-details">
            <div class="tasks-list">
              <strong>任务:</strong>
              <ul>
                <li v-for="task in role.tasks" :key="task.name">{{ task.name }}</li>
              </ul>
            </div>
            
            <div class="defaults-list">
              <strong>默认变量:</strong>
              <ul>
                <li v-for="(value, key) in role.defaults" :key="key">{{ key }}: {{ value }}</li>
              </ul>
            </div>
            
            <div class="dependencies-list">
              <strong>依赖:</strong>
              <ul>
                <li v-for="dep in role.dependencies" :key="dep">{{ dep }}</li>
              </ul>
            </div>
          </div>
          
          <div class="role-meta">
            创建时间: {{ new Date(role.created_at).toLocaleString() }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'RoleManager',
  data() {
    return {
      roles: [],
      newRole: {
        name: '',
        description: '',
        tasks: [],
        defaults: {},
        dependencies: []
      }
    }
  },
  methods: {
    async addRole() {
      try {
        const response = await fetch('http://localhost:8080/roles/add', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(this.newRole)
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        await this.fetchRoles();
        this.resetForm();
      } catch (error) {
        console.error('Error adding role:', error);
      }
    },
    async fetchRoles() {
      try {
        const response = await fetch('http://localhost:8080/roles');
        const data = await response.json();
        this.roles = data;
      } catch (error) {
        console.error('Error fetching roles:', error);
      }
    },
    addTask() {
      this.newRole.tasks.push({
        name: '',
        content: ''
      });
    },
    removeTask(index) {
      this.newRole.tasks.splice(index, 1);
    },
    addDefault() {
      this.newRole.defaults = { ...this.newRole.defaults, [this.newRole.defaults.length]: '' };
    },
    removeDefault(key) {
      delete this.newRole.defaults[key];
    },
    updateDefaultKey(key, value) {
      this.newRole.defaults[key] = value;
    },
    updateDefaultValue(key, value) {
      this.newRole.defaults[key] = value;
    },
    addDependency() {
      this.newRole.dependencies.push('');
    },
    removeDependency(index) {
      this.newRole.dependencies.splice(index, 1);
    },
    editRole(role) {
      this.newRole = { ...role };
    },
    exportRole(role) {
      this.$emit('export-role', role);
    },
    deleteRole(id) {
      this.$emit('delete-role', id);
    },
    resetForm() {
      this.newRole = {
        name: '',
        description: '',
        tasks: [],
        defaults: {},
        dependencies: []
      };
    }
  },
  mounted() {
    this.fetchRoles();
  }
}
</script>

<style scoped>
.role-manager {
  margin-top: 20px;
}

.tasks-editor,
.variables-editor,
.dependencies-editor {
  margin-top: 10px;
}

.task-item {
  margin-bottom: 15px;
  padding: 10px;
  background: #fff;
  border-radius: 4px;
}

.code-editor {
  font-family: monospace;
  min-height: 100px;
  margin-top: 5px;
}

.role-item {
  background: #f8f9fa;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.role-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.role-actions {
  display: flex;
  gap: 10px;
}

.role-description {
  color: #666;
  margin-bottom: 10px;
}

.role-details {
  margin: 15px 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.role-meta {
  font-size: 0.9em;
  color: #666;
}

.btn-remove {
  padding: 0 10px;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-left: 10px;
}

ul {
  list-style: none;
  padding-left: 0;
}

li {
  margin: 5px 0;
}
</style> 