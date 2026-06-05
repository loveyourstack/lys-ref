import { ref } from 'vue'
import { defineStore } from 'pinia'
import { type SelectionItem, fetchOnce } from 'lys-vue'
import ax from '@/api'

export const usePublisherStore = defineStore('publisher', () => {

  const authors = ref<SelectionItem[]>([])

  function loadAuthors() {
    const myUrl = '/a/publisher/authors?xfields=id,name&xsort=name&xper_page=5000'
    fetchOnce({ ax, myUrl, result: authors })
  }

  return { 
    authors,
    loadAuthors,
  }
})

