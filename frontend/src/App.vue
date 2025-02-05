<template>
  <div id="app">
    <NotificationCenter />
    <h1>Ansible 控制面板</h1>
    <div class="container">
      <!-- 添加选项卡 -->
      <div class="tabs">
        <button 
          v-for="tab in tabs" 
          :key="tab.id"
          @click="currentTab = tab.id"
          :class="{ active: currentTab === tab.id }"
        >
          {{ tab.name }}
        </button>
        <button 
          @click="currentTab = 'playbook-editor'"
          :class="{ active: currentTab === 'playbook-editor' }"
        >
          可视化编辑器
        </button>
      </div>

      <!-- Ansible 操作表单 -->
      <div v-show="currentTab === 'ansible'">
        <AnsibleOperation />
      </div>

      <!-- 主机管理组件 -->
      <div v-show="currentTab === 'hosts'">
        <HostManager />
      </div>

      <div v-show="currentTab === 'playbook-templates'">
        <PlaybookManager />
      </div>

      <div v-show="currentTab === 'inventory-templates'">
        <InventoryManager />
      </div>

      <div v-show="currentTab === 'roles'">
        <RoleManager @use-role="useRole" />
      </div>

      <div v-show="currentTab === 'files'">
        <FileManager @use-file="useFile" />
      </div>

      <div v-show="currentTab === 'playbook-editor'">
        <PlaybookEditor />
      </div>
    </div>
  </div>
</template>

<script>
import HostManager from './components/HostManager.vue'
import PlaybookManager from './components/PlaybookManager.vue'
import InventoryManager from './components/InventoryManager.vue'
import RoleManager from './components/RoleManager.vue'
import FileManager from './components/FileManager.vue'
import NotificationCenter from './components/NotificationCenter.vue'
import AnsibleOperation from './components/AnsibleOperation.vue'
import PlaybookEditor from './components/PlaybookEditor.vue'

export default {
  name: 'App',
  components: {
    HostManager,
    PlaybookManager,
    InventoryManager,
    RoleManager,
    FileManager,
    NotificationCenter,
    AnsibleOperation,
    PlaybookEditor
  },
  data() {
    return {
      currentTab: 'ansible',
      tabs: [
        { id: 'ansible', name: 'Ansible 操作' },
        { id: 'hosts', name: '主机管理' },
        { id: 'playbook-templates', name: 'Playbook 模板' },
        { id: 'inventory-templates', name: 'Inventory 模板' },
        { id: 'roles', name: '角色管理' },
        { id: 'files', name: '文件管理' }
      ],
      playbook: '',
      inventory: '',
      output: '',
      tasks: [],
      loading: false,
      variables: '{}'
    }
  },
  methods: {
    async runAnsible() {
      if (!this.validateInputs()) {
        this.output = "Invalid input: Please check your playbook and inventory paths.";
        return;
      }
      this.loading = true;
      try {
        const response = await fetch('http://localhost:8080/run', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            playbook: this.playbook,
            inventory: this.inventory,
            variables: JSON.parse(this.variables)
          })
        });
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        this.output = data.output;
        await this.fetchTasks();
      } catch (error) {
        console.error('Error running Ansible:', error);
        this.output = `Error: ${error.message}`;
      } finally {
        this.loading = false;
      }
    },
    validateInputs() {
      return this.playbook.trim() !== '' && this.inventory.trim() !== '';
    },
    async fetchTasks() {
      try {
        const response = await fetch('http://localhost:8080/tasks');
        const data = await response.json();
        this.tasks = data;
      } catch (error) {
        console.error('Error fetching tasks:', error);
      }
    },
    useTemplate(template) {
      this.currentTab = 'ansible';
      this.playbook = template.content;
      // 如果需要，也可以设置其他相关字段
    },
    updateVariables(variables) {
      this.variables = variables
    },
    async saveAsTemplate(template) {
      this.currentTab = 'templates'
      // 触发模板保存
      this.$refs.playbookManager.createTemplate(template)
    },
    useRole(role) {
      console.log('Using role:', role)
    },
    useFile(file) {
      switch (file.type) {
        case 'inventory':
          this.inventory = file.content;
          break;
        case 'playbook':
          this.playbook = file.content;
          break;
      }
      this.currentTab = 'ansible';
    }
  },
  mounted() {
    this.fetchTasks();
  }
}
</script>

<style>
#app {
  font-family: Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  margin: 0;
  padding: 20px;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  height: calc(100vh - 150px);
  overflow: hidden;
}

.form {
  background: #f5f5f5;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
}

.form-control {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

textarea.form-control {
  height: 100px;
}

.btn {
  background: #42b983;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
}

.btn:hover {
  background: #3aa876;
}

.loading {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 4px;
}

.output {
  margin-top: 20px;
}

.output pre {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
}

.task-item {
  background: #f8f9fa;
  padding: 15px;
  margin-bottom: 10px;
  border-radius: 4px;
  list-style: none;
}

.task-output {
  margin-top: 10px;
  padding: 10px;
  background: #fff;
  border-radius: 4px;
  overflow-x: auto;
}

/* 添加选项卡样式 */
.tabs {
  margin-bottom: 20px;
}

.tabs button {
  padding: 10px 20px;
  margin-right: 10px;
  border: none;
  background: #f5f5f5;
  cursor: pointer;
  border-radius: 4px;
}

.tabs button.active {
  background: #42b983;
  color: white;
}

.tab-btn {
  padding: 10px 20px;
  margin-right: 10px;
  border: none;
  background: #f5f5f5;
  cursor: pointer;
  border-radius: 4px;
  text-decoration: none;
  color: inherit;
}

.tab-btn.router-link-active {
  background: #42b983;
  color: white;
}
</style>