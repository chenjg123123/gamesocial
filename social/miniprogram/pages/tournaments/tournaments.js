const { request } = require('../../utils/request')

Page({
  data: {
    loading: false,
    items: [],
  },
  onShow() {
    this.refresh()
  },
  refresh() {
    const that = this
    that.setData({ loading: true })
    request('/api/tournaments?offset=0&limit=50')
      .then(res => {
        const items = Array.isArray(res) ? res : (res && res.items) || []
        that.setData({ items })
      })
      .catch(() => {
        that.setData({ items: [] })
      })
      .finally(() => {
        that.setData({ loading: false })
      })
  },
  join(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    request(`/api/tournaments/${id}/join`, { method: 'POST' })
      .then(() => this.refresh())
      .catch(err => wx.showToast({ title: (err && err.message) || '报名失败', icon: 'none' }))
  },
  cancelJoin(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    request(`/api/tournaments/${id}/cancel`, { method: 'PUT' })
      .then(() => this.refresh())
      .catch(err => wx.showToast({ title: (err && err.message) || '取消失败', icon: 'none' }))
  },
})

