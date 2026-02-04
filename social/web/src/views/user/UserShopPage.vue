<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import {
  cancelRedeemOrder,
  createRedeemOrder,
  getPointsBalance,
  getRedeemOrder,
  listGoods,
  listRedeemOrders,
} from '../../api'
import { useToastStore } from '../../stores/toast'

type GoodsItem = {
  id?: number
  name?: string
  coverUrl?: string
  pointsPrice?: number
  stock?: number
  status?: number
  [k: string]: unknown
}

type OrderItem = {
  id?: number
  orderNo?: string
  status?: string
  totalPoints?: number
  createdAt?: string
  items?: unknown
  [k: string]: unknown
}

type PointsBalance = { balance?: number; [k: string]: unknown }

const toast = useToastStore()
const route = useRoute()

const loading = ref(false)
const goodsLoading = ref(false)
const ordersLoading = ref(false)
const balanceLoading = ref(false)

const goods = ref<GoodsItem[]>([])
const orders = ref<OrderItem[]>([])
const pointsBalance = ref(0)

type TabKey = 'goods' | 'orders'
const activeTab = ref<TabKey>('goods')

watch(
  () => route.query.tab,
  q => {
    const raw = Array.isArray(q) ? q[0] : q
    const v = typeof raw === 'string' ? raw : ''
    if (v === 'orders') activeTab.value = 'orders'
    else if (v === 'goods') activeTab.value = 'goods'
  },
  { immediate: true }
)

const goodsKeyword = ref('')

type OrderStatusFilter = 'ALL' | 'CREATED' | 'USED' | 'CANCELED'
const orderFilter = ref<OrderStatusFilter>('ALL')
const expandedOrderIds = ref<Record<string, boolean>>({})
const orderDetailLoading = ref<Record<string, boolean>>({})
const orderDetails = ref<Record<string, OrderItem | undefined>>({})

const safeStr = (v: unknown) => (v === undefined || v === null ? '' : String(v))

const loadBalance = async () => {
  balanceLoading.value = true
  try {
    const res = (await getPointsBalance()) as PointsBalance
    pointsBalance.value = typeof res.balance === 'number' ? res.balance : 0
  } catch {
    pointsBalance.value = 0
  } finally {
    balanceLoading.value = false
  }
}

const loadGoods = async () => {
  goodsLoading.value = true
  try {
    goods.value = (await listGoods(0, 50)) as GoodsItem[]
  } catch (e) {
    const err = e as { message?: unknown }
    goods.value = []
    toast.show((typeof err.message === 'string' && err.message) || '加载商品失败', 'error')
  } finally {
    goodsLoading.value = false
  }
}

const loadOrders = async () => {
  ordersLoading.value = true
  try {
    orders.value = (await listRedeemOrders(0, 50)) as OrderItem[]
  } catch (e) {
    const err = e as { message?: unknown }
    orders.value = []
    if (err.message === 'unauthorized') return
    toast.show((typeof err.message === 'string' && err.message) || '加载订单失败', 'error')
  } finally {
    ordersLoading.value = false
  }
}

const refresh = async () => {
  loading.value = true
  try {
    await Promise.all([loadGoods(), loadOrders(), loadBalance()])
  } finally {
    loading.value = false
  }
}

const statusMeta = (statusRaw: unknown) => {
  const s = String(statusRaw || '').toUpperCase()
  if (s === 'CREATED') return { label: '待使用', cls: 'badge badge--primary' }
  if (s === 'USED') return { label: '已核销', cls: 'badge badge--success' }
  if (s === 'CANCELED') return { label: '已取消', cls: 'badge badge--danger' }
  return { label: s || '-', cls: 'badge badge--muted' }
}

const goodsFiltered = computed(() => {
  const kw = goodsKeyword.value.trim().toLowerCase()
  if (!kw) return goods.value
  return goods.value.filter(g => safeStr(g.name).toLowerCase().includes(kw) || safeStr(g.id).includes(kw))
})

const orderCounts = computed(() => {
  const all = orders.value.length
  const created = orders.value.filter(o => String(o.status || '').toUpperCase() === 'CREATED').length
  const used = orders.value.filter(o => String(o.status || '').toUpperCase() === 'USED').length
  const canceled = orders.value.filter(o => String(o.status || '').toUpperCase() === 'CANCELED').length
  return { all, created, used, canceled }
})

const ordersFiltered = computed(() => {
  if (orderFilter.value === 'ALL') return orders.value
  const f = orderFilter.value
  return orders.value.filter(o => String(o.status || '').toUpperCase() === f)
})

