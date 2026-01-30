const { request } = require('../../utils/request')

const asNumber = v => {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

const safeStr = v => (v === undefined || v === null ? '' : String(v))

Page({
  data: {
    tab: 'goods',
    loading: false,

    goods: [],
    tournaments: [],
    taskDefs: [],
    users: [],
    orders: [],

    goodsForm: { id: '', name: '', coverUrl: '', pointsPrice: '', stock: '', status: '1' },
    tournamentForm: {
      id: '',
      title: '',
      content: '',
      coverUrl: '',
      startAt: '',
      endAt: '',
      status: 'DRAFT',
      createdByAdminId: '1',
    },
    taskDefForm: {
      id: '',
      taskCode: '',
      name: '',
      periodType: 'DAILY',
      targetCount: '1',
      rewardPoints: '1',
      status: '1',
    },
    userForm: {
      id: '',
      nickname: '',
      avatarUrl: '',
      status: '1',
    },
    orderForm: {
      userId: '',
      goodsId: '',
      quantity: '1',
      pointsPrice: '',
    },
  },

  onShow() {
    this.refreshCurrent()
  },

  switchTab(e) {
    const tab = e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.tab
    if (!tab) return
    this.setData({ tab })
    this.refreshCurrent()
  },

  refreshCurrent() {
    const tab = this.data.tab
    if (tab === 'goods') return this.loadGoods()
    if (tab === 'tournaments') return this.loadTournaments()
    if (tab === 'taskDefs') return this.loadTaskDefs()
    if (tab === 'users') return this.loadUsers()
    if (tab === 'orders') return this.loadOrders()
  },

  withLoading(promise) {
    this.setData({ loading: true })
    return promise.finally(() => this.setData({ loading: false }))
  },

  loadGoods() {
    return this.withLoading(
      request('/admin/goods?offset=0&limit=50&status=0')
        .then(res => {
          const items = (res && res.items) || res || []
          this.setData({ goods: Array.isArray(items) ? items : [] })
        })
        .catch(err => {
          wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' })
          this.setData({ goods: [] })
        })
    )
  },

  editGoods(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    this.withLoading(
      request(`/admin/goods/${id}`)
        .then(item => {
          this.setData({
            goodsForm: {
              id: safeStr(item && item.id),
              name: safeStr(item && item.name),
              coverUrl: safeStr(item && item.coverUrl),
              pointsPrice: safeStr(item && item.pointsPrice),
              stock: safeStr(item && item.stock),
              status: safeStr(item && item.status) || '1',
            },
          })
        })
        .catch(err => {
          wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' })
        })
    )
  },

  deleteGoods(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    wx.showModal({
      title: '确认删除',
      content: `删除商品 #${id}？`,
      success: res => {
        if (!res.confirm) return
        this.withLoading(
          request(`/admin/goods/${id}`, { method: 'DELETE' })
            .then(() => {
              wx.showToast({ title: '已删除', icon: 'success' })
              this.resetGoodsForm()
              this.loadGoods()
            })
            .catch(err => wx.showToast({ title: safeStr((err && err.message) || '删除失败'), icon: 'none' }))
        )
      },
    })
  },

  onGoodsName(e) {
    this.setData({ 'goodsForm.name': safeStr(e && e.detail && e.detail.value) })
  },
  onGoodsCoverUrl(e) {
    this.setData({ 'goodsForm.coverUrl': safeStr(e && e.detail && e.detail.value) })
  },
  onGoodsPoints(e) {
    this.setData({ 'goodsForm.pointsPrice': safeStr(e && e.detail && e.detail.value) })
  },
  onGoodsStock(e) {
    this.setData({ 'goodsForm.stock': safeStr(e && e.detail && e.detail.value) })
  },
  onGoodsStatus(e) {
    this.setData({ 'goodsForm.status': safeStr(e && e.detail && e.detail.value) })
  },
  resetGoodsForm() {
    this.setData({ goodsForm: { id: '', name: '', coverUrl: '', pointsPrice: '', stock: '', status: '1' } })
  },
  submitGoods() {
    const form = this.data.goodsForm || {}
    const statusStr = safeStr(form.status).trim()
    const payload = {
      name: String(form.name || '').trim(),
      coverUrl: String(form.coverUrl || '').trim(),
      pointsPrice: asNumber(form.pointsPrice),
      stock: asNumber(form.stock),
      status: statusStr === '' ? 1 : asNumber(statusStr),
    }
    if (!payload.name) {
      wx.showToast({ title: '请输入商品名', icon: 'none' })
      return
    }

    const id = asNumber(form.id)
    const req = id
      ? request(`/admin/goods/${id}`, { method: 'PUT', data: payload })
      : request('/admin/goods', { method: 'POST', data: payload })

    this.withLoading(
      req
        .then(() => {
          wx.showToast({ title: id ? '已更新' : '已创建', icon: 'success' })
          this.resetGoodsForm()
          this.loadGoods()
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '提交失败'), icon: 'none' }))
    )
  },

  loadTournaments() {
    return this.withLoading(
      request('/admin/tournaments?offset=0&limit=50')
        .then(res => {
          const items = (res && res.items) || res || []
          this.setData({ tournaments: Array.isArray(items) ? items : [] })
        })
        .catch(err => {
          wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' })
          this.setData({ tournaments: [] })
        })
    )
  },
  editTournament(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    this.withLoading(
      request(`/admin/tournaments/${id}`)
        .then(item => {
          this.setData({
            tournamentForm: {
              id: safeStr(item && item.id),
              title: safeStr(item && item.title),
              content: safeStr(item && item.content),
              coverUrl: safeStr(item && item.coverUrl),
              startAt: safeStr(item && item.startAt),
              endAt: safeStr(item && item.endAt),
              status: safeStr(item && item.status) || 'DRAFT',
              createdByAdminId: safeStr(item && item.createdByAdminId) || '1',
            },
          })
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' }))
    )
  },
  deleteTournament(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    wx.showModal({
      title: '确认删除',
      content: `删除赛事 #${id}？`,
      success: res => {
        if (!res.confirm) return
        this.withLoading(
          request(`/admin/tournaments/${id}`, { method: 'DELETE' })
            .then(() => {
              wx.showToast({ title: '已删除', icon: 'success' })
              this.resetTournamentForm()
              this.loadTournaments()
            })
            .catch(err => wx.showToast({ title: safeStr((err && err.message) || '删除失败'), icon: 'none' }))
        )
      },
    })
  },
  onTournamentTitle(e) {
    this.setData({ 'tournamentForm.title': safeStr(e && e.detail && e.detail.value) })
  },
  onTournamentStartAt(e) {
    this.setData({ 'tournamentForm.startAt': safeStr(e && e.detail && e.detail.value) })
  },
  onTournamentEndAt(e) {
    this.setData({ 'tournamentForm.endAt': safeStr(e && e.detail && e.detail.value) })
  },
  onTournamentContent(e) {
    this.setData({ 'tournamentForm.content': safeStr(e && e.detail && e.detail.value) })
  },
  onTournamentCoverUrl(e) {
    this.setData({ 'tournamentForm.coverUrl': safeStr(e && e.detail && e.detail.value) })
  },
  onTournamentStatus(e) {
    this.setData({ 'tournamentForm.status': safeStr(e && e.detail && e.detail.value) })
  },
  onTournamentCreatedBy(e) {
    this.setData({ 'tournamentForm.createdByAdminId': safeStr(e && e.detail && e.detail.value) })
  },
  resetTournamentForm() {
    this.setData({
      tournamentForm: {
        id: '',
        title: '',
        content: '',
        coverUrl: '',
        startAt: '',
        endAt: '',
        status: 'DRAFT',
        createdByAdminId: '1',
      },
    })
  },
  submitTournament() {
    const form = this.data.tournamentForm || {}
    const payload = {
      title: String(form.title || '').trim(),
      content: String(form.content || '').trim(),
      coverUrl: String(form.coverUrl || '').trim(),
      startAt: String(form.startAt || '').trim(),
      endAt: String(form.endAt || '').trim(),
      status: String(form.status || '').trim() || 'DRAFT',
      createdByAdminId: asNumber(form.createdByAdminId) || 1,
    }
    if (!payload.title) {
      wx.showToast({ title: '请输入标题', icon: 'none' })
      return
    }
    const id = asNumber(form.id)
    const req = id
      ? request(`/admin/tournaments/${id}`, { method: 'PUT', data: payload })
      : request('/admin/tournaments', { method: 'POST', data: payload })
    this.withLoading(
      req
        .then(() => {
          wx.showToast({ title: id ? '已更新' : '已创建', icon: 'success' })
          this.resetTournamentForm()
          this.loadTournaments()
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '提交失败'), icon: 'none' }))
    )
  },

  loadTaskDefs() {
    return this.withLoading(
      request('/admin/task-defs?offset=0&limit=50&status=0')
        .then(res => {
          const items = (res && res.items) || res || []
          this.setData({ taskDefs: Array.isArray(items) ? items : [] })
        })
        .catch(err => {
          wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' })
          this.setData({ taskDefs: [] })
        })
    )
  },
  editTaskDef(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    this.withLoading(
      request(`/admin/task-defs/${id}`)
        .then(item => {
          this.setData({
            taskDefForm: {
              id: safeStr(item && item.id),
              taskCode: safeStr(item && item.taskCode),
              name: safeStr(item && item.name),
              periodType: safeStr(item && item.periodType) || 'DAILY',
              targetCount: safeStr(item && item.targetCount),
              rewardPoints: safeStr(item && item.rewardPoints),
              status: safeStr(item && item.status) || '1',
            },
          })
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' }))
    )
  },
  deleteTaskDef(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    wx.showModal({
      title: '确认删除',
      content: `删除任务定义 #${id}？`,
      success: res => {
        if (!res.confirm) return
        this.withLoading(
          request(`/admin/task-defs/${id}`, { method: 'DELETE' })
            .then(() => {
              wx.showToast({ title: '已删除', icon: 'success' })
              this.resetTaskDefForm()
              this.loadTaskDefs()
            })
            .catch(err => wx.showToast({ title: safeStr((err && err.message) || '删除失败'), icon: 'none' }))
        )
      },
    })
  },
  onTaskDefCode(e) {
    this.setData({ 'taskDefForm.taskCode': safeStr(e && e.detail && e.detail.value) })
  },
  onTaskDefName(e) {
    this.setData({ 'taskDefForm.name': safeStr(e && e.detail && e.detail.value) })
  },
  onTaskDefPeriod(e) {
    this.setData({ 'taskDefForm.periodType': safeStr(e && e.detail && e.detail.value) })
  },
  onTaskDefTarget(e) {
    this.setData({ 'taskDefForm.targetCount': safeStr(e && e.detail && e.detail.value) })
  },
  onTaskDefReward(e) {
    this.setData({ 'taskDefForm.rewardPoints': safeStr(e && e.detail && e.detail.value) })
  },
  onTaskDefStatus(e) {
    this.setData({ 'taskDefForm.status': safeStr(e && e.detail && e.detail.value) })
  },
  resetTaskDefForm() {
    this.setData({
      taskDefForm: {
        id: '',
        taskCode: '',
        name: '',
        periodType: 'DAILY',
        targetCount: '1',
        rewardPoints: '1',
        status: '1',
      },
    })
  },
  submitTaskDef() {
    const form = this.data.taskDefForm || {}
    const taskCode = String(form.taskCode || '').trim()
    const name = String(form.name || '').trim()
    const periodType = String(form.periodType || '').trim() || 'DAILY'
    const targetCount = asNumber(form.targetCount)
    const rewardPoints = asNumber(form.rewardPoints)
    const statusStr = safeStr(form.status).trim()
    const status = statusStr === '' ? 1 : asNumber(statusStr)

    if (!name) {
      wx.showToast({ title: '请输入任务名', icon: 'none' })
      return
    }
    const id = asNumber(form.id)
    if (!id && !taskCode) {
      wx.showToast({ title: '请输入 taskCode', icon: 'none' })
      return
    }

    const payload = id
      ? { name, periodType, targetCount, rewardPoints, status, ruleJson: {} }
      : { taskCode, name, periodType, targetCount, rewardPoints, status, ruleJson: {} }
    const req = id
      ? request(`/admin/task-defs/${id}`, { method: 'PUT', data: payload })
      : request('/admin/task-defs', { method: 'POST', data: payload })
    this.withLoading(
      req
        .then(() => {
          wx.showToast({ title: id ? '已更新' : '已创建', icon: 'success' })
          this.resetTaskDefForm()
          this.loadTaskDefs()
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '提交失败'), icon: 'none' }))
    )
  },

  loadUsers() {
    return this.withLoading(
      request('/admin/users?offset=0&limit=50')
        .then(res => {
          const items = (res && res.items) || res || []
          this.setData({ users: Array.isArray(items) ? items : [] })
        })
        .catch(err => {
          wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' })
          this.setData({ users: [] })
        })
    )
  },
  editUser(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    this.withLoading(
      request(`/admin/users/${id}`)
        .then(item => {
          this.setData({
            userForm: {
              id: safeStr(item && item.id),
              nickname: safeStr(item && item.nickname),
              avatarUrl: safeStr(item && item.avatarUrl),
              status: safeStr(item && item.status),
            },
          })
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' }))
    )
  },
  onUserNickname(e) {
    this.setData({ 'userForm.nickname': safeStr(e && e.detail && e.detail.value) })
  },
  onUserAvatarUrl(e) {
    this.setData({ 'userForm.avatarUrl': safeStr(e && e.detail && e.detail.value) })
  },
  onUserStatus(e) {
    this.setData({ 'userForm.status': safeStr(e && e.detail && e.detail.value) })
  },
  resetUserForm() {
    this.setData({ userForm: { id: '', nickname: '', avatarUrl: '', status: '1' } })
  },
  submitUser() {
    const form = this.data.userForm || {}
    const id = asNumber(form.id)
    if (!id) {
      wx.showToast({ title: '请选择用户', icon: 'none' })
      return
    }
    const statusStr = safeStr(form.status).trim()
    const payload = {
      nickname: String(form.nickname || '').trim(),
      avatarUrl: String(form.avatarUrl || '').trim(),
      status: statusStr === '' ? 1 : asNumber(statusStr),
    }
    this.withLoading(
      request(`/admin/users/${id}`, { method: 'PUT', data: payload })
        .then(() => {
          wx.showToast({ title: '已更新', icon: 'success' })
          this.resetUserForm()
          this.loadUsers()
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '提交失败'), icon: 'none' }))
    )
  },

  loadOrders() {
    return this.withLoading(
      request('/admin/redeem/orders?offset=0&limit=50')
        .then(res => {
          const items = (res && res.items) || res || []
          this.setData({ orders: Array.isArray(items) ? items : [] })
        })
        .catch(err => {
          wx.showToast({ title: safeStr((err && err.message) || '加载失败'), icon: 'none' })
          this.setData({ orders: [] })
        })
    )
  },
  useOrder(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    this.withLoading(
      request(`/admin/redeem/orders/${id}/use`, { method: 'PUT' })
        .then(() => {
          wx.showToast({ title: '已核销', icon: 'success' })
          this.loadOrders()
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '操作失败'), icon: 'none' }))
    )
  },
  cancelOrder(e) {
    const id = Number(e && e.currentTarget && e.currentTarget.dataset && e.currentTarget.dataset.id)
    if (!id) return
    this.withLoading(
      request(`/admin/redeem/orders/${id}/cancel`, { method: 'PUT' })
        .then(() => {
          wx.showToast({ title: '已取消', icon: 'success' })
          this.loadOrders()
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '操作失败'), icon: 'none' }))
    )
  },
  onOrderUserId(e) {
    this.setData({ 'orderForm.userId': safeStr(e && e.detail && e.detail.value) })
  },
  onOrderGoodsId(e) {
    this.setData({ 'orderForm.goodsId': safeStr(e && e.detail && e.detail.value) })
  },
  onOrderQuantity(e) {
    this.setData({ 'orderForm.quantity': safeStr(e && e.detail && e.detail.value) })
  },
  onOrderPointsPrice(e) {
    this.setData({ 'orderForm.pointsPrice': safeStr(e && e.detail && e.detail.value) })
  },
  resetOrderForm() {
    this.setData({ orderForm: { userId: '', goodsId: '', quantity: '1', pointsPrice: '' } })
  },
  submitOrder() {
    const form = this.data.orderForm || {}
    const userId = asNumber(form.userId)
    const goodsId = asNumber(form.goodsId)
    const quantity = asNumber(form.quantity)
    const pointsPriceStr = safeStr(form.pointsPrice).trim()
    const pointsPrice = asNumber(pointsPriceStr)
    if (!userId || !goodsId || !quantity || pointsPriceStr === '' || pointsPrice < 0) {
      wx.showToast({ title: '请填写 userId/goodsId/数量/pointsPrice', icon: 'none' })
      return
    }
    const payload = {
      userId,
      items: [{ goodsId, quantity, pointsPrice }],
    }
    this.withLoading(
      request('/admin/redeem/orders', { method: 'POST', data: payload })
        .then(res => {
          const orderNo = res && (res.orderNo || res.id)
          wx.showModal({
            title: '创建成功',
            content: orderNo ? `订单：${orderNo}` : '已创建',
            showCancel: false,
          })
          this.resetOrderForm()
          this.loadOrders()
        })
        .catch(err => wx.showToast({ title: safeStr((err && err.message) || '提交失败'), icon: 'none' }))
    )
  },
})
