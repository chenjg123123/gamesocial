const { request } = require('../../utils/request')
const {
  getToken,
  setToken,
  clearToken,
  getUserCache,
  setUserCache,
  clearUserCache,
} = require('../../utils/storage')

Page({
  data: {
    loading: false,
    saving: false,
    reloginLoading: false,
    loadingMore: false,
    userId: undefined,
    nickname: '',
    avatarUrl: '',
    ledger: [],
    nextCursor: '',
  },
  onShow() {
    this.refresh()
  },
  applyProfile(profile) {
    if (!profile) return
    this.setData({
      userId: profile.id,
      nickname: profile.nickname || '',
      avatarUrl: profile.avatarUrl || '',
    })
    setUserCache(profile)
  },
  refresh() {
    const that = this
    that.setData({ loading: true })
    Promise.all([
      request('/api/user/me').catch(() => null),
      request('/api/points/ledger?cursor=&limit=20').catch(() => ({ items: [], nextCursor: '' })),
    ])
      .then(resList => {
        const profile = resList[0]
        const ledgerRes = resList[1] || { items: [], nextCursor: '' }
        if (profile) that.applyProfile(profile)
        that.setData({
          ledger: ledgerRes.items || [],
          nextCursor: ledgerRes.nextCursor || '',
        })
      })
      .finally(() => {
        that.setData({ loading: false })
      })
  },
  onChooseAvatar(e) {
    const detail = e && e.detail ? e.detail : {}
    const avatarUrl = detail.avatarUrl || ''
    this.setData({ avatarUrl })
  },
  onNicknameInput(e) {
    const value = e && e.detail ? e.detail.value : ''
    this.setData({ nickname: value })
  },
  saveProfile() {
    const that = this
    const nickname = String(that.data.nickname || '').trim()
    const avatarUrl = String(that.data.avatarUrl || '').trim()
    that.setData({ saving: true })
    request('/api/user/me', { method: 'PUT', data: { nickname, avatarUrl } })
      .then(profile => {
        that.applyProfile(profile)
        wx.showToast({ title: '已保存', icon: 'success' })
      })
      .finally(() => {
        that.setData({ saving: false })
      })
  },
  relogin() {
    const that = this
    that.setData({ reloginLoading: true })
    wx.login({
      success: res => {
        if (!res.code) {
          that.setData({ reloginLoading: false })
          wx.showToast({ title: '登录失败', icon: 'none' })
          return
        }
        request('/api/auth/wechat/login', { method: 'POST', data: { code: res.code } })
          .then(result => {
            if (result && result.token) {
              setToken(result.token)
              const app = getApp()
              if (app && app.globalData) app.globalData.token = result.token
            }
            if (result && result.user) {
              setUserCache(result.user)
              const app = getApp()
              if (app && app.globalData) app.globalData.user = result.user
            }
            wx.showToast({ title: '登录成功', icon: 'success' })
            that.refresh()
          })
          .finally(() => {
            that.setData({ reloginLoading: false })
          })
      },
      fail: () => {
        that.setData({ reloginLoading: false })
        wx.showToast({ title: '登录失败', icon: 'none' })
      },
    })
  },
  doLogout() {
    clearToken()
    clearUserCache()
    const app = getApp()
    if (app && app.globalData) {
      app.globalData.token = getToken()
      app.globalData.user = undefined
    }
    this.setData({
      userId: undefined,
      nickname: '',
      avatarUrl: '',
      ledger: [],
      nextCursor: '',
    })
    wx.showToast({ title: '已退出', icon: 'success' })
  },
  loadMore() {
    const that = this
    if (!that.data.nextCursor) return
    that.setData({ loadingMore: true })
    request(
      `/api/points/ledger?cursor=${encodeURIComponent(that.data.nextCursor)}&limit=20`
    )
      .then(res => {
        const items = (res && res.items) || []
        const nextCursor = (res && res.nextCursor) || ''
        that.setData({
          ledger: that.data.ledger.concat(items),
          nextCursor: nextCursor,
        })
      })
      .finally(() => {
        that.setData({ loadingMore: false })
      })
  },
})

