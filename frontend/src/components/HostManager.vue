<template>
  <div class="host-manager">
    <h2>主机管理</h2>
    
    <!-- 添加主机表单 -->
    <form @submit.prevent="addHost" class="form">
      <div class="form-group">
        <label for="hostname">主机名:</label>
        <input 
          type="text" 
          v-model="newHost.hostname" 
          id="hostname" 
          required 
          class="form-control"
        />
      </div>
      <div class="form-group">
        <label for="ip">IP地址:</label>
        <input 
          type="text" 
          v-model="newHost.ip" 
          id="ip" 
          required 
          class="form-control"
        />
      </div>
      <div class="form-group">
        <label for="group">组:</label>
        <input 
          type="text" 
          v-model="newHost.group" 
          id="group" 
          class="form-control"
        />
      </div>
      <div class="form-group">
        <label for="description">描述:</label>
        <textarea 
          v-model="newHost.description" 
          id="description" 
          class="form-control"
        ></textarea>
      </div>
      <button type="submit" class="btn">添加主机</button>
    </form>

    <!-- 主机列表 -->
    <div class="host-list">
      <h3>主机列表</h3>
      <button @click="checkHealth" class="btn btn-secondary">检查健康状态</button>
      <table class="hosts-table">
        <thead>
          <tr>
            <th>主机名</th>
            <th>IP地址</th>
            <th>组</th>
            <th>状态</th>
            <th>最后检查时间</th>
            <th>描述</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="host in hosts" :key="host.id" :class="{'unhealthy': host.status === 'unhealthy'}">
            <td>{{ host.hostname }}</td>
            <td>{{ host.ip }}</td>
            <td>{{ host.group }}</td>
            <td>{{ host.status }}</td>
            <td>{{ new Date(host.last_check).toLocaleString() }}</td>
            <td>{{ host.description }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
export default {
  name: 'HostManager',
  data() {
    return {
      hosts: [],
      newHost: {
        hostname: '',
        ip: '',
        group: '',
        description: ''
      }
    }
  },
  methods: {
    async addHost() {
      try {
        const response = await fetch('http://localhost:8080/hosts/add', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(this.newHost)
        });
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        await this.fetchHosts();
        // 清空表单
        this.newHost = {
          hostname: '',
          ip: '',
          group: '',
          description: ''
        };
      } catch (error) {
        console.error('Error adding host:', error);
      }
    },
    async fetchHosts() {
      try {
        const response = await fetch('http://localhost:8080/hosts');
        const data = await response.json();
        this.hosts = data;
      } catch (error) {
        console.error('Error fetching hosts:', error);
      }
    },
    async checkHealth() {
      try {
        const response = await fetch('http://localhost:8080/hosts/health');
        const data = await response.json();
        this.hosts = data;
      } catch (error) {
        console.error('Error checking health:', error);
      }
    }
  },
  mounted() {
    this.fetchHosts();
  }
}
</script>

<style scoped>
.host-manager {
  margin-top: 20px;
}

.hosts-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.hosts-table th,
.hosts-table td {
  padding: 10px;
  border: 1px solid #ddd;
  text-align: left;
}

.hosts-table th {
  background-color: #f5f5f5;
}

.unhealthy {
  background-color: #ffebee;
}

.btn-secondary {
  background-color: #6c757d;
  margin-bottom: 10px;
}

.btn-secondary:hover {
  background-color: #5a6268;
}
</style> 