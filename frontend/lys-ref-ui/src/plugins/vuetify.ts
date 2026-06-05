/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        dark: false,
        colors: {
          primary: '#E06000',
          'on-primary': '#FFFFFF',
          secondary: '#1D58D6',
          'on-secondary': '#FFFFFF',
          info: '#2D77F3',
        },
      },
      dark: {
        dark: true,
        colors: {
          primary: '#FFB066',
          'on-primary': '#201104',
          secondary: '#78BAFF',
          'on-secondary': '#0A1C36',
          info: '#98C9FF',
        },
      },
    },
  },
})
