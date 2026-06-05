import { ref } from 'vue'
import { defineStore } from 'pinia'
import { type SelectionItem, fetchOnce } from 'lys-vue'
import ax from '@/api'

export const useEcbStore = defineStore('ecb', () => {

  const activeCurrenciesExEur = ref<SelectionItem[]>([])

  function loadActiveCurrencies() {
    const myUrl = '/a/ecb/currencies?xfields=id,name&xsort=name&is_active=true&code=!EUR&xper_page=5000'
    fetchOnce({ ax, myUrl, result: activeCurrenciesExEur })
  }

  return { 
    activeCurrenciesExEur,
    loadActiveCurrencies,
  }
})

