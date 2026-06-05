import { ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { fetchOnce } from 'lys-vue'
import ax from '@/api'

export const useCoreStore = defineStore('core', () => {

  const mandatoryEnums = ref<string[]>([])
  const optionalEnums = ref<string[]>([])
  const periods = ref<string[]>([])
  const xrPeriods = ref<string[]>([])
  
  function loadMandatoryEnums() {
    const myUrl = '/a/core/mandatory-enums'
    fetchOnce({ ax, myUrl, result: mandatoryEnums })
  }

  function loadOptionalEnums() {
    const myUrl = '/a/core/optional-enums'
    fetchOnce({ ax, myUrl, result: optionalEnums })
  }

  function loadPeriods() {
    const myUrl = '/a/core/performance-periods'
    fetchOnce({ ax, myUrl, result: periods })
  }

  // use watch rather than onSuccess callback since store values are set directly on initialization in Main.vue, not just via load funcs

  watch([periods], () => {
    xrPeriods.value = periods.value.filter((p: string) => !['Today', 'Yesterday', 'Last 3 days'].includes(p))
  })

  return { 
    mandatoryEnums, optionalEnums, periods, xrPeriods,
    loadMandatoryEnums, loadOptionalEnums, loadPeriods,
  }
})

