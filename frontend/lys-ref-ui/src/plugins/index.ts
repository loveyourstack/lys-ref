import {createPinia} from 'pinia'
/**
 * plugins/index.ts
 *
 * Automatically included in `./src/main.ts`
 */

// Types
import type { App } from 'vue'

// Plugins
import vuetify from './vuetify'
import LysVue from 'lys-vue'

export function registerPlugins (app: App) {
 app.use(vuetify)
 app.use(LysVue)
 app.use(createPinia())
}