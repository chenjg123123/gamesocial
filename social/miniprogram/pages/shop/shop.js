const { request } = require('../../utils/request')

Page({
  data: {
    loading: false,
    exchangeLoading: false,
    drinkQuantity: 0,
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
      request('/api/redeem/drinks/balance').catch(() => null),
      request('/api/goods').catch(() => ({ items: [] })),
      request('/api/redeem/orders?cursor=&limit=20').catch(() => ({ items: [] })),
    ])
      .then(resList => {
        const drinks = resList[0]
        const goods = resList[1] || { items: [] }
        const orders = resList[2] || { items: [] }
        that.setData({
          drinkQuantity: drinks && typeof drinks.quantity === 'number' ? drinks.quantity : 0,
          goods: goods.items || [],
          orders: orders.items || [],
        })
      })
      .finally(() => {
        that.setData({ loading: false })
      })
  },
  exchangeDrink() {
    const that = this
    that.setData({ exchangeLoading: true })
    request('/api/redeem/drinks/exchange', { method: 'POST' })
      .then(res => {
        that.setData({ drinkQuantity: res && typeof res.quantity === 'number' ? res.quantity : that.data.drinkQuantity })
        wx.showToast({ title: '兑换成功', icon: 'success' })
      })
      .finally(() => {
        that.setData({ exchangeLoading: false })
      })
  },
  redeemGoods(e) {
    const goodsId = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!goodsId) return
    request('/api/redeem/orders', { method: 'POST', data: { items: [{ goodsId: goodsId, quantity: 1 }] } }).then(order => {
      wx.showModal({
        title: '下单成功',
        content: `订单号：${order && order.orderNo ? order.orderNo : ''}`,
        showCancel: false,
      })
      this.refresh()
    })
  },
})

