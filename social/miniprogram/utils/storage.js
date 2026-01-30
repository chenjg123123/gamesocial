const STORAGE_KEYS = {
  token: 'gs_token',
  user: 'gs_user',
}

const getToken = () => wx.getStorageSync(STORAGE_KEYS.token)

const setToken = token => {
  wx.setStorageSync(STORAGE_KEYS.token, token)
}

const clearToken = () => {
  wx.removeStorageSync(STORAGE_KEYS.token)
}

const getUserCache = () => wx.getStorageSync(STORAGE_KEYS.user)

const setUserCache = user => {
  wx.setStorageSync(STORAGE_KEYS.user, user)
}

const clearUserCache = () => {
  wx.removeStorageSync(STORAGE_KEYS.user)
}

module.exports = {
  STORAGE_KEYS,
  getToken,
  setToken,
  clearToken,
  getUserCache,
  setUserCache,
  clearUserCache,
}

