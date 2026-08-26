<template>
  <v-row class="fill-height">
    <v-col>
      <v-sheet height="64">
        <v-toolbar flat>
          <v-btn class="me-4" color="grey-darken-2" variant="outlined" @click="setToday">
            Today
          </v-btn>

          <v-btn color="grey-darken-2" size="small" variant="text" icon @click="prev">
            <v-icon size="small">
              mdi-chevron-left
            </v-icon>
          </v-btn>

          <v-btn color="grey-darken-2" size="small" variant="text" icon @click="next">
            <v-icon size="small">
              mdi-chevron-right
            </v-icon>
          </v-btn>

          <v-toolbar-title v-if="calendar">
            {{ calendar.title }}
          </v-toolbar-title>

          <v-spacer></v-spacer>

          <v-form class="d-flex align-center">
            <v-checkbox v-model="showAnniversaries" :color="anniversaryColor" class="me-4" label="Anniversaries" hide-details dense />
            <v-checkbox v-model="showBirthdays" :color="birthdayColor" class="me-4" label="Birthdays" hide-details dense />
          </v-form>

        </v-toolbar>
      </v-sheet>
      <v-sheet height="550">
        <v-calendar ref="calendar" v-model="focus" type="month"
          :events="events" @change="loadItems" :locale="calendarLocale"
        >
          <template #event="{ event }">
            <v-tooltip location="top">
              <template #activator="{ props }">
                <div v-bind="props" class="px-1 text-truncate cursor-pointer">
                  {{ event.name }}
                </div>
              </template>
              <span>{{ event.message }}</span>
            </v-tooltip>
          </template>

        </v-calendar>
      </v-sheet>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { i18n } from '@/plugins'
import ax from '@/api'
import { type CalendarEvent, type CalendarChangeEvent, type Event } from '@/types/lysinc'

const calendar = ref()
const focus = ref('')

const showAnniversaries = ref(true)
const showBirthdays = ref(true)

const anniversaryColor = 'green'
const birthdayColor = 'pink'

const calendarLocale = computed(() => i18n.global.locale.value)

const baseUrl = '/a/lysinc/events-in-month'

const items = ref(<Event[]>([]))

// computed: reacts to both db items and event type filter changes
const events = computed<CalendarEvent[]>(() =>
  items.value
    .filter((item) =>
      (item.event_type !== 'Anniversary' || showAnniversaries.value) &&
      (item.event_type !== 'Birthday' || showBirthdays.value)
    )
    .map((item) => ({
      name: item.full_name,
      category: item.event_type,
      start: new Date(item.event_date),
      end: new Date(item.event_date),
      color: getEventColor(item.event_type),
      timed: false,
      message: item.message
    }))
)

function getEventColor(eventType: string): string {
  switch (eventType) {
    case 'Anniversary':
      return anniversaryColor
    case 'Birthday':
      return birthdayColor
    default:
      return 'primary'
  }
}

function loadItems(range?: CalendarChangeEvent) {

  // define default fallback if range is not provided
  let baseDate = new Date().toISOString().split('T')[0] // YYYY-MM-DD format

  // define base date as 1st day of the month from the change event passed by v-calendar
  if (range?.start) {
    const year = range.start.year
    const month = String(range.start.month).padStart(2, '0')
    baseDate = `${year}-${month}-01`
  }

  ax.get(`${baseUrl}?base_date=${baseDate}`).then((res) => {
    items.value = res.data.data
  })
  .catch() // handled by interceptor
}

function next () {
  calendar.value.next()
}

function prev () {
  calendar.value.prev()
}

function setToday () {
  focus.value = ''
}

onMounted(() => {
  calendar.value.checkChange()
})

</script>

<style scoped>
/* weekday headers are by default squished against the top border: add some padding */
:deep(.v-calendar-weekly__head-weekday) {
  padding-top: 8px !important;
}
</style>