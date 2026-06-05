<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <v-row density="comfortable">
                <v-col class="mb-6">
                  <div class="dt-subtitle">This page shows supplier employees and products as the logged-in supplier employee would see them.</div>
                  <div class="dt-subtitle">Suppliers have a separate backend API (and normally a separate UI) so that endpoints and deployments are separated from the internal application.</div>
                  <div class="dt-subtitle">Tenant isolation (i.e. each supplier can only see his own data) is enforced using database row-level security.</div>
                </v-col>
              </v-row>

              <div v-if="suppStore.selectedEmpEmail" id="logged-in-email"><v-icon>mdi-account</v-icon> {{ suppStore.selectedEmpEmail }}</div>

              <supp-employee-table v-if="suppStore.selectedEmpEmail" :internal="false" />
              <div v-else>
                <div class="dt-title">Supplier employees</div>
                <div class="dt-subtitle">Please login as an employee on the Internal view page.</div>
              </div>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <supp-product-table v-if="suppStore.selectedEmpEmail" :internal="false" />
              <div v-else>
                <div class="dt-title">Supplier products</div>
                <div class="dt-subtitle">Please login as an employee on the Internal view page.</div>
              </div>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { useSupplierStore } from '@/stores/supplier'

const suppStore = useSupplierStore()

</script>

<style scoped>
#logged-in-email {
  font-size: 1.25rem;
  color: rgb(var(--v-theme-primary));
  margin-bottom: 2rem;
}
</style>