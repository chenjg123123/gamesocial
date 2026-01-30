const { clearToken, getToken } = require('./storage')

const BASE_URL = 'http://localhost:8080'

const normalizeErrorMessage = data => {
  if (!data || typeof data !== 'object') return '请求失败'
  return data.message || data.msg || '请求失败'
}

const extractData = data => {
  if (!data || typeof data !== 'object') return data
  if (Object.prototype.hasOwnProperty.call(data, 'data')) return data.data
  return data
}

const request = (path, options) => {
  const opt = options || {}
  return new Promise((resolve, reject) => {
    const token = getToken()
    const header = Object.assign({ 'Content-Type': 'application/json' }, opt.header || {})
    if (token) header.Authorization = `Bearer ${token}`

    wx.request({
      url: `${BASE_URL}${path}`,
      method: opt.method || 'GET',
      data: opt.data,
      header,
      timeout: typeof opt.timeout === 'number' ? opt.timeout : 15000,
      success: res => {
        const statusCode = res.statusCode
        if (statusCode === 401) {
          clearToken()
          wx.showToast({ title: '请重新登录', icon: 'none' })
          wx.reLaunch({ url: '/pages/me/me' })
          reject(new Error('unauthorized'))
          return
        }

        if (statusCode < 200 || statusCode >= 300) {
          const msg = normalizeErrorMessage(res.data)
          wx.showToast({ title: msg, icon: 'none' })
          reject(new Error(msg))
          return
        }

        const envelope = res.data || {}
        const code = envelope.code
        if (typeof code === 'number' && code !== 0) {
          const msg = envelope.message || envelope.msg || '请求失败'
          wx.showToast({ title: msg, icon: 'none' })
          reject(new Error(msg))
          return
        }

        resolve(extractData(res.data))
      },
      fail: err => {
        const msg = (err && err.errMsg) || '网络异常'
        wx.showToast({ title: msg, icon: 'none' })
        reject(err)
      },
    })
  })
}

module.exports = {
  BASE_URL,
  request,
}

