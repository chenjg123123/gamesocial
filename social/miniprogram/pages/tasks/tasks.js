const { request } = require('../../utils/request')

Page({
  data: {
    loading: false,
    checkinLoading: false,
    items: [],
  },
  onShow() {
    this.refresh()
  },
  refresh() {
    const that = this
    that.setData({ loading: true })
    request('/api/tasks/overview')
      .then(res => {
        that.setData({ items: (res && res.items) || [] })
      })
      .catch(() => {
        that.setData({ items: [] })
      })
      .finally(() => {
        that.setData({ loading: false })
      })
  },
  doCheckin() {
    const that = this
    that.setData({ checkinLoading: true })
    request('/api/tasks/checkin', { method: 'POST' })
      .then(() => {
        wx.showToast({ title: '打卡成功', icon: 'success' })
        return that.refresh()
      })
      .finally(() => {
        that.setData({ checkinLoading: false })
      })
  },
  reward(e) {
    const taskId = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!taskId) return
    request(`/api/tasks/${taskId}/reward`, { method: 'POST' }).then(() => {
      wx.showToast({ title: '领取成功', icon: 'success' })
      this.refresh()
    })
  },
})

