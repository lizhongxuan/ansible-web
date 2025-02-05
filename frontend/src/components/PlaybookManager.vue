<template>
  <div class="playbook-manager">
    <h2>{{ isEditing ? '编辑 Playbook 模板' : 'Playbook 模板管理' }}</h2>
    
    <!-- 添加模板表单 -->
    <form @submit.prevent="addTemplate" class="form">
      <div class="form-group">
        <label for="name">模板名称:</label>
        <input 
          type="text" 
          v-model="newTemplate.name" 
          id="name" 
          required 
          class="form-control"
        />
      </div>
      <div class="form-group">
        <label for="description">描述:</label>
        <input 
          type="text" 
          v-model="newTemplate.description" 
          id="description" 
          class="form-control"
        />
      </div>
      <div class="form-group">
        <label for="content">Playbook 内容:</label>
        <textarea 
          v-model="newTemplate.content" 
          id="content" 
          class="form-control code-editor"
          required
        ></textarea>
      </div>
      <div class="form-group">
        <label for="variables">变量 (每行一个):</label>
        <textarea 
          v-model="variablesText" 
          id="variables" 
          class="form-control"
          placeholder="variable1&#10;variable2"
        ></textarea>
      </div>
      <div class="form-actions">
        <button type="submit" class="btn">{{ isEditing ? '保存修改' : '保存模板' }}</button>
        <button v-if="isEditing" type="button" @click="cancelEdit" class="btn btn-secondary">取消</button>
      </div>
    </form>

    <!-- 模板列表 -->
    <div class="template-list">
      <h3>已保存的模板</h3>
      <div class="templates">
        <div v-for="template in templates" :key="template.id" class="template-item">
          <div class="template-header">
            <h4>{{ template.name }}</h4>
            <span class="template-filename">{{ template.filename }}</span>
            <div class="template-actions">
              <button @click="useTemplate(template)" class="btn btn-secondary">使用</button>
              <button @click="editTemplate(template)" class="btn btn-secondary">编辑</button>
              <button @click="deleteTemplate(template.id)" class="btn btn-danger">删除</button>
            </div>
          </div>
          <p class="template-description">{{ template.description }}</p>
          <pre class="template-content">{{ template.content }}</pre>
          <div class="template-variables">
            <strong>变量:</strong>
            <ul>
              <li v-for="(variable, index) in template.variables" :key="index">
                {{ variable }}
              </li>
            </ul>
          </div>
          <div class="template-meta">
            创建时间: {{ new Date(template.created_at).toLocaleString() }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PlaybookManager',
  data() {
    return {
      templates: [],
      newTemplate: {
        name: '',
        description: '',
        content: '',
        type: 'playbook',
        variables: []
      },
      variablesText: '',
      editingTemplate: null,
      isEditing: false
    }
  },
  methods: {
    editTemplate(template) {
      this.isEditing = true;
      this.editingTemplate = { ...template };
      this.newTemplate = { ...template };
      this.variablesText = template.variables.join('\n');
    },
    async saveEdit() {
      try {
        const template = {
          id: this.editingTemplate.id,
          name: this.newTemplate.name,
          description: this.newTemplate.description,
          content: this.newTemplate.content,
          type: this.editingTemplate.type,
          filename: this.editingTemplate.filename,
          created_at: this.editingTemplate.created_at,
          updated_at: this.editingTemplate.updated_at,
          variables: this.variablesText.split('\n').filter(v => v.trim())
        };
        
        console.log('Updating template:', template);
        
        const response = await fetch('http://localhost:8080/templates/update', {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(template)
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
        }
        
        const updatedTemplate = await response.json();
        console.log('Template updated:', updatedTemplate);
        
        await this.fetchTemplates();
        this.resetForm();
      } catch (error) {
        console.error('Error updating template:', error);
        alert('更新模板失败: ' + error.message);
      }
    },
    cancelEdit() {
      this.isEditing = false;
      this.editingTemplate = null;
      this.resetForm();
    },
    async addTemplate() {
      if (this.isEditing) {
        await this.saveEdit();
        return;
      }

      try {
        const template = {
          ...this.newTemplate,
          type: 'playbook',
          variables: this.variablesText.split('\n').filter(v => v.trim())
        };
        
        const response = await fetch('http://localhost:8080/templates/add', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(template)
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        await this.fetchTemplates();
        this.resetForm();
      } catch (error) {
        console.error('Error adding template:', error);
      }
    },
    async fetchTemplates() {
      try {
        const response = await fetch('http://localhost:8080/templates?type=playbook');
        const data = await response.json();
        this.templates = data;
      } catch (error) {
        console.error('Error fetching templates:', error);
      }
    },
    useTemplate(template) {
      // 发出事件，让父组件知道用户选择了一个模板
      this.$emit('use-template', template);
    },
    resetForm() {
      this.newTemplate = {
        name: '',
        description: '',
        content: '',
        type: 'playbook',
        variables: []
      };
      this.variablesText = '';
      this.isEditing = false;
      this.editingTemplate = null;
    }
  },
  mounted() {
    this.fetchTemplates();
  }
}
</script>

<style scoped>
.playbook-manager {
  margin-top: 20px;
}

.code-editor {
  font-family: monospace;
  min-height: 200px;
  white-space: pre;
}

.template-item {
  background: #f8f9fa;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.template-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.template-description {
  color: #666;
  margin-bottom: 10px;
}

.template-content {
  background: #fff;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 10px;
  overflow-x: auto;
}

.template-variables {
  margin-bottom: 10px;
}

.template-meta {
  font-size: 0.9em;
  color: #666;
}

.template-actions {
  display: flex;
  gap: 10px;
}

.template-filename {
  color: #666;
  font-size: 0.9em;
  font-style: italic;
}

.form-actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
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