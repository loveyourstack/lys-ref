<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>
              <span class="dt-title">{{ $t('calendar.title') }}</span>

              <div class="dt-subtitle mt-6">{{ $t('calendar.p1') }}</div>
              <div class="dt-subtitle">{{ $t('calendar.p2') }}</div>

              <lysinc-hr-calendar class="mt-4" style="min-width: 900px;" />

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
</script>
