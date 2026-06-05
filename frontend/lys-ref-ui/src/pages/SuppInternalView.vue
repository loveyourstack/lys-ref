<template>
  <v-container fluid>
    <v-responsive>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <v-row density="comfortable">
                <v-col class="mb-6">
                  <div class="dt-subtitle">This example is based on the classic Northwind database.</div>
                  <div class="dt-subtitle">We are a food trader, and we have suppliers. The supplier employees can log in to manage their products.</div>
                  <div class="dt-subtitle">This page shows supplier employees and products as we would see them in our internal application.</div>
                  <div class="dt-subtitle">The employees table allows a fake login of a supplier employee in order to show data in the Tenant view. The logged-in employee's email is highlighted in bold.</div>
                </v-col>
              </v-row>

              <supp-employee-table :internal="true" />

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <supp-product-table :internal="true" />

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
</script>
