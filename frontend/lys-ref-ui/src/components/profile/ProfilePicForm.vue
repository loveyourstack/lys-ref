<template>
  <v-form ref="formRef">

    <v-file-upload v-model="files" density="compact" clearable outlined inset-file-list min-width="400px" max-width="800px"
      label="Profile picture" filter-by-type="image/gif,image/jpeg,image/png" :title="$t('image_cropping.control_title')"
      :error-messages="uploadError"
      :rules="[v => !!v || 'Profile picture is required', imageDimensionsRule, imageFileSizeRule]"
      @update:model-value="validateImage"
    ></v-file-upload>

    <!-- scale-step="0" prevents zooming with mouse wheel -->
    <cropper-canvas v-if="previewUrl && requiresCropping" class="profile-cropper" background scale-step="0"
      :style="{ width: `${cropperSize.width}px`, height: `${cropperSize.height}px` }">
      
      <cropper-image :src="previewUrl" alt="Profile picture" initial-center-size="contain"
      ></cropper-image>

      <cropper-shade hidden></cropper-shade>

      <cropper-selection ref="cropperSelectionRef" 
        :x="(cropperSize.width - MAX_SIZE) / 2"
        :y="(cropperSize.height - MAX_SIZE) / 2"
        :width="MAX_SIZE"
        :height="MAX_SIZE"
        aspect-ratio="1" movable @change="constrainSelection">

        <cropper-grid role="grid" covered></cropper-grid>
        <cropper-crosshair centered></cropper-crosshair>
        <cropper-handle action="move" theme-color="rgba(255, 255, 255, 0.35)"></cropper-handle>

      </cropper-selection>
    </cropper-canvas>

    <!-- show the cropped image if available -->
    <img v-if="croppedPreviewUrl" :src="croppedPreviewUrl" alt="Cropped profile picture" class="cropped-preview">

    <div class="d-flex align-center ga-6">
      <v-btn v-if="requiresCropping" color="secondary" :disabled="!cropperSelectionRef" @click="crop">{{ $t('actions.crop') }}</v-btn>
      
      <v-btn color="primary" :disabled="!auth.isWriter() || !croppedFile" :loading="saving" @click="save">{{ $t('actions.save') }}</v-btn>

      <v-fade-transition mode="out-in">
        <v-chip v-if="showSaved" color="success">{{ $t('feedback.saved') }}</v-chip>
      </v-fade-transition>
    </div>

  </v-form>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onBeforeUnmount, watch } from 'vue'
import ax from '@/api'
import auth from '@/auth'
import { type SelectionChange } from '@/types/uploads'

// register the cropper-* custom elements
import 'cropperjs'
import type { CropperSelection } from 'cropperjs'

const baseUrl = '/a/system/set-user-profile-pic'
const formRef = ref<{ resetValidation: () => void }>()

// cropperSize is the size of the initial uploaded image, which may be larger than 400x400. It is set in the files watch handler
const cropperSize = ref({ width: 400, height: 400 })

const files = ref<File[]>([])

// previewUrl is the temp URL of the initial uploaded image. Is set in the files watch handler
const previewUrl = ref<string>()

// cropperSelectionRef is a ref to the cropper-selection element, which is used to get the cropped image
const cropperSelectionRef = ref<CropperSelection>()

// croppedFile is the file that will be uploaded to the server. It is set when the user clicks the "Crop" button,
// or when the initial uploaded image is exactly 400x400 and no cropping is required
const croppedFile = ref<File>()

// croppedPreviewUrl is the temp URL of the cropped image. It is set when the user clicks the "Crop" button
const croppedPreviewUrl = ref<string>()

// MAX_SIZE is the height and width of the image that will be uploaded to the server
const MAX_SIZE = 400

const MAX_FILE_SIZE = 2 * 1024 * 1024

const uploadError = ref<string>()

const requiresCropping = computed(() =>
  cropperSize.value.width !== MAX_SIZE || cropperSize.value.height !== MAX_SIZE,
)

const saving = ref(false)
const showSaved = ref(false)

function clearCroppedImage() {
  if (croppedPreviewUrl.value) URL.revokeObjectURL(croppedPreviewUrl.value)
  croppedPreviewUrl.value = undefined
  croppedFile.value = undefined
}

