<template>
  <v-container fluid>
    <v-responsive>
      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">{{ $t('uploads.title') }}</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text>

              <div class="dt-subtitle">{{ $t('uploads.p1') }}</div>
              <div class="dt-subtitle">{{ $t('uploads.p2') }}</div>
              <div class="dt-subtitle">{{ $t('uploads.p3') }}</div>
              <div class="dt-subtitle">{{ $t('uploads.p4') }}</div>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pt-0">

              <v-container>
                <v-row>
                  <v-col>
                    <v-list width="500px" v-model:selected="selectedConstraintValue" :items="constraintItems" select-strategy="single-independent" mandatory
                      @update:selected="files = []; uploadResults = []" :disabled="uploading">
                    </v-list>
                  </v-col>
                  <v-col>
                    <v-form>
                      <v-file-upload v-model="files" clearable density="compact" width="500px" inset-file-list
                        :filter-by-type="selectedConstraint?.mime_type" :title="$t('uploads.control_title')" :loading="uploading" :disabled="uploading"></v-file-upload>
                      <v-btn :disabled="!auth.isWriter() || !files.length || uploading" size="large" color="primary" @click="uploadFiles"
                        :loading="uploading">
                        <v-icon class="mr-1">mdi-upload</v-icon>{{ $t('actions.upload') }}
                      </v-btn>
                    </v-form>
                  </v-col>
                </v-row>
              </v-container>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row density="compact">
        <v-col cols="auto">
          <v-card variant="flat" class="mt-2">
            <v-card-title>{{ $t('uploads.results') }}</v-card-title>
            <v-card-text>
              <span v-if="!uploadResults.length" class="dt-subtitle">{{ $t('uploads.results_placeholder') }}</span>

              <v-table v-else class="response-table">
                <thead>
                  <tr>
                    <th>Original filename</th>
                    <th>MIME type</th>
                    <th>Size (bytes)</th>
                    <th>Stored name</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(res, idx) in uploadResults" :key="idx">
                    <td>{{ res.original_name }}</td>
                    <td>{{ res.mime_type }}</td>
                    <td>{{ res.size_bytes }}</td>
                    <td>{{ res.stored_name }}</td>
                  </tr>
                </tbody>
              </v-table>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="auto">
          <v-card variant="flat" class="mt-2">
            <v-card-title>{{ $t('uploads.test_files') }}</v-card-title>
            <v-card-text>
              <v-btn size="large" color="secondary" @click="downloadTestFiles"><v-icon class="mr-1">mdi-download</v-icon>{{ $t('uploads.test_files_download') }}</v-btn>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ax from '@/api'
import auth from '@/auth'
import { type Constraint, type UploadResult } from '@/types/uploads'

const s3Bucket = import.meta.env.VITE_S3_BUCKET

const baseUrl = '/a/uploads'

const files = ref<File[]>([])
const uploadResults = ref<UploadResult[]>([])
const uploading = ref(false)

const { t } = useI18n()

// computed: so that the titles are reactive to language changes
const constraintItems = computed<Constraint[]>(() => [
  { value: 1, title: t('uploads.constraint1'), path: '/single-text-file', mime_type: 'text/plain' },
  { value: 2, title: t('uploads.constraint2'), path: '/multiple-text-files', mime_type: 'text/plain' },
  { value: 3, title: t('uploads.constraint3'), path: '/image-file', mime_type: 'image/png' },
])
const selectedConstraintValue = ref([1])
const selectedConstraint = computed(() => {
  return constraintItems.value.find(item => item.value === selectedConstraintValue.value[0]) || null
})

function downloadTestFiles() {

  // requires "content-disposition: attachment" to be added to the S3 object system-defined metadata
  // otherwise the browser will try to open the file as a link instead of downloading it
  const fileUrl = `${s3Bucket}${encodeURIComponent('upload_test_files.7z')}`
  window.location.assign(fileUrl)
}

function uploadFiles() {
  if (files.value.length === 0) { return }

  const constraint = selectedConstraint.value
  if (!constraint) { return }

  uploading.value = true

  const formData = new FormData()
  files.value.forEach((file) => {
    formData.append('files', file)
  })
  ax.post(`${baseUrl}` + constraint.path, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
    .then((response) => {
      uploadResults.value = response.data.data.file_results
      files.value = []  // Clear the file input after successful upload
    })
    .catch((error) => {
      console.error('Upload failed:', error)
    })
    .finally(() => {
      uploading.value = false
    })
}

</script>

<style scoped>
.response-table { min-width: 500px; max-width: 100%; }
@media (max-width: 600px) { 
  .response-table { min-width: 0; width: 100%; }
}
</style>