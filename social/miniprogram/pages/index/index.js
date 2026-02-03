const { request } = require('../../utils/request')
const { getUserCache, setUserCache, setToken } = require('../../utils/storage')

Page({
  data: {
    loading: false,
    user: getUserCache() || {},
    vipLabel: '普通用户',
    pointsBalance: 0,
  },
  onShow() {
    this.refresh()
  },
  refresh() {
    const that = this
    that.setData({ loading: true })

    Promise.all([
      request('/api/users/me').catch(() => null),
      request('/api/vip/status').catch(() => null),
      request('/api/points/balance').catch(() => null),
    ])
      .then(resList => {
        const profile = resList[0]
        const vip = resList[1]
        const points = resList[2]

        if (profile) setUserCache(profile)

        let vipLabel = '普通用户'
        if (vip && vip.active) {
          const plan = vip.plan ? vip.plan : ''
          vipLabel = (`VIP ${plan}`).trim()
        }

        that.setData({
          user: profile || that.data.user,
          vipLabel,
          pointsBalance:
            points && typeof points.balance === 'number' ? points.balance : that.data.pointsBalance,
        })
      })
      .finally(() => {
        that.setData({ loading: false })
      })
  },
  switchTab(e) {
    const url = e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.url
    if (!url) return
    wx.switchTab({ url })
  },
  authorizeLogin() {
    const that = this
    if (!wx.getUserProfile) {
      wx.showToast({ title: '当前微信版本不支持授权', icon: 'none' })
      return
    }

    wx.getUserProfile({
      desc: '用于完善用户资料（昵称、头像）',
      success: res => {
        const userInfo = (res && res.userInfo) || {}
        const nickname = String(userInfo.nickName || '').trim()
        const avatarUrl = String(userInfo.avatarUrl || '').trim()

        that.setData({
          user: Object.assign({}, that.data.user || {}, { nickname, avatarUrl }),
        })

        wx.login({
          success: loginRes => {
            if (!loginRes || !loginRes.code) {
              wx.showToast({ title: '登录失败', icon: 'none' })
              return
            }

            request('/api/auth/wechat/login', { method: 'POST', data: { code: loginRes.code } })
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

                if (nickname || avatarUrl) {
                  return request('/api/users/me', { method: 'PUT', data: { nickname, avatarUrl } })
                    .then(profile => {
                      if (profile) setUserCache(profile)
                      that.setData({ user: profile || that.data.user })
                    })
                    .catch(() => null)
                }
                return null
              })
              .then(() => {
                that.refresh()
                wx.showToast({ title: '已授权登录', icon: 'success' })
              })
              .catch(() => {
                wx.showToast({ title: '登录失败', icon: 'none' })
              })
          },
          fail: () => {
            wx.showToast({ title: '登录失败', icon: 'none' })
          },
        })
      },
      fail: () => {
        wx.showToast({ title: '已取消授权', icon: 'none' })
      },
    })
  },
})

