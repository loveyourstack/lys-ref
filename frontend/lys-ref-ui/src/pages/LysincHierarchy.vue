<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>
              <span class="dt-title">Hierarchical data: Organization</span>

              <div class="dt-subtitle mt-6">The employees of a fictional organization are shown below. Names and pictures were generated.</div>
              <div class="dt-subtitle">The data is hierarchical: other than the CEO at the top, each employee has a manager. Each manager's employees are shown as child nodes in the tree view.</div>
              <div class="dt-subtitle">The tree nodes can be expanded and collapsed by clicking on them. Selecting an employee using the checkbox shows their profile on the right.</div>

              <v-row class="ga-8 mt-4">
                <v-col cols="auto">
                  <employee-treeview @selected="selectedEmployee = $event" />
                </v-col>
                <v-col>
                  <employee-profile :employee="selectedEmployee" />
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
