<template>
  <v-list density="compact" class="nav-list nav-list-right">
    <v-list-subheader :title="$t('nav_header.external_data')" class="mt-2 clickable" @click="showExtDataItems = !showExtDataItems"></v-list-subheader>

    <div v-if="showExtDataItems">
      <v-list-item link :title="$t('currencies.nav')" to="/ecb/currencies" prepend-icon="mdi-currency-eur"></v-list-item>
      <v-list-item link :title="$t('exchange_rates.nav')" to="/ecb/exchange-rates" prepend-icon="mdi-currency-eur"></v-list-item>
      <v-list-item link :title="$t('geo_ip.nav')" to="/maxmind/geo-ip" prepend-icon="mdi-earth"></v-list-item>
    </div>

    <v-list-subheader :title="$t('nav_header.charts')" class="mt-2 clickable" @click="showChartsItems = !showChartsItems"></v-list-subheader>

    <div v-if="showChartsItems">
      <v-list-item link :title="$t('xr_performance.nav')" to="/ecb/xr-performance" prepend-icon="mdi-chart-bell-curve-cumulative"></v-list-item>
      <v-list-item link :title="$t('campaign_charts.nav')" to="/digital-marketing/campaign-charts" prepend-icon="mdi-chart-pie"></v-list-item>
    </div>

    <v-list-subheader :title="$t('nav_header.parallel_processing')" class="mt-2 clickable" @click="showParProcItems = !showParProcItems"></v-list-subheader>

    <div v-if="showParProcItems">
      <v-list-item link :title="$t('flows.nav')" to="/process/flows" prepend-icon="mdi-sitemap-outline"></v-list-item>
      <v-list-item link :title="$t('runs.nav')" to="/process/runs" prepend-icon="mdi-repeat"></v-list-item>
    </div>

    <v-list-subheader :title="$t('nav_header.saas')" class="mt-2 clickable" @click="showMultiTenItems = !showMultiTenItems"></v-list-subheader>

    <div v-if="showMultiTenItems">
      <v-list-item link :title="$t('internal_view.nav')" to="/supplier/internal-view" prepend-icon="mdi-home-city-outline"></v-list-item>
      <v-list-item link :title="$t('tenant_view.nav')" to="/supplier/tenant-view" prepend-icon="mdi-web"></v-list-item>
    </div>

    <v-list-subheader :title="$t('nav_header.server_push')" class="mt-2 clickable" @click="showServerPushItems = !showServerPushItems"></v-list-subheader>

    <div v-if="showServerPushItems">
      <v-list-item link :title="$t('notifications.nav')" to="/server-push/notifications" prepend-icon="mdi-bell-outline"></v-list-item>
    </div>

    <v-list-subheader v-if="auth.hasRole(Role.Tech)" :title="$t('nav_header.monitoring')" class="mt-2 clickable" @click="showMonitoringItems = !showMonitoringItems"></v-list-subheader>

    <div v-if="auth.hasRole(Role.Tech) && showMonitoringItems">
      <v-list-item link :title="$t('application.nav')" to="/monitoring/application" prepend-icon="mdi-application"></v-list-item>
      <v-list-item link :title="$t('database.nav')" to="/monitoring/database" prepend-icon="mdi-database-eye-outline"></v-list-item>
    </div>

  </v-list>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useJsonLs } from 'lys-vue'
import auth from '@/auth'
import { Role } from '@/types/system'

const showExtDataItems = ref(true)
const showChartsItems = ref(true)
const showParProcItems = ref(true)
const showMultiTenItems = ref(true)
const showServerPushItems = ref(true)
const showMonitoringItems = ref(true)

useJsonLs({
  lsKey: 'right_nav_list',
  refs: {
    showExtDataItems,
    showChartsItems,
    showParProcItems,
    showMultiTenItems,
    showServerPushItems,
    showMonitoringItems,
  },
})

</script>