// constrainSelection ensures that the cropper selection box stays within the bounds of the image
function constrainSelection(event: SelectionChange) {
  clearCroppedImage()

  const { x, y, width, height } = event.detail
  const maximumSize = Math.min(cropperSize.value.width, cropperSize.value.height)
  const size = Math.min(width, height, maximumSize)

  const boundedX = Math.min(Math.max(x, 0), cropperSize.value.width - size)
  const boundedY = Math.min(Math.max(y, 0), cropperSize.value.height - size)

  if (x === boundedX && y === boundedY && width === size && height === size) {
    return
  }

  event.preventDefault()
  cropperSelectionRef.value?.$change(boundedX, boundedY, size, size, 1)
}

// crop gets the cropped image from the cropper-selection element and sets it as the croppedFile and croppedPreviewUrl
async function crop() {
  const file = files.value[0]
  const selection = cropperSelectionRef.value
  if (!file || !selection) return

  const canvas = await selection.$toCanvas({ width: MAX_SIZE, height: MAX_SIZE })
  const blob: Blob | null = await new Promise(resolve => canvas.toBlob(resolve, file.type))
  if (blob) {
    clearCroppedImage()
    croppedFile.value = new File([blob], file.name, { type: file.type })
    croppedPreviewUrl.value = URL.createObjectURL(croppedFile.value)
  }
}

async function getImageDimensions(file: File) {
  const objectUrl = URL.createObjectURL(file)
  const image = new Image()
  image.src = objectUrl

  try {
    await image.decode()
    return { width: image.naturalWidth, height: image.naturalHeight }
  } catch {
    return undefined
  } finally {
    URL.revokeObjectURL(objectUrl)
  }
}

async function imageDimensionsRule(selectedFiles: File[]) {
  const file = selectedFiles[0]
  if (!file) return true

  const dimensions = await getImageDimensions(file)
  return !dimensions || (dimensions.width >= MAX_SIZE && dimensions.height >= MAX_SIZE)
    || 'Profile picture must be at least 400 x 400 pixels'
}

function imageFileSizeRule(selectedFiles: File[]) {
  const file = selectedFiles[0]
  return !file || file.size <= MAX_FILE_SIZE || 'Profile picture must be 2 MB or smaller'
}

function save() {
  if (!croppedFile.value) { return }

  saving.value = true

  const formData = new FormData()
  formData.append('files', croppedFile.value)

  ax.patch(baseUrl, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
    .then(async () => {
      showSaved.value = true
      setTimeout(() => { showSaved.value = false }, import.meta.env.VITE_FADE_MS)

      croppedFile.value = undefined
      files.value = []  // Clear the file input after successful upload
      uploadError.value = undefined

      await nextTick()
      formRef.value?.resetValidation()
    })
    .catch((error) => {
      console.error('Upload failed:', error)
    })
    .finally(() => {
      saving.value = false
    })
}

async function validateImage(selectedFiles: File[]) {
  const file = selectedFiles[0]
  uploadError.value = undefined

  if (!file) return

  if (file.size > MAX_FILE_SIZE) {
    uploadError.value = 'Profile picture must be 2 MB or smaller'
    files.value = []
    return
  }

  const dimensions = await getImageDimensions(file)
  if (!dimensions || dimensions.width < MAX_SIZE || dimensions.height < MAX_SIZE) {
    uploadError.value = 'Profile picture must be at least 400 x 400 pixels'
    files.value = []
  }
}

watch(files, async (newFiles, _old, onCleanup) => {
  clearCroppedImage()
  const file = newFiles[0]
  if (!file) {
    previewUrl.value = undefined
    return
  }

  const dimensions = await getImageDimensions(file)

  // ensure that the image is at least MAX_SIZE x MAX_SIZE
  if (!dimensions || dimensions.width < MAX_SIZE || dimensions.height < MAX_SIZE) {
    previewUrl.value = undefined
    return
  }

  const objectUrl = URL.createObjectURL(file)
  cropperSize.value = dimensions

  // if the image is exactly MAX_SIZE x MAX_SIZE, no cropping is required and we can use the original file
  if (!requiresCropping.value) croppedFile.value = file

  previewUrl.value = objectUrl

  onCleanup(() => URL.revokeObjectURL(objectUrl))
})

onBeforeUnmount(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  clearCroppedImage()
})

</script>

<style scoped>
.profile-cropper {
  display: block;
  margin-bottom: 1rem;
}

.cropped-preview {
  display: block;
  width: 400px;
  height: 400px;
  margin-bottom: 1rem;
}
</style>