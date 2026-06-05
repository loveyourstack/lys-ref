import axios, { type AxiosInstance } from 'axios'
import { useAppStore } from '@/stores/app'

function createAxiosInstance(baseURL: string): AxiosInstance {
  const instance = axios.create({
    baseURL,
    headers: {
      'Content-type': 'application/json',
    },
  })

  instance.interceptors.response.use(
    (response) => { return response },
    (error) => {
      const appStore = useAppStore()
      //console.log(error.toJSON())

      // special handling for blob request (file downloads: see file.ts) errors in order to get err_description
      // from https://github.com/axios/axios/issues/815
      if (error.request.responseType === 'blob') {
        return new Promise((resolve, reject) => {
            let reader = new FileReader();
            reader.onload = () => {
                error.response.data = JSON.parse(String(reader.result))

                appStore.apiErr = { 
                  method: error.config.method.toUpperCase(), 
                  url: error.config.url, 
                  errMsg: error.response.status + ' - ' + error.response.data.err_description 
                }

                resolve(Promise.reject(error))
            }

            reader.onerror = () => {
                reject(error)
            }
            reader.readAsText(error.response.data)
        })
      }

      // don't use appStore for urls that have special error handling
      if (error.config.url === '/session-token-login') {
        return Promise.reject(error)
      }

      // regular requests

      // if no response is received from server, error.response is undefined.
      const errMsg = error.response ? (error.response.status + ' - ' + error.response.data.err_description) : 'No response received from server'

      appStore.apiErr = { 
        method: error.config.method.toUpperCase(), 
        url: error.config.url, 
        errMsg: errMsg
      }

      return Promise.reject(error)
    }
  )

  return instance
}

export function setAuthToken(axInstance: AxiosInstance, token: string) {
  axInstance.defaults.headers.common.Authorization = `Bearer ${token}`
}

export function deleteAuthToken(axInstance: AxiosInstance) {
  delete axInstance.defaults.headers.common.Authorization
}

const ax = createAxiosInstance(import.meta.env.VITE_API_URL)
export const axSupplier = createAxiosInstance(import.meta.env.VITE_SUPPLIER_API_URL)

export default ax