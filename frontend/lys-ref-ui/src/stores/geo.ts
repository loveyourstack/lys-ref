import { ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { type SelectionItem, fetchOnce } from 'lys-vue'
import ax from '@/api'
import { type Country } from '@/types/geo'

export const useGeoStore = defineStore('geo', () => {

  const countries = ref<Country[]>([])
  const euCountries = ref<Country[]>([])
  const mandatoryCountries = ref<Country[]>([])
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
    euCountries.value = countries.value.filter((c: Country) => c.is_eu)
    mandatoryCountries.value = countries.value.filter((c: Country) => c.id >= 1) // excludes -1 (None)
  })

  return { 
    countries, euCountries, mandatoryCountries, oceans,
    loadCountries, loadOceans,
  }
})

