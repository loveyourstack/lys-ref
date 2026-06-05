import { ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { type SelectionItem, fetchOnce } from 'lys-vue'
import ax from '@/api'

export const useDigmarkStore = defineStore('digmark', () => {

  const managers = ref<string[]>([])
  const managersSelectable = ref<string[]>([])
  const verticals = ref<SelectionItem[]>([])

  function loadManagers() {
    const myUrl = '/a/digmark/managers'
    fetchOnce({ ax, myUrl, result: managers })
  }

  function loadVerticals() {
    const myUrl = '/a/digmark/verticals?xfields=id,name&xsort=name&xper_page=5000'
    fetchOnce({ ax, myUrl, result: verticals })
  }

  watch([managers], () => {
    managersSelectable.value = managers.value.filter(v => v !== 'All')
  })

  return { 
    managers, managersSelectable, verticals,
    loadManagers, loadVerticals,
  }
})

