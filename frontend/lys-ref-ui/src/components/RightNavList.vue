<template>
  <v-list density="compact" class="nav-list nav-list-right">
    <v-list-subheader :title="$t('right_nav.external_data.title')" class="mt-2 clickable" @click="showExtDataItems = !showExtDataItems"></v-list-subheader>

    <div v-if="showExtDataItems">
      <v-list-item link :title="$t('right_nav.external_data.currencies')" to="/ecb/currencies" prepend-icon="mdi-currency-eur"></v-list-item>
      <v-list-item link :title="$t('right_nav.external_data.exchange_rates')" to="/ecb/exchange-rates" prepend-icon="mdi-currency-eur"></v-list-item>
      <v-list-item link :title="$t('right_nav.external_data.geo_ip')" to="/maxmind/geo-ip" prepend-icon="mdi-earth"></v-list-item>
    </div>

    <v-list-subheader :title="$t('right_nav.charts.title')" class="mt-2 clickable" @click="showChartsItems = !showChartsItems"></v-list-subheader>

    <div v-if="showChartsItems">
      <v-list-item link :title="$t('right_nav.charts.xr_performance')" to="/ecb/xr-performance" prepend-icon="mdi-chart-bell-curve-cumulative"></v-list-item>
      <v-list-item link :title="$t('right_nav.charts.campaign_charts')" to="/digital-marketing/campaign-charts" prepend-icon="mdi-chart-pie"></v-list-item>
    </div>

    <v-list-subheader :title="$t('right_nav.parallel_processing.title')" class="mt-2 clickable" @click="showParProcItems = !showParProcItems"></v-list-subheader>

    <div v-if="showParProcItems">
      <v-list-item link :title="$t('right_nav.parallel_processing.flows')" to="/process/flows" prepend-icon="mdi-sitemap-outline"></v-list-item>
      <v-list-item link :title="$t('right_nav.parallel_processing.runs')" to="/process/runs" prepend-icon="mdi-repeat"></v-list-item>
    </div>

    <v-list-subheader :title="$t('right_nav.saas.title')" class="mt-2 clickable" @click="showMultiTenItems = !showMultiTenItems"></v-list-subheader>

    <div v-if="showMultiTenItems">
      <v-list-item link :title="$t('right_nav.saas.internal_view')" to="/supplier/internal-view" prepend-icon="mdi-home-city-outline"></v-list-item>
      <v-list-item link :title="$t('right_nav.saas.tenant_view')" to="/supplier/tenant-view" prepend-icon="mdi-web"></v-list-item>
    </div>

    <v-list-subheader :title="$t('right_nav.server_push.title')" class="mt-2 clickable" @click="showServerPushItems = !showServerPushItems"></v-list-subheader>

    <div v-if="showServerPushItems">
      <v-list-item link :title="$t('right_nav.server_push.notifications')" to="/server-push/notifications" prepend-icon="mdi-bell-outline"></v-list-item>
    </div>

    <v-list-subheader v-if="auth.hasRole(Role.Tech)" :title="$t('right_nav.monitoring.title')" class="mt-2 clickable" @click="showMonitoringItems = !showMonitoringItems"></v-list-subheader>

    <div v-if="auth.hasRole(Role.Tech) && showMonitoringItems">
      <v-list-item link :title="$t('right_nav.monitoring.application')" to="/monitoring/application" prepend-icon="mdi-application"></v-list-item>
      <v-list-item link :title="$t('right_nav.monitoring.database')" to="/monitoring/database" prepend-icon="mdi-database-eye-outline"></v-list-item>
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
