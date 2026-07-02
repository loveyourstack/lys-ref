<template v-slot:[`item.step`]="{ item }">
  <v-rating v-model="props.step" :length="props.maxSteps" readonly density="comfortable" class="ga-1">
    <template v-slot:item="slotProps">
      <v-icon :color="slotProps.isFilled ? 'primary' : 'grey-lighten-1'">
        {{ slotProps.isFilled ? 'mdi-circle' : 'mdi-circle-outline' }}
      </v-icon>
    </template>
  </v-rating>
</template>

<script lang="ts" setup>
const props = defineProps<{
  step: number | undefined
  maxSteps: number
}>()

</script>
