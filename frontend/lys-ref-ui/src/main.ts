/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Composables
import { createApp } from 'vue'

// Plugins
import { registerPlugins } from '@/plugins'

// Components
import App from './App.vue'

// Styles
import 'unfonts.css'

import auth from '@/auth'
import router from '@/router'

const app = createApp(App)
registerPlugins(app)

// complete auth bootstrap before router (due to nav guard) and app mount
await auth.bootstrap()

app.use(router)

app.mount('#app')
