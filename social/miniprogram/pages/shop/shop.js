const { request } = require('../../utils/request')
const { getUserCache } = require('../../utils/storage')

Page({
  data: {
    loading: false,
    goods: [],
    orders: [],
  },
  onShow() {
    this.refresh()
  },
  refresh() {
    const that = this
    that.setData({ loading: true })
    Promise.all([
      request('/api/goods?offset=0&limit=50').catch(() => []),
      request('/api/redeem/orders?offset=0&limit=20').catch(() => []),
    ])
      .then(resList => {
        const goods = resList[0]
        const orders = resList[1]
        that.setData({
          goods: Array.isArray(goods) ? goods : (goods && goods.items) || [],
          orders: Array.isArray(orders) ? orders : (orders && orders.items) || [],
        })
      })
      .finally(() => {
        that.setData({ loading: false })
      })
  },
  redeemGoods(e) {
    const goodsId = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!goodsId) return
    const user = getUserCache && getUserCache()
    const userId = user && typeof user.id === 'number' ? user.id : Number(user && user.id)
    if (!userId) {
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }
    const goods = (this.data.goods || []).find(g => Number(g && g.id) === goodsId)
    const pointsPrice = goods && typeof goods.pointsPrice === 'number' ? goods.pointsPrice : NaN
    if (!Number.isFinite(pointsPrice)) {
      wx.showToast({ title: '商品价格异常', icon: 'none' })
      return
    }
    const item = { goodsId: goodsId, quantity: 1, pointsPrice }
    request('/api/redeem/orders', { method: 'POST', data: { userId, items: [item] } })
      .then(order => {
        wx.showModal({
          title: '下单成功',
          content: `订单号：${order && order.orderNo ? order.orderNo : ''}`,
          showCancel: false,
        })
        this.refresh()
      })
      .catch(err => wx.showToast({ title: (err && err.message) || '下单失败', icon: 'none' }))
  },
  cancelOrder(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    wx.showModal({
      title: '确认取消',
      content: '取消该订单？',
      success: res => {
        if (!res.confirm) return
        request(`/api/redeem/orders/${id}/cancel`, { method: 'PUT' }).then(() => {
          wx.showToast({ title: '已取消', icon: 'success' })
          this.refresh()
        })
      },
    })
  },
})

