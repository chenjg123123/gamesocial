const { request } = require('../../utils/request')
const { getUserCache, setUserCache } = require('../../utils/storage')

Page({
  data: {
    loading: false,
    user: getUserCache() || {},
    vipLabel: '普通用户',
    pointsBalance: 0,
    drinkQuantity: 0,
  },
  onShow() {
    this.refresh()
  },
  refresh() {
    const that = this
    that.setData({ loading: true })

    Promise.all([
      request('/api/user/me').catch(() => null),
      request('/api/vip/status').catch(() => null),
      request('/api/points/balance').catch(() => null),
      request('/api/redeem/drinks/balance').catch(() => null),
    ])
      .then(resList => {
        const profile = resList[0]
        const vip = resList[1]
        const points = resList[2]
        const drinks = resList[3]

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
          drinkQuantity:
            drinks && typeof drinks.quantity === 'number'
              ? drinks.quantity
              : that.data.drinkQuantity,
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
})