const toggleOrder = async (id: number) => {
  const key = String(id)
  const next = !expandedOrderIds.value[key]
  expandedOrderIds.value = { ...expandedOrderIds.value, [key]: next }
  if (!next) return
  if (orderDetails.value[key]) return
  if (orderDetailLoading.value[key]) return
  orderDetailLoading.value = { ...orderDetailLoading.value, [key]: true }
  try {
    const detail = (await getRedeemOrder(id)) as OrderItem
    orderDetails.value = { ...orderDetails.value, [key]: detail }
  } catch {
    toast.show('加载订单详情失败', 'error')
  } finally {
    const { [key]: _, ...rest } = orderDetailLoading.value
    orderDetailLoading.value = rest
  }
}

const redeemGoods = async (goodsId: number) => {
  const g = goods.value.find(x => Number(x.id) === goodsId)
  const pointsPrice = g && typeof g.pointsPrice === 'number' ? g.pointsPrice : NaN
  if (!Number.isFinite(pointsPrice)) {
    toast.show('商品价格异常', 'error')
    return
  }
  const stock = g && typeof g.stock === 'number' ? g.stock : NaN
  if (Number.isFinite(stock) && stock <= 0) {
    toast.show('库存不足', 'error')
    return
  }
  try {
    const item = { goodsId, quantity: 1, pointsPrice }
    const order = await createRedeemOrder([item])
    const orderNo = typeof order.orderNo === 'string' ? order.orderNo : ''
    toast.show(orderNo ? `兑换成功：${orderNo}` : '兑换成功', 'success')
    activeTab.value = 'orders'
    await Promise.all([loadOrders(), loadBalance(), loadGoods()])
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '兑换失败', 'error')
  }
}

