<template>
  <v-app class="rounded rounded-md">
    <v-app-bar density="compact" elevation="8" class="app-shell-bar">

      <v-app-bar-nav-icon variant="text" v-tooltip:bottom="'Toggle left menu'" @click.stop="showLeftNav = !showLeftNav"></v-app-bar-nav-icon>

      <v-img max-height="30px" max-width="30px" src="./../assets/logo.png" class="ml-1"></v-img>
      <v-toolbar-title>
        <span class="font-weight-bold">{{ appStore.company }} - {{ appStore.projectTitle }}</span>
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <notification-menu />

      <v-menu :close-on-content-click="false">
        <template v-slot:activator="{ props }">
          <div v-if="auth.user" class="d-flex align-center" v-bind="props">
            <v-icon icon="mdi-account" color="primary" class="mr-1"></v-icon>
            <div class="text-body-1 mr-3">{{ auth.user.name }}</div>
          </div>
        </template>

        <v-list density="compact" class="mt-1">
          <v-list-item prepend-icon="mdi-logout" class="clickable" @click="logout()">
            <v-list-item-title>Logout</v-list-item-title>
          </v-list-item>
        </v-list>

      </v-menu>

      <v-menu :close-on-content-click="false">
        <template v-slot:activator="{ props }">
          <v-btn icon="mdi-dots-vertical" flat v-bind="props"></v-btn>
        </template>

        <v-card>
          <v-list density="compact">

            <v-list-item>
              <v-switch label="Dark mode" color="primary" v-model="darkMode" hide-details></v-switch>
            </v-list-item>

            <v-list-item v-if="auth.user.has_aws_sg_rules" class="clickable" @click="updateAwsFirewall()">
              <v-btn prepend-icon="mdi-wall-fire" style="letter-spacing: normal !important;" 
                :loading="updatingAwsFirewall">Update AWS firewall</v-btn>
            </v-list-item>

          </v-list>
        </v-card>
      </v-menu>

      <v-app-bar-nav-icon variant="text" v-tooltip:bottom="'Toggle right menu'" @click.stop="showRightNav = !showRightNav"></v-app-bar-nav-icon>
    </v-app-bar>

    <v-navigation-drawer v-model="showLeftNav" elevation="8" floating>
      <left-nav-list />
    </v-navigation-drawer>

    <v-navigation-drawer location="right" v-model="showRightNav" elevation="8" floating>
      <right-nav-list />
    </v-navigation-drawer>

    <v-main class="d-flex">
      <api-error />
      <router-view />
    </v-main>
  </v-app>
</template>

<script lang="ts" setup>
import { ref, watch, onBeforeMount, onMounted } from 'vue'
import { useTheme } from 'vuetify'
import { useRouter } from 'vue-router'
import { fetchOnce, notify } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { useAppStore } from '@/stores/app'
import { useCoreStore } from '@/stores/core'
import { useDigmarkStore } from '@/stores/digmark'
import { useEcbStore } from '@/stores/ecb'
import { useGeoStore } from '@/stores/geo'
import { useProcessStore } from '@/stores/process'
import { usePublisherStore } from '@/stores/publisher'
import { useSupplierStore } from '@/stores/supplier'
import { type StoreData } from '@/types/system'

const theme = useTheme()
const router = useRouter()

const appStore = useAppStore()
const coreStore = useCoreStore()
const digmarkStore = useDigmarkStore()
const ecbStore = useEcbStore()
const geoStore = useGeoStore()
const processStore = useProcessStore()
const pubStore = usePublisherStore()
const suppStore = useSupplierStore()

const darkMode = ref(false)

const storeData = ref<StoreData>()

const showLeftNav = ref(true)
const showRightNav = ref(true)

const updatingAwsFirewall = ref(false)

const lsKey = 'main'

function loadStoreData() {

  // load store data in single API call: prevents rate limit issues and doesn't spam server request log

  fetchOnce({ ax, myUrl: '/a/system/ui-store-data', result: storeData, onSuccess: () => {
    if (!storeData.value) {
      return
    }

    coreStore.mandatoryEnums = storeData.value.core_mandatory_enums
    coreStore.optionalEnums = storeData.value.core_optional_enums
    coreStore.periods = storeData.value.core_periods

    digmarkStore.managers = storeData.value.digmark_managers
    digmarkStore.verticals = storeData.value.digmark_verticals

    ecbStore.activeCurrenciesExEur = storeData.value.ecb_active_currencies_ex_eur

    geoStore.countries = storeData.value.geo_countries
    geoStore.oceans = storeData.value.geo_oceans

    processStore.flows = storeData.value.process_flows

    pubStore.authors = storeData.value.pub_authors

    suppStore.companies = storeData.value.supp_companies
    suppStore.productCategories = storeData.value.supp_product_categories
  }})

}

function logout() {
  // make post call to remove session from server
  var myURL = '/a/logout'
  ax.post(myURL)
    .then(() => {
      appStore.apiErr = undefined
    })
    .catch() // handled by interceptor
    .finally(() => {
      auth.logout()
      router.push({ path: '/login' })
    })
}

function updateAwsFirewall() {
  updatingAwsFirewall.value = true
  ax.patch('/a/aws/update-user-security-group-rules')
    .then(resp => {
      notify(appStore.company, resp.data.data, appStore.logoUrl)
    })
    .catch() // handled by interceptor
    .finally(() => updatingAwsFirewall.value = false )
}

watch(darkMode, (newVal) => {
  if (newVal) { theme.change('dark') }
  else { theme.change('light') }
})

watch([darkMode, showRightNav], () => {

  let lsObj = {
    'darkMode': darkMode.value,
    'showRightNav': showRightNav.value,
  }
  localStorage.setItem(lsKey, JSON.stringify(lsObj))
})

onBeforeMount(() => {
  const lsJSON = localStorage.getItem(lsKey)
  if (!lsJSON) {
    return
  }

  const lsObj = JSON.parse(lsJSON)
  if (lsObj['darkMode'] !== undefined) { darkMode.value = lsObj['darkMode'] }
  if (lsObj['showRightNav'] !== undefined) { showRightNav.value = lsObj['showRightNav'] }
})

onMounted(() => {
  loadStoreData()
})
</script>
