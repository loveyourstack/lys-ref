<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>
              <span class="dt-title">{{ $t('organization.title') }}</span>

              <div class="dt-subtitle mt-6">{{ $t('organization.p1') }}</div>
              <div class="dt-subtitle">{{ $t('organization.p2') }}</div>
              <div class="dt-subtitle">{{ $t('organization.p3') }}</div>

              <v-row class="ga-8 mt-4">
                <v-col cols="auto">
                  <lysinc-emp-treeview @selected="(e: Employee | undefined) => selectedEmployee = e" />
                </v-col>
                <v-col>
                  <lysinc-emp-profile :employee="selectedEmployee" />
                </v-col>
              </v-row>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { type Employee } from '@/types/lysinc'

const selectedEmployee = ref<Employee | undefined>(undefined)

</script>
