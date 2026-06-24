<template>
  <v-container fluid>
    <v-responsive>
      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">{{ $t('external_data.geo_ip.title') }}</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <div class="dt-subtitle">{{ $t('external_data.geo_ip.p1') }}</div>

              <v-list>
                <v-list-item>
                  <v-list-item-title>{{ $t('external_data.geo_ip.your_ip') }} <span class="v-code">{{ auth.user.ip }}</span></v-list-item-title>
                  <v-list-item-title>
                    {{ $t('external_data.geo_ip.your_location') }} 
                    <span class="v-code">{{ auth.user.geo_ip_location }}</span>
                    <span class="fi ml-1" :class="'fi-' + auth.user.geo_ip_country_iso_code.toLowerCase()" style="height: 1.5rem; padding: 0.2em 0.4em;"></span>
                  </v-list-item-title>
                </v-list-item>
              </v-list>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import auth from '@/auth'
</script>
