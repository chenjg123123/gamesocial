const { clearToken, getToken } = require('./storage')

const BASE_URL = 'https://gamesocial2-223439-8-1362326232.sh.run.tcloudbase.com'

const normalizeErrorMessage = data => {
  if (!data || typeof data !== 'object') return '请求失败'
  return data.message || data.msg || '请求失败'
}

const extractData = data => {
  if (!data || typeof data !== 'object') return data
  if (Object.prototype.hasOwnProperty.call(data, 'data')) return data.data
  return data
}

const normalizeNetworkErrorMessage = err => {
  const raw = (err && err.errMsg) || ''
  const lower = String(raw).toLowerCase()
  if (lower.includes('not in domain list')) {
    return '请求域名未配置：请配置小程序 request 合法域名或在开发工具关闭域名校验'
  }
  if (lower.includes('timeout')) return '请求超时'
  if (lower.includes('ssl')) return 'HTTPS/证书异常'
  return raw || '网络异常'
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
        const hasBizCode = envelope && typeof envelope === 'object' && Object.prototype.hasOwnProperty.call(envelope, 'code')
        if (hasBizCode) {
          const code = envelope.code
          if (code === 200) {
            resolve(extractData(envelope))
            return
          }

          const msg = envelope.message || envelope.msg || '请求失败'
          if (code === 401) {
            clearToken()
            wx.showToast({ title: '请重新登录', icon: 'none' })
            wx.reLaunch({ url: '/pages/me/me' })
            reject(new Error('unauthorized'))
            return
          }
          wx.showToast({ title: msg, icon: 'none' })
          reject(new Error(msg))
          return
        }

        resolve(extractData(res.data))
      },
      fail: err => {
        const msg = normalizeNetworkErrorMessage(err)
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

