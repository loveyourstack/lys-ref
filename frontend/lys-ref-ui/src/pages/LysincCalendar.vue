<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>
              <span class="dt-title">HR Calendar</span>

              <div class="dt-subtitle mt-6">TODO</div>

              <lysinc-hr-calendar class="mt-4" style="min-width: 900px;" />

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
