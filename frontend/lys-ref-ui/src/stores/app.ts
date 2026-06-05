import { ref } from 'vue'
import { defineStore } from 'pinia'
import { type ApiError } from '@/types/app'
import logoLink from '@/assets/logo.png'

export const useAppStore = defineStore('app', () => {
  const apiErr = ref<ApiError>()

  const company = 'LoveYourStack'
  const logoUrl = logoLink
  const projectTitle = 'Reference'

  return { apiErr, company, logoUrl, projectTitle }
})

