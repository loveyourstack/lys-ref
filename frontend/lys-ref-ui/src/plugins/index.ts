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
import { createI18n } from 'vue-i18n'
import type { LocaleMessages, VueMessageType } from 'vue-i18n'
import LysVue from 'lys-vue'

// load all locale json files from the /locales folder
const localeModules = import.meta.glob('../locales/*.json', {
  eager: true,
  import: 'default',
}) as Record<string, Record<string, unknown>>

// convert the localeModules object into a messages object that can be used by vue-i18n
const messages = Object.fromEntries(
  Object.entries(localeModules).map(([path, message]) => {
    const file = path.split('/').pop() || ''
    const locale = file.replace('.json', '')
    return [locale, message]
  })
) as LocaleMessages<Record<string, VueMessageType>>

export const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  globalInjection: true,
  messages,
})

export function registerPlugins (app: App) {
 app.use(vuetify)
 app.use(i18n)
 app.use(LysVue)
 app.use(createPinia())
}