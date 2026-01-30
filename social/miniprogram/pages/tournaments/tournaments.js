const { request } = require('../../utils/request')

Page({
  data: {
    loading: false,
    items: [],
    results: [],
  },
  onShow() {
    this.refresh()
  },
  refresh() {
    const that = this
    that.setData({ loading: true })
    Promise.all([
      request('/api/tournaments').catch(() => ({ items: [] })),
      request('/api/tournaments/my/results').catch(() => ({ items: [] })),
    ])
      .then(resList => {
        const tournaments = resList[0] || { items: [] }
        const results = resList[1] || { items: [] }
        that.setData({
          items: tournaments.items || [],
          results: results.items || [],
        })
      })
      .finally(() => {
        that.setData({ loading: false })
      })
  },
  join(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    request(`/api/tournaments/${id}/join`, { method: 'POST' }).then(() => this.refresh())
  },
  cancelJoin(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    request(`/api/tournaments/${id}/join`, { method: 'DELETE' }).then(() => this.refresh())
  },
})

