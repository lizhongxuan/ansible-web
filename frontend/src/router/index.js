import Vue from 'vue'
import VueRouter from 'vue-router'
import PlaybookEditor from '../components/PlaybookEditor.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    redirect: '/ansible'
  },
  {
    path: '/ansible',
    name: 'Ansible',
    component: () => import('../components/AnsibleOperation.vue')
  },
  {
    path: '/playbook-editor',
    name: 'PlaybookEditor',
    component: PlaybookEditor
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router