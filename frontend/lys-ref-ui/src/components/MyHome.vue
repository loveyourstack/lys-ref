<template>
  <v-container class="fill-height d-flex flex-column justify-center" max-width="1200">
    <v-img class="mb-4" height="150" src="@/assets/logo.png" />

    <div class="mb-8 text-center">
      <div class="text-body-medium font-weight-light mb-n1">{{ $t("home.welcome") }}</div>
      <div class="text-display-medium font-weight-bold">LoveYourStack</div>
    </div>

    <v-row>

      <v-col cols="12">
        <HomeProjectStack />
      </v-col>

      <v-col cols="12" md="6">
        <HomeTypeHandling />
      </v-col>

      <v-col cols="12" md="6">
        <HomeExternalData />
      </v-col>

      <v-col cols="12" md="6">
        <HomeEntityRelationships />
      </v-col>

      <v-col cols="12" md="6">
        <HomeCharts />
      </v-col>

      <v-col cols="12" md="6">
        <HomeAdvancedTables />
      </v-col>

      <v-col cols="12" md="6">
        <HomeParallelProcessing />
      </v-col>

      <v-col cols="12" md="6">
        <HomeAi />
      </v-col>

      <v-col cols="12" md="6">
        <HomeSaaS />
      </v-col>

      <v-col cols="12" md="6">
        <HomeUserDataRetention />
      </v-col>

      <v-col cols="12" md="6">
        <HomeServerPush />
      </v-col>

    </v-row>
  </v-container>
</template>

<script setup lang="ts">
</script>
