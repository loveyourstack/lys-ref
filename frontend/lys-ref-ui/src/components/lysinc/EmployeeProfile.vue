<template>
  <div v-if="props.employee">
    <div class="d-flex justify-center">
      <v-avatar v-if="props.employee.profile_pic" :image="`${s3Bucket}profiles/${props.employee.profile_pic}`" size="400"></v-avatar>
      <v-avatar v-else :image="`${s3Bucket}profiles/generic_profile_400x400.png`" size="400"></v-avatar>
    </div>

    <div class="text-headline-small mt-4 text-center">{{ props.employee.honorific }} {{ props.employee.full_name }}</div>
    <div class="text-body-large text-center">{{ props.employee.job_title }}</div>
    <div class="text-body-small text-center">{{ props.employee.department }}</div>

    <div class="text-body-medium mt-2 text-center"><v-icon>mdi-email</v-icon> {{ props.employee.email }}</div>

    <div v-if="props.employee.reports_to !== props.employee.id" class="text-body-large text-center mt-4">
      Reports to: {{ props.employee.reports_to_full_name }} ({{ props.employee.reports_to_job_title }})
    </div>

  </div>
</template>

<script setup lang="ts">
import { type Employee } from '@/types/lysinc'

const s3Bucket = import.meta.env.VITE_S3_BUCKET

const props = defineProps<{
  employee: Employee | undefined
}>()

</script>
