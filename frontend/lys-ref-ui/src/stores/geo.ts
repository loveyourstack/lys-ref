import { ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { type SelectionItem, fetchOnce } from 'lys-vue'
import ax from '@/api'

export const useGeoStore = defineStore('geo', () => {

  const countries = ref<SelectionItem[]>([])
  const mandatoryCountries = ref<SelectionItem[]>([])
  const oceans = ref<SelectionItem[]>([])

  function loadCountries() {
    const myUrl = '/a/geo/countries?xfields=id,name&xsort=name&xper_page=5000'
    fetchOnce({ ax, myUrl, result: countries })
  }

  function loadOceans() {
    const myUrl = '/a/geo/oceans?xfields=id,name&xsort=name&xper_page=5000'
    fetchOnce({ ax, myUrl, result: oceans })
  }

  watch([countries], () => {
    mandatoryCountries.value = countries.value.filter((c: SelectionItem) => c.id >= 1) // excludes -1 (None)
  })

  return { 
    countries, mandatoryCountries, oceans,
    loadCountries, loadOceans,
  }
})

