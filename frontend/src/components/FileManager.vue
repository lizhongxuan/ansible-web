<template>
  <div class="file-manager">
    <h2>文件管理</h2>
    
    <!-- 文件类型过滤器 -->
    <div class="file-filters">
      <button 
        v-for="type in fileTypes" 
        :key="type.id"
        @click="currentType = type.id"
        :class="['filter-btn', { active: currentType === type.id }]"
      >
        {{ type.name }}
      </button>
    </div>

    <!-- 添加文件表单 -->
    <form @submit.prevent="addFile" class="form">
      <div class="form-group">
        <label for="name">文件名称:</label>
        <input 
          type="text" 
          v-model="newFile.name" 
          id="name" 
          required 
          class="form-control"
        />
      </div>

      <div class="form-group">
        <label for="type">文件类型:</label>
        <select 
          v-model="newFile.type" 
          id="type" 
          required 
          class="form-control"
        >
          <option v-for="type in fileTypes" :key="type.id" :value="type.id">
            {{ type.name }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label for="description">描述:</label>
        <textarea 
          v-model="newFile.description" 
          id="description" 
          class="form-control"
        ></textarea>
      </div>

      <div class="form-group">
        <label for="content">内容:</label>
        <textarea 
          v-model="newFile.content" 
          id="content" 
          class="form-control code-editor"
          required
        ></textarea>
      </div>

      <button type="submit" class="btn">{{ editMode ? '更新' : '添加' }}文件</button>
      <button 
        v-if="editMode" 
        type="button" 
        @click="cancelEdit" 
        class="btn btn-secondary"
      >
        取消
      </button>
    </form>

    <!-- 文件列表 -->
    <div class="files-list">
      <h3>文件列表</h3>
      <div class="files">
        <div 
          v-for="file in filteredFiles" 
          :key="file.id" 
          class="file-item"
        >
          <div class="file-header">
            <h4>{{ file.name }}</h4>
            <div class="file-actions">
              <button @click="editFile(file)" class="btn btn-secondary">编辑</button>
              <button @click="useFile(file)" class="btn">使用</button>
            </div>
          </div>
          <p class="file-description">{{ file.description }}</p>
          <div class="file-type">类型: {{ getFileTypeName(file.type) }}</div>
          <pre class="file-content">{{ file.content }}</pre>
          <div class="file-meta">
            创建时间: {{ new Date(file.created_at).toLocaleString() }}
            <br>
            更新时间: {{ new Date(file.updated_at).toLocaleString() }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'FileManager',
  data() {
    return {
      files: [],
      currentType: 'all',
      editMode: false,
      fileTypes: [
        { id: 'all', name: '全部' },
        { id: 'inventory', name: 'Inventory' },
        { id: 'playbook', name: 'Playbook' },
        { id: 'config', name: '配置文件' }
      ],
      newFile: {
        name: '',
        type: '',
        content: '',
        description: ''
      }
    }
  },
  computed: {
    filteredFiles() {
      if (this.currentType === 'all') {
        return this.files;
      }
      return this.files.filter(file => file.type === this.currentType);
    }
  },
  methods: {
    async addFile() {
      try {
        const url = this.editMode ? 
          'http://localhost:8080/files/update' : 
          'http://localhost:8080/files/add';
        
        const method = this.editMode ? 'PUT' : 'POST';
        
        const response = await fetch(url, {
          method,
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(this.newFile)
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        await this.fetchFiles();
        this.resetForm();
      } catch (error) {
        console.error('Error saving file:', error);
      }
    },
    async fetchFiles() {
      try {
        const response = await fetch('http://localhost:8080/files');
        const data = await response.json();
        this.files = data;
      } catch (error) {
        console.error('Error fetching files:', error);
      }
    },
    editFile(file) {
      this.editMode = true;
      this.newFile = { ...file };
    },
    cancelEdit() {
      this.editMode = false;
      this.resetForm();
    },
    useFile(file) {
      this.$emit('use-file', file);
    },
    resetForm() {
      this.editMode = false;
      this.newFile = {
        name: '',
        type: '',
        content: '',
        description: ''
      };
    },
    getFileTypeName(type) {
      const fileType = this.fileTypes.find(t => t.id === type);
      return fileType ? fileType.name : type;
    }
  },
  mounted() {
    this.fetchFiles();
  }
}
</script>

<style scoped>
.file-manager {
  margin-top: 20px;
}

.file-filters {
  margin-bottom: 20px;
}

.filter-btn {
  padding: 8px 16px;
  margin-right: 10px;
  border: none;
  background: #f5f5f5;
  border-radius: 4px;
  cursor: pointer;
}

.filter-btn.active {
  background: #42b983;
  color: white;
}

.code-editor {
  font-family: monospace;
  min-height: 200px;
}

.file-item {
  background: #f8f9fa;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.file-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.file-actions {
  display: flex;
  gap: 10px;
}

.file-description {
  color: #666;
  margin-bottom: 10px;
}

.file-type {
  font-size: 0.9em;
  color: #666;
  margin-bottom: 10px;
}

.file-content {
  background: #fff;
  padding: 10px;
  border-radius: 4px;
  margin: 10px 0;
  overflow-x: auto;
}

.file-meta {
  font-size: 0.9em;
  color: #666;
}
</style> 