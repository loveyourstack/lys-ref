<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <v-row density="comfortable">
                <v-col class="mb-6">
                  <div class="dt-subtitle">{{ $t('saas.tenant_view.p1') }}</div>
                  <div class="dt-subtitle">{{ $t('saas.tenant_view.p2') }}</div>
                  <div class="dt-subtitle">{{ $t('saas.tenant_view.p3') }}</div>
                </v-col>
              </v-row>

              <div v-if="suppStore.selectedEmpEmail" id="logged-in-email"><v-icon>mdi-account</v-icon> {{ suppStore.selectedEmpEmail }}</div>

              <supp-employee-table v-if="suppStore.selectedEmpEmail" :internal="false" />
              <div v-else>
                <div class="dt-title">{{ $t('saas.supplier_employees') }}</div>
                <div class="dt-subtitle">{{ $t('saas.tenant_view.please_login') }}</div>
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
                <div class="dt-title">{{ $t('saas.supplier_products') }}</div>
                <div class="dt-subtitle">{{ $t('saas.tenant_view.please_login') }}</div>
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