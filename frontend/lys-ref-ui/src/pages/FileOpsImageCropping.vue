<template>
  <v-container fluid>
    <v-responsive>
      <v-row density="compact" class="mt-2">
        <v-col cols="auto">
          <v-card variant="flat">
            <v-card-text class="pb-0">
              <span class="dt-title">{{ $t('image_cropping.title') }}</span>

              <i18n-t scope="global" keypath="image_cropping.p1" tag="div" class="dt-subtitle mt-6">
                <template #CropperJs>
                  <a href="https://github.com/fengyuanchen/cropperjs" target="_blank" rel="noopener noreferrer">Cropper.js</a> 
                </template>
              </i18n-t>

              <div class="dt-subtitle">{{ $t('image_cropping.p2') }}</div>
              <div class="dt-subtitle">{{ $t('image_cropping.p3') }}</div>
              <div class="dt-subtitle">{{ $t('image_cropping.p4') }}</div>

              <profile-pic-form class="mt-6"></profile-pic-form>

            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

    </v-responsive>
  </v-container>
</template>

<script lang="ts" setup>
</script>

