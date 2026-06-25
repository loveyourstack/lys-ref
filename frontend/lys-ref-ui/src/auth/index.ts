import ax, { setAuthToken, deleteAuthToken } from '@/api'
import { i18n } from '@/plugins'
import { useAppStore } from '@/stores/app'
import { useNotsStore } from '@/stores/notifications'
import { Role, WriterRoles, type LoginResponse } from '@/types/system'

type AuthUserData = Omit<LoginResponse, 'token'>

export default {
  ready: false,

  // move to pinia auth store? but need to ensure that the pinia store is ready before router tries to use it in nav guard
  user: {
    authenticated: false,
    force_password_change: false,
    has_aws_sg_rules: false,
    geo_ip_country_iso_code: '',
    geo_ip_location: '',
    id: 0,
    ip: '',
    name: '',
    roles: [] as Role[],
  },

  assignUser (data: AuthUserData) {
    this.user.authenticated = true
    this.user.force_password_change = data.force_password_change
    this.user.has_aws_sg_rules = data.has_aws_sg_rules
    this.user.geo_ip_country_iso_code = data.geo_ip_country_iso_code
    this.user.geo_ip_location = data.geo_ip_location
    this.user.id = data.user_id
    this.user.ip = data.ip
    this.user.name = data.user_name
    this.user.roles = data.roles

    this.setLocale(data.geo_ip_country_iso_code)

    // note: lazy-load store instances with (), or else they won't be initialized when this runs during auth bootstrap before app init

    // start websocket connection for notifications
    useNotsStore().wsStart()
  },
  revokeUser () {
    this.user.authenticated = false
    this.user.force_password_change = false
    this.user.has_aws_sg_rules = false
    this.user.geo_ip_country_iso_code = ''
    this.user.geo_ip_location = ''
    this.user.id = 0
    this.user.ip = ''
    this.user.name = ''
    this.user.roles = []

    useNotsStore().wsStop()
  },

  setLocale (geoIpCountryIsoCode: string) {
    if (!geoIpCountryIsoCode) {
      //console.log('geoIpCountryIsoCode is empty, skipping locale set')
      return
    }
    const geoCountryIso = geoIpCountryIsoCode.toLowerCase()

    // exit if a locale is already set in LS
    const lsJSON = localStorage.getItem('main')
    if (lsJSON) {
      const lsObj = JSON.parse(lsJSON)
      if (lsObj['locale']) {
        //console.log('locale already set in localStorage, skipping geoIP locale set')
        return
      }
    }

    // exit if the geoCountryIso is not one of the available locales
    if (!useAppStore().locales.some(l => l.code === geoCountryIso)) {
      //console.log(`geoCountryIso ${geoCountryIso} is not an available locale, skipping locale set`)
      return
    }

    // set default locale
    i18n.global.locale.value = geoCountryIso
  },

  // bootstrap auth state from session token, if it exists. Should be called once on app startup before router and app init
  async bootstrap() {
    this.ready = false

    const token = sessionStorage.getItem('token')
    if (!token) {
      this.user.authenticated = false
      this.ready = true
      return
    }

    setAuthToken(ax, token)

    try {
      const response = await ax.post('/session-token-login')
      this.assignUser(response.data.data)
    } catch {
      sessionStorage.removeItem('token')
      deleteAuthToken(ax)
      this.user.authenticated = false
    } finally {
      this.ready = true
    }
  },

  hasRole (role: Role) {
    return this.user.roles.includes(role)
  },
  isWriter () {
    return this.user.roles.some(role => WriterRoles.includes(role))
  },

  logout () {
    this.revokeUser()
    sessionStorage.removeItem('token')
    deleteAuthToken(ax)
  },
}