import { ref } from 'vue'
import { defineStore } from 'pinia'
import { type SelectionItem, fetchOnce } from 'lys-vue'
import ax from '@/api'

export const useProcessStore = defineStore('process', () => {

  const flows = ref<SelectionItem[]>([])

  function loadFlows() {
    const myUrl = '/a/process/flows?xfields=id,name&xsort=name&xper_page=5000'
    fetchOnce({ ax, myUrl, result: flows })
  }

  return { 
    flows,
    loadFlows,
  }
})

