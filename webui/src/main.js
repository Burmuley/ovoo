import { createApp } from 'vue'
import CoreuiVue from '@coreui/vue'
import CIcon from '@coreui/icons-vue'
import * as icons from '@coreui/icons'
import './styles/style.scss'
import App from './App.vue'

const app = createApp(App)
app.use(CoreuiVue)
app.provide('icons', icons)
app.component('CIcon', CIcon)
app.mount('#app')
