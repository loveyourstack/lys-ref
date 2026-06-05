import { ref } from 'vue'
import { defineStore } from 'pinia'
import { type SelectionItem, fetchOnce } from 'lys-vue'
import ax from '@/api'

export const useSupplierStore = defineStore('supplier', () => {

  const companies = ref<SelectionItem[]>([])
  const productCategories = ref<SelectionItem[]>([])
  const selectedEmpEmail = ref<string>('elio.rossi@example.com')
  const selectedCompId = ref<number>(2)

  function loadCompanies() {
    const myUrl = '/a/supplier/companies?xfields=id,name&xsort=name&xper_page=5000'
    fetchOnce({ ax, myUrl, result: companies })
  }

  function loadProductCategories() {
    const myUrl = '/a/supplier/product-categories?xfields=id,name&xsort=name&xper_page=5000'
    fetchOnce({ ax, myUrl, result: productCategories })
  }

  return { 
    companies, productCategories, selectedEmpEmail, selectedCompId,
    loadCompanies, loadProductCategories,
  }
})

