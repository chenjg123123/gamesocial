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
    request('/api/tasks')
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
  doCheckin() {
    const that = this
    that.setData({ checkinLoading: true })
    request('/api/tasks/checkin', { method: 'POST' })
      .then(() => {
        wx.showToast({ title: '打卡成功', icon: 'success' })
        return that.refresh()
      })
      .catch(err => wx.showToast({ title: (err && err.message) || '打卡失败', icon: 'none' }))
      .finally(() => {
        that.setData({ checkinLoading: false })
      })
  },
  claim(e) {
    const code = e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.code
    const taskCode = code ? String(code) : ''
    if (!taskCode) return
    request(`/api/tasks/${encodeURIComponent(taskCode)}/claim`, { method: 'POST' })
      .then(() => {
        wx.showToast({ title: '领取成功', icon: 'success' })
        this.refresh()
      })
      .catch(err => wx.showToast({ title: (err && err.message) || '领取失败', icon: 'none' }))
  },
})

