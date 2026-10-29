<template>
  <!-- source = referencing (many) column, target = referenced (one) column; markers are defined in CoreSchemaDesigner -->
  <BaseEdge :id="id" :path="path[0]" marker-start="url(#schema-many)" marker-end="url(#schema-one)"
    :style="{ stroke: '#888', strokeWidth: 1.5 }" />
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { BaseEdge, getSmoothStepPath, type EdgeProps } from '@vue-flow/core'

const props = defineProps<EdgeProps>()

const path = computed(() => getSmoothStepPath({
  sourceX: props.sourceX,
  sourceY: props.sourceY,
  sourcePosition: props.sourcePosition,
  targetX: props.targetX,
  targetY: props.targetY,
  targetPosition: props.targetPosition,
  borderRadius: 10,
  offset: 30,
}))
</script>
