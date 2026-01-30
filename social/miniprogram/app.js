const { request } = require('./utils/request')
const { getToken, getUserCache, setToken, setUserCache } = require('./utils/storage')

App({
  globalData: {},
  onLaunch() {
    const token = getToken()
    if (token) {
      this.globalData.token = token
      this.globalData.user = getUserCache()
      return
    }

    wx.login({
      success: res => {
        if (!res.code) {
          wx.showToast({ title: '登录失败', icon: 'none' })
          return
        }
        request('/api/auth/wechat/login', { method: 'POST', data: { code: res.code } })
          .then(result => {
            if (!result || !result.token) return
            setToken(result.token)
            this.globalData.token = result.token
            if (result.user) {
              setUserCache(result.user)
              this.globalData.user = result.user
            }
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
})