const cancelOrder = async (id: number) => {
  const okConfirm = window.confirm('确认取消该订单？')
  if (!okConfirm) return
  try {
    await cancelRedeemOrder(id)
    toast.show('已取消', 'success')
    await Promise.all([loadOrders(), loadBalance()])
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '取消失败', 'error')
  }
}

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="grid">
    <div class="row">
      <div class="title">积分商城</div>
      <div class="spacer" />
      <button class="btn btn--ghost" :disabled="loading" @click="refresh">刷新</button>
    </div>

    <div class="card section">
      <div class="row section__head">
        <div>
          <div class="title">概览</div>
          <div class="muted" style="margin-top: 6px">查看积分概览、商品列表与订单状态。</div>
        </div>
        <div class="spacer" />
        <div class="chip">
          <span class="muted">积分</span>
          <span style="margin-left: 8px; font-weight: 800">{{ balanceLoading ? '…' : pointsBalance }}</span>
        </div>
      </div>
      <div class="grid" style="grid-template-columns: repeat(3, minmax(0, 1fr)); margin-top: 10px">
        <div class="card card--flat metric">
          <div class="muted">可兑换商品</div>
          <div class="metric__val">{{ goods.length }}</div>
        </div>
        <div class="card card--flat metric">
          <div class="muted">订单总数</div>
          <div class="metric__val">{{ orders.length }}</div>
        </div>
        <div class="card card--flat metric">
          <div class="muted">待使用</div>
          <div class="metric__val">{{ orderCounts.created }}</div>
        </div>
      </div>
    </div>

    <div class="tabs">
      <button class="tab" :class="{ 'tab--active': activeTab === 'goods' }" @click="activeTab = 'goods'">商品区</button>
      <button class="tab" :class="{ 'tab--active': activeTab === 'orders' }" @click="activeTab = 'orders'">订单中心</button>
    </div>

    <div v-show="activeTab === 'goods'" class="card section">
      <div class="row section__head">
        <div>
          <div class="title">商品区</div>
          <div class="muted" style="margin-top: 6px">点击兑换会创建订单（默认数量 1）。</div>
        </div>
        <div class="spacer" />
        <button class="btn btn--ghost" :disabled="goodsLoading" @click="loadGoods">刷新商品</button>
      </div>
      <div class="row" style="margin-top: 10px">
        <input v-model="goodsKeyword" class="input" placeholder="搜索商品名称 / ID" />
      </div>
      <div class="grid" style="margin-top: 10px; grid-template-columns: repeat(2, minmax(0, 1fr))">
        <div v-if="goodsLoading" class="card muted" style="grid-column: 1 / -1">加载中…</div>
        <div v-else-if="goodsFiltered.length === 0" class="card muted" style="grid-column: 1 / -1">暂无商品</div>
        <div v-for="g in goodsFiltered" :key="g.id || g.name" class="card card--flat product">
          <div class="row" style="align-items: flex-start">
            <div class="product__cover">
              <img v-if="g.coverUrl" :src="g.coverUrl" alt="cover" />
              <div v-else class="product__coverPh">{{ safeStr(g.name).slice(0, 1) || '兑' }}</div>
            </div>
            <div class="grid" style="gap: 8px; flex: 1">
              <div class="row">
                <div class="title">{{ g.name || `#${g.id}` }}</div>
                <div class="spacer" />
                <span class="badge badge--muted">库存 {{ g.stock ?? '-' }}</span>
              </div>
              <div class="row" style="gap: 8px">
                <span class="badge badge--primary">{{ g.pointsPrice ?? '-' }} 积分</span>
                <span class="badge badge--muted">ID {{ g.id ?? '-' }}</span>
              </div>
              <div class="row" style="gap: 10px; margin-top: 2px">
                <button class="btn" :disabled="!g.id" @click="g.id && redeemGoods(g.id)">立即兑换</button>
                <button class="btn btn--ghost" @click="activeTab = 'orders'">去订单中心</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'orders'" class="card section">
      <div class="row section__head">
        <div>
          <div class="title">订单中心</div>
          <div class="muted" style="margin-top: 6px">按状态管理：待使用 / 已核销 / 已取消。</div>
        </div>
        <div class="spacer" />
        <button class="btn btn--ghost" :disabled="ordersLoading" @click="loadOrders">刷新订单</button>
      </div>

      <div class="seg" style="margin-top: 10px">
        <button class="seg__item" :class="{ 'seg__item--active': orderFilter === 'ALL' }" @click="orderFilter = 'ALL'">
          全部 <span class="seg__count">{{ orderCounts.all }}</span>
        </button>
        <button class="seg__item" :class="{ 'seg__item--active': orderFilter === 'CREATED' }" @click="orderFilter = 'CREATED'">
          待使用 <span class="seg__count">{{ orderCounts.created }}</span>
        </button>
        <button class="seg__item" :class="{ 'seg__item--active': orderFilter === 'USED' }" @click="orderFilter = 'USED'">
          已核销 <span class="seg__count">{{ orderCounts.used }}</span>
        </button>
        <button class="seg__item" :class="{ 'seg__item--active': orderFilter === 'CANCELED' }" @click="orderFilter = 'CANCELED'">
          已取消 <span class="seg__count">{{ orderCounts.canceled }}</span>
        </button>
      </div>

      <div class="grid" style="margin-top: 12px">
        <div v-if="ordersLoading" class="card muted">加载中…</div>
        <div v-else-if="ordersFiltered.length === 0" class="card muted">暂无订单</div>
        <div v-for="o in ordersFiltered" :key="o.id || o.orderNo" class="card card--flat order">
          <div class="row">
            <div class="grid" style="gap: 6px">
              <div class="row" style="gap: 8px">
                <div class="title">{{ o.orderNo || `订单#${o.id}` }}</div>
                <span :class="statusMeta(o.status).cls">{{ statusMeta(o.status).label }}</span>
              </div>
              <div class="muted">
                总积分：{{ o.totalPoints ?? '-' }} <span v-if="o.createdAt">· {{ o.createdAt }}</span>
              </div>
            </div>
            <div class="spacer" />
            <button class="btn btn--ghost" :disabled="!o.id" @click="o.id && toggleOrder(o.id)">
              {{ expandedOrderIds[String(o.id)] ? '收起' : '详情' }}
            </button>
            <button
              v-if="String(o.status || '').toUpperCase() === 'CREATED'"
              class="btn btn--danger"
              :disabled="!o.id"
              @click="o.id && cancelOrder(o.id)"
            >
              取消
            </button>
          </div>

          <div v-if="o.id && expandedOrderIds[String(o.id)]" class="grid" style="margin-top: 10px">
            <div v-if="orderDetailLoading[String(o.id)]" class="muted">加载详情中…</div>
            <div v-else class="grid" style="gap: 10px">
              <div class="row">
                <span class="badge badge--muted">订单ID {{ o.id }}</span>
                <span class="badge badge--muted">状态 {{ safeStr(o.status).toUpperCase() || '-' }}</span>
              </div>
              <div class="card card--flat" style="background: rgba(255, 255, 255, 0.03)">
                <div class="title" style="margin-bottom: 8px">明细</div>
                <div class="muted" v-if="!orderDetails[String(o.id)]">暂无明细</div>
                <div v-else class="grid" style="gap: 8px">
                  <div
                    v-for="(it, idx) in (orderDetails[String(o.id)]?.items as any[]) || []"
                    :key="idx"
                    class="row"
                    style="justify-content: space-between"
                  >
                    <div class="muted">goodsId {{ it.goodsId ?? '-' }} · 数量 {{ it.quantity ?? '-' }}</div>
                    <div style="font-weight: 700">{{ it.pointsPrice ?? '-' }} 积分</div>
                  </div>
                </div>
              </div>
              <div class="row" style="justify-content: flex-end; gap: 10px">
                <button class="btn btn--ghost" :disabled="ordersLoading" @click="loadOrders">刷新状态</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
