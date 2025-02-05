<template>
  <div class="inventory-manager">
    <h2>Inventory 模板管理</h2>
    
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
        <label for="content">Inventory 内容:</label>
        <textarea 
          v-model="newTemplate.content" 
          id="content" 
          class="form-control code-editor"
          required
        ></textarea>
      </div>

      <button type="submit" class="btn">保存模板</button>
    </form>

    <div class="template-list">
      <h3>已保存的模板</h3>
      <div v-for="template in templates" :key="template.id" class="template-item">
        <div class="template-header">
          <h4>{{ template.name }}</h4>
          <div class="template-actions">
            <button @click="editTemplate(template)" class="btn btn-secondary">编辑</button>
            <button @click="deleteTemplate(template.id)" class="btn btn-danger">删除</button>
          </div>
        </div>
        <p class="template-description">{{ template.description }}</p>
        <pre class="template-content">{{ template.content }}</pre>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'InventoryManager',
  data() {
    return {
      templates: [],
      newTemplate: {
        name: '',
        description: '',
        content: '',
        type: 'inventory'
      }
    }
  },
  methods: {
    async fetchTemplates() {
      try {
        const response = await fetch('http://localhost:8080/templates?type=inventory')
        this.templates = await response.json()
      } catch (error) {
        console.error('Error fetching templates:', error)
      }
    },
    async addTemplate() {
      try {
        const response = await fetch('http://localhost:8080/templates/add', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(this.newTemplate)
        })

        if (response.ok) {
          await this.fetchTemplates()
          this.resetForm()
        }
      } catch (error) {
        console.error('Error adding template:', error)
      }
    },
    resetForm() {
      this.newTemplate = {
        name: '',
        description: '',
        content: '',
        type: 'inventory'
      }
    },
    editTemplate(template) {
      this.newTemplate = { ...template }
    },
    async deleteTemplate(id) {
      if (!confirm('确定要删除这个模板吗？')) return

      try {
        const response = await fetch(`http://localhost:8080/templates/${id}`, {
          method: 'DELETE'
        })

        if (response.ok) {
          await this.fetchTemplates()
        }
      } catch (error) {
        console.error('Error deleting template:', error)
      }
    }
  },
  mounted() {
    this.fetchTemplates()
  }
}
</script>

<style scoped>
.inventory-manager {
  padding: 20px;
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
}

.template-actions {
  display: flex;
  gap: 10px;
}

.template-content {
  background: #fff;
  padding: 10px;
  border-radius: 4px;
  margin: 10px 0;
  overflow-x: auto;
}

.btn-danger {
  background: #dc3545;
}
</style> 