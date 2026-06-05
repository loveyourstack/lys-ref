<template>
  <v-card-title class="pl-1 mb-1">
    {{ formTitle }}
  </v-card-title>

  <v-form v-if="!!item" ref="itemForm">
    <v-row>
      <v-col class="form-col">
        
        <v-text-field label="Constrained text" v-model="item.c_constrained_text"
          :rules="[
            (v: string) => !!v || 'Constrained text is required',
            (v: string) => v.length == 6 || 'Constrained text must be 6 characters',
            (v: string) => v.toUpperCase() === v || 'Constrained text must be uppercase'
          ]"
        ></v-text-field>

        <v-text-field label="IP" v-model="item.c_ip"
          :rules="[
            (v: string) => !!v || 'IP is required',
            (v: string) => isValidIp(v) || 'IP must be a valid IPv4 or IPv6 address'
          ]"
        ></v-text-field>

        <v-textarea label="Long text" v-model="item.c_long_text" rows="6"
          :rules="[
            (v: string) => !!v || 'Long text is required',
            (v: string) => v.length <= 1000 || 'Long text must be 1000 characters or less'
          ]"
        ></v-textarea>

        <v-text-field label="Money amount" type="number" v-model.number="item.c_money_amount"
          :rules="[(v: number) => v != undefined || 'Money amount is required']"
        ></v-text-field>

        <v-text-field label="Percent" type="number" v-model.number="item.c_percent"
          :rules="[
            (v: number) => v != undefined || 'Percent is required',
            (v: number) => v >= 0 && v <= 10 || 'Percent must be between 0.0 and 10.0'
          ]"
        ></v-text-field>

      </v-col>
    </v-row>

    <l-cancel-and-save-actions :saving="saving" :showSaved="showSaved" :saveBtnLabel="saveBtnLabel" :saveDisabled="!auth.isWriter()"
      @cancel="emit('cancel')" @save="saveItem">
      <template #extra>
        <v-spacer />
        <v-btn v-if="props.id !== 0" :disabled="!auth.isWriter()" color="error" @click="deleteItem">Delete</v-btn>
      </template>
    </l-cancel-and-save-actions>

  </v-form>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useFormCrud } from 'lys-vue'
import ax from '@/api'
import auth from '@/auth'
import { type VariantType, NewVariantType, GetVariantTypeInputFromItem } from '@/types/core'

const props = defineProps<{
  id: number
}>()

const emit = defineEmits<{
  (e: 'cancel'): void
  (e: 'create', newID: number): void
  (e: 'delete'): void
  (e: 'load', id: number): void
  (e: 'update'): void
}>()

const { item, itemForm, saving, saveBtnLabel, showSaved, deleteItem, saveItem } =
  useFormCrud<VariantType>({
    ax,
    id: props.id,
    baseUrl: '/a/core/variant-types',
    newItem: NewVariantType,
    getInput: GetVariantTypeInputFromItem,

    onCreate: (id) => { emit('create', id) },
    onDelete: () => { emit('delete') },
    onLoad: (id) => emit('load', id),
    onUpdate: () => { emit('update') },
  })

const formTitle = computed(() => {
  return props.id !== 0 ? 'ID ' + props.id : 'New variant type'
})

function isValidIp(input: string | undefined | null): boolean {
  if (!input) return false
  const s = input.trim()

  // Reject CIDR/subnet suffix and zone IDs (e.g. /24, /64, %eth0)
  if (s.includes('/') || s.includes('%')) return false

  return isValidIPv4(s) || isValidIPv6(s)
}

function isValidIPv4(s: string): boolean {
  const parts = s.split('.')
  if (parts.length !== 4) return false

  for (const part of parts) {
    // Digits only
    if (!/^\d+$/.test(part)) return false
    // No leading zeros except "0"
    if (part.length > 1 && part.startsWith('0')) return false

    const n = Number(part)
    if (!Number.isInteger(n) || n < 0 || n > 255) return false
  }

  return true
}

function isValidIPv6(s: string): boolean {
  // At most one "::"
  if ((s.match(/::/g) || []).length > 1) return false

  let hasIpv4Tail = false
  let ipv4Tail = ''

  // Detect embedded IPv4 tail (e.g. ::ffff:192.168.0.1)
  const lastColon = s.lastIndexOf(':')
  if (s.includes('.') && lastColon !== -1) {
    ipv4Tail = s.slice(lastColon + 1)
    if (!isValidIPv4(ipv4Tail)) return false
    hasIpv4Tail = true
  }

  // Remove IPv4 tail from IPv6 hextet parsing
  const head = hasIpv4Tail ? s.slice(0, s.lastIndexOf(':')) : s
  const halves = head.split('::')
  if (halves.length > 2) return false

  const left = halves[0] ? halves[0].split(':') : []
  const right = halves.length === 2 && halves[1] ? halves[1].split(':') : []

  const isHextet = (h: string) => /^[0-9A-Fa-f]{1,4}$/.test(h)
  if (!left.every(isHextet) || !right.every(isHextet)) return false

  // Embedded IPv4 tail counts as 2 hextets
  const totalHextets = left.length + right.length + (hasIpv4Tail ? 2 : 0)

  if (halves.length === 1) {
    // No compression, must be exactly 8 hextets
    return totalHextets === 8
  }

  // With "::", compression must replace at least one hextet
  return totalHextets < 8
}

</script>
