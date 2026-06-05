<template>
  <v-app class="rounded rounded-md">
    <v-main class="d-flex" style="min-height: 300px;">
      <v-container class="fill-height">
        <v-responsive class="align-center text-center fill-height">
          <v-row>
            <v-col>

              <ApiError></ApiError>

              <v-card max-width="500" class="mx-auto">
                <v-card-title class="d-flex flex-wrap justify-center mt-4">
                  <v-img max-height="35px" max-width="35px" src="./assets/logo.png"></v-img>
                  <span class="ml-4">{{ appStore.company }} - {{ appStore.projectTitle }}</span>
                </v-card-title>
                <v-card-text class="pa-5">
                  <v-form ref="loginForm" @submit.prevent="login">

                    <v-text-field label="User name" v-model="userName"
                      :rules="[(v: string) => !!v || 'User name is required']"
                    ></v-text-field>

                    <v-text-field label="Password" v-model="password"
                      :rules="[(v: string) => !!v || 'Password is required']"
                      :type="showPw ? 'text' : 'password'"
                      :append-inner-icon="showPw ? 'mdi-eye' : 'mdi-eye-off'"
                      @click:append-inner="showPw = !showPw"
                    ></v-text-field>

                    <v-btn color="secondary" block class="mt-2" type="submit" :loading="loading">Login</v-btn>
                  </v-form>

                </v-card-text>

              </v-card>

            </v-col>
          </v-row>
        </v-responsive>
      </v-container>
    </v-main>
  </v-app>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { type VForm } from 'vuetify/components'
import ax, { setAuthToken } from '@/api'
import auth from '@/auth'
import { useAppStore } from '@/stores/app'
import { type LoginResponse } from '@/types/system'

const route = useRoute()
const router = useRouter()

const appStore = useAppStore()

const loginForm = ref<InstanceType<typeof VForm>>()
const userName = ref('')
const password = ref('')
const showPw = ref(false)

const loading = ref(false)

async function login() {
  const result = await loginForm.value?.validate()
  if (!result?.valid) { return }

  loading.value = true

  await ax.post('/login', { user_name: userName.value, password: password.value })
    .then(response => { 
      //console.log(response)

      // replace session token
      sessionStorage.removeItem('token')
      sessionStorage.setItem('token', response.data.data.token)

      // set token as default auth header in axios
      setAuthToken(ax, response.data.data.token)

      loginSuccess(response.data.data)
    })
    .catch() // handled by interceptor
    .finally(() => loading.value = false )
}

function loginSuccess(respData: LoginResponse) {

  // flag user as authenticated, write user props
  auth.assignUser(respData)

  // redirect to intended path, or home if no intended path or invalid path
  const rawTo = route.query.to?.toString()
  let dest = { name: 'Home', params: {} }

  if (rawTo) {
    try {
      const parsed = JSON.parse(decodeURIComponent(rawTo)) as {
        name?: string
        params?: Record<string, string | number>
      }

      if (parsed.name && router.hasRoute(parsed.name.toString())) {
        dest = {
          name: parsed.name.toString(),
          params: parsed.params ?? {},
        }
      }
    } catch {
      // fall back to Home
    }
  }

  router.push(dest)
}

</script>

<style>
</style>