<template>
  <span class="text-title-medium">Select profile picture</span>
  
  <v-card-text>
    <v-form>

      <v-file-upload v-model="files" density="compact" clearable outlined inset-file-list min-width="400px"
        label="Profile picture" filter-by-type="image/gif,image/jpeg,image/png" title="Click or drag image here"
        :error-messages="uploadError"
        :rules="[v => !!v || 'Profile picture is required', imageDimensionsRule]"
        @update:model-value="rejectSmallImage"
      ></v-file-upload>

      <cropper-canvas v-if="previewUrl && requiresCropping" class="profile-cropper" background scale-step="0"
        :style="{ width: `${cropperSize.width}px`, height: `${cropperSize.height}px` }">
        
        <cropper-image :src="previewUrl" alt="Profile picture" initial-center-size="contain" rotatable 
          scalable skewable translatable></cropper-image>

        <cropper-shade hidden></cropper-shade>
        <cropper-handle action="select" plain></cropper-handle>

        <cropper-selection ref="cropperSelectionRef" 
          :x="(cropperSize.width - MAX_SIZE) / 2"
          :y="(cropperSize.height - MAX_SIZE) / 2"
          :width="MAX_SIZE"
          :height="MAX_SIZE"
          aspect-ratio="1" movable resizable @change="constrainSelection">

          <cropper-grid role="grid" covered></cropper-grid>
          <cropper-crosshair centered></cropper-crosshair>
          <cropper-handle action="move" theme-color="rgba(255, 255, 255, 0.35)"></cropper-handle>
          <cropper-handle action="n-resize"></cropper-handle>
          <cropper-handle action="e-resize"></cropper-handle>
          <cropper-handle action="s-resize"></cropper-handle>
          <cropper-handle action="w-resize"></cropper-handle>
          <cropper-handle action="ne-resize"></cropper-handle>
          <cropper-handle action="nw-resize"></cropper-handle>
          <cropper-handle action="se-resize"></cropper-handle>
          <cropper-handle action="sw-resize"></cropper-handle>

        </cropper-selection>
      </cropper-canvas>

      <img v-if="croppedPreviewUrl" :src="croppedPreviewUrl" alt="Cropped profile picture" class="cropped-preview">

      <div class="d-flex align-center ga-6">
        <v-btn icon class="" @click="$emit('cancelled')">
          <v-icon icon="mdi-arrow-left"></v-icon>
        </v-btn>

        <v-btn v-if="requiresCropping" color="secondary" :disabled="!cropperSelectionRef" @click="crop">Crop</v-btn>
        <v-btn color="primary" :disabled="!croppedFile" @click="save">Save</v-btn>
      </div>

    </v-form>
  </v-card-text>

</template>

<script setup lang="ts">
import { computed, ref, onBeforeUnmount, watch } from 'vue'
// register the cropper-* custom elements
import 'cropperjs'
import type { CropperSelection } from 'cropperjs'

const cropperSize = ref({ width: 400, height: 400 })

const emit = defineEmits<{
  (e: 'cancelled'): void
  (e: 'cropped', file: File): void
}>()

const files = ref<File[]>([])
const previewUrl = ref<string>()
const cropperSelectionRef = ref<CropperSelection>()
const croppedFile = ref<File>()
const croppedPreviewUrl = ref<string>()

const MAX_SIZE = 400
const uploadError = ref<string>()
const requiresCropping = computed(() =>
  cropperSize.value.width !== MAX_SIZE || cropperSize.value.height !== MAX_SIZE,
)

type SelectionChange = CustomEvent<{
  x: number
  y: number
  width: number
  height: number
}>

function clearCroppedImage() {
  if (croppedPreviewUrl.value) URL.revokeObjectURL(croppedPreviewUrl.value)
  croppedPreviewUrl.value = undefined
  croppedFile.value = undefined
}

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

async function rejectSmallImage(selectedFiles: File[]) {
  const file = selectedFiles[0]
  uploadError.value = undefined

  if (!file) return

  const dimensions = await getImageDimensions(file)
  if (!dimensions || dimensions.width < MAX_SIZE || dimensions.height < MAX_SIZE) {
    uploadError.value = 'Profile picture must be at least 400 x 400 pixels'
    files.value = []
  }
}

function save() {
  if (croppedFile.value) emit('cropped', croppedFile.value)
}

watch(files, async (newFiles, _old, onCleanup) => {
  clearCroppedImage()
  const file = newFiles[0]
  if (!file) {
    previewUrl.value = undefined
    return
  }

  const dimensions = await getImageDimensions(file)
  if (!dimensions || dimensions.width < MAX_SIZE || dimensions.height < MAX_SIZE) {
    previewUrl.value = undefined
    return
  }

  const objectUrl = URL.createObjectURL(file)
  cropperSize.value = dimensions
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