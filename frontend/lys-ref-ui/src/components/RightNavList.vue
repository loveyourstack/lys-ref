<template>
  <v-list density="compact" class="nav-list nav-list-right">
    <v-list-subheader :title="$t('external_data.nav')" class="mt-2 clickable" @click="showExtDataItems = !showExtDataItems"></v-list-subheader>

    <div v-if="showExtDataItems">
      <v-list-item link :title="$t('external_data.currencies.nav')" to="/ecb/currencies" prepend-icon="mdi-currency-eur"></v-list-item>
      <v-list-item link :title="$t('external_data.exchange_rates.nav')" to="/ecb/exchange-rates" prepend-icon="mdi-currency-eur"></v-list-item>
      <v-list-item link :title="$t('external_data.geo_ip.nav')" to="/maxmind/geo-ip" prepend-icon="mdi-earth"></v-list-item>
    </div>

    <v-list-subheader :title="$t('charts.nav')" class="mt-2 clickable" @click="showChartsItems = !showChartsItems"></v-list-subheader>

    <div v-if="showChartsItems">
      <v-list-item link :title="$t('charts.xr_performance.nav')" to="/ecb/xr-performance" prepend-icon="mdi-chart-bell-curve-cumulative"></v-list-item>
      <v-list-item link :title="$t('charts.campaign_charts.nav')" to="/digital-marketing/campaign-charts" prepend-icon="mdi-chart-pie"></v-list-item>
    </div>

    <v-list-subheader :title="$t('parallel_processing.nav')" class="mt-2 clickable" @click="showParProcItems = !showParProcItems"></v-list-subheader>

    <div v-if="showParProcItems">
      <v-list-item link :title="$t('parallel_processing.flows.nav')" to="/process/flows" prepend-icon="mdi-sitemap-outline"></v-list-item>
      <v-list-item link :title="$t('parallel_processing.runs.nav')" to="/process/runs" prepend-icon="mdi-repeat"></v-list-item>
    </div>

    <v-list-subheader :title="$t('saas.nav')" class="mt-2 clickable" @click="showMultiTenItems = !showMultiTenItems"></v-list-subheader>

    <div v-if="showMultiTenItems">
      <v-list-item link :title="$t('saas.internal_view.nav')" to="/supplier/internal-view" prepend-icon="mdi-home-city-outline"></v-list-item>
      <v-list-item link :title="$t('saas.tenant_view.nav')" to="/supplier/tenant-view" prepend-icon="mdi-web"></v-list-item>
    </div>

    <v-list-subheader :title="$t('server_push.nav')" class="mt-2 clickable" @click="showServerPushItems = !showServerPushItems"></v-list-subheader>

    <div v-if="showServerPushItems">
      <v-list-item link :title="$t('server_push.notifications.nav')" to="/server-push/notifications" prepend-icon="mdi-bell-outline"></v-list-item>
    </div>

    <v-list-subheader v-if="auth.hasRole(Role.Tech)" :title="$t('monitoring.nav')" class="mt-2 clickable" @click="showMonitoringItems = !showMonitoringItems"></v-list-subheader>

    <div v-if="auth.hasRole(Role.Tech) && showMonitoringItems">
      <v-list-item link :title="$t('monitoring.application.nav')" to="/monitoring/application" prepend-icon="mdi-application"></v-list-item>
      <v-list-item link :title="$t('monitoring.database.nav')" to="/monitoring/database" prepend-icon="mdi-database-eye-outline"></v-list-item>
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
