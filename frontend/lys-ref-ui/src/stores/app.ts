import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { i18n } from '@/plugins'
import { type ApiError } from '@/types/app'
import logoLink from '@/assets/logo.png'

export const useAppStore = defineStore('app', () => {
  const apiErr = ref<ApiError>()

  // locales are imported and defined in plugins/index.ts: use that exported object
  const locales = computed(() =>
    i18n.global.availableLocales.map(code => ({
      code,
      name: new Intl.DisplayNames([code], { type: 'language' }).of(code) ?? code,
    }))
  )  
  
  const company = 'LoveYourStack'
  const logoUrl = logoLink
  const projectTitle = 'Reference'

  return { apiErr, locales,
    company, logoUrl, projectTitle }
})

