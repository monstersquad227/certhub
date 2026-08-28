<template>
  <div class="page-container balance-page">
    <div class="balance-top">
      <div class="page-header">
        <h1 class="page-title">余额中心</h1>
        <p class="page-subtitle">为账户余额充值，用于支付证书申请与续签费用</p>
      </div>

      <div class="summary-card">
      <div class="summary-left">
        <div class="summary-label">账户余额</div>
        <div class="summary-balance">
          <span class="balance-amount">¥ {{ displayBalance }}</span>
          <span class="balance-currency">CNY</span>
        </div>
      </div>
      <div class="summary-right">
        <div class="summary-label">本月支出</div>
        <div class="monthly-expense">¥ {{ monthlyExpense.toFixed(2) }}</div>
      </div>
    </div>

    <section class="section">
      <h3 class="section-title">选择充值金额</h3>
      <div class="amount-grid">
        <button
          v-for="option in amountOptions"
          :key="option.value"
          type="button"
          class="amount-option"
          :class="{ selected: selectedAmount === option.value && !isCustomAmount }"
          @click="selectAmount(option.value)"
        >
          <span class="amount-value">¥ {{ option.value }}</span>
          <span v-if="option.bonus" class="amount-bonus">赠 ¥{{ option.bonus }} 体验金</span>
        </button>
        <button
          type="button"
          class="amount-option custom-option"
          :class="{ selected: isCustomAmount }"
          @click="selectCustom"
        >
          <span class="custom-label">自定义</span>
          <div class="custom-input-wrap" @click.stop>
            <span class="custom-prefix">¥</span>
            <input
              ref="customInputRef"
              v-model="customAmountStr"
              type="number"
              min="1"
              step="0.01"
              class="custom-input"
              placeholder="输入金额"
              @focus="selectCustom"
              @input="onCustomInput"
            />
          </div>
        </button>
      </div>
    </section>

    <section class="section">
      <h3 class="section-title">支付方式</h3>
      <div class="payment-grid">
        <button
          v-for="method in paymentMethods"
          :key="method.value"
          type="button"
          class="payment-option"
          :class="{ selected: paymentMethod === method.value }"
          @click="paymentMethod = method.value"
        >
          <img
            :src="method.iconUrl"
            :alt="method.label"
            class="payment-icon-img"
          />
        </button>
      </div>
    </section>

    <div class="submit-row">
      <a-button
        type="primary"
        size="large"
        class="submit-btn"
        :loading="submitting"
        :disabled="!rechargeAmount"
        @click="handleRecharge"
      >
        确认充值 ¥{{ rechargeAmountDisplay }}
      </a-button>
    </div>
    </div>

    <a-modal
      v-model:open="alipayQrVisible"
      :footer="null"
      :width="400"
      centered
      destroy-on-close
    >
      <div class="alipay-qr-modal">
        <p class="alipay-qr-hint">请使用支付宝扫描下方二维码完成付款</p>
        <p class="alipay-qr-amount">¥ {{ rechargeAmountDisplay }}</p>
        <img :src="alipayQrImg" alt="支付宝收款二维码" class="alipay-qr-img" />
        <a-button
          type="primary"
          size="large"
          class="alipay-qr-confirm"
          :loading="submitting"
          block
          @click="confirmAlipayPaid"
        >
          我已完成付款
        </a-button>
      </div>
    </a-modal>

    <section class="section records-section">
      <h3 class="section-title">充值记录</h3>
      <div class="table-wrapper">
        <a-table
          :columns="columns"
          :data-source="dataSource"
          :loading="loading"
          :pagination="pagination"
          :scroll="tableScroll"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'amount'">
              <span class="record-amount">+ ¥{{ Math.abs(record.amount).toFixed(2) }}</span>
            </template>
            <template v-else-if="column.key === 'payment_method'">
              {{ formatPaymentMethod(record.payment_method) }}
            </template>
            <template v-else-if="column.key === 'created_at'">
              {{ formatDateTime(record.created_at) }}
            </template>
            <template v-else-if="column.key === 'status'">
              <span class="status-badge success">已到账</span>
            </template>
          </template>
        </a-table>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { message } from 'ant-design-vue'
import api from '@/utils/api'
import { useAuthStore } from '@/store/auth'
import { useBalanceStore } from '@/store/balance'
import { formatDateTime } from '@/utils/date'
import type { TableColumnsType, TableProps } from 'ant-design-vue'
import alipayQrImg from '@/assets/alipay.jpg'
import alipayIcon from '@/assets/ALIPAY_CN.svg'

const auth = useAuthStore()
const balanceStore = useBalanceStore()

const loading = ref(false)
const submitting = ref(false)
const alipayQrVisible = ref(false)
const dataSource = ref<any[]>([])
const monthlyExpense = ref(0)
const tableScroll = { x: 800 }

const selectedAmount = ref(200)
const isCustomAmount = ref(false)
const customAmountStr = ref('')
const customInputRef = ref<HTMLInputElement>()
const paymentMethod = ref('alipay')

const amountOptions = [
  { value: 50, bonus: null },
  { value: 200, bonus: 10 },
  { value: 500, bonus: 40 },
]

const paymentMethods = [
  { value: 'alipay', label: '支付宝', iconUrl: alipayIcon },
]

const columns: TableColumnsType = [
  { title: '时间', dataIndex: 'created_at', key: 'created_at', width: 190 },
  { title: '金额', key: 'amount', width: 120 },
  { title: '方式', key: 'payment_method', width: 100 },
  { title: '流水号', dataIndex: 'order_no', key: 'order_no' },
  { title: '状态', key: 'status', width: 100 },
]

const pagination = reactive({
  current: 1,
  pageSize: 3,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
})

const displayBalance = computed(() => {
  const value = balanceStore.balance ?? auth.user?.balance ?? 0
  return value.toFixed(2)
})

const rechargeAmount = computed(() => {
  if (isCustomAmount.value) {
    const val = parseFloat(customAmountStr.value)
    return val > 0 ? val : null
  }
  return selectedAmount.value
})

const rechargeAmountDisplay = computed(() => {
  return rechargeAmount.value ? rechargeAmount.value.toFixed(2) : '0.00'
})

function selectAmount(value: number) {
  isCustomAmount.value = false
  customAmountStr.value = ''
  selectedAmount.value = value
}

function selectCustom() {
  isCustomAmount.value = true
  nextTick(() => customInputRef.value?.focus())
}

function onCustomInput() {
  isCustomAmount.value = true
}

function formatPaymentMethod(method: string) {
  if (method === 'alipay') return '支付宝'
  if (method === 'wechat') return '微信'
  return method || '-'
}

async function fetchMonthlyExpense() {
  try {
    const res = await api.get('/api/v1/balance/records', {
      params: { type: 'consume', page: 1, page_size: 100 },
    })
    const now = new Date()
    const monthStart = new Date(now.getFullYear(), now.getMonth(), 1)
    const list = res.data.data.list || []
    monthlyExpense.value = list
      .filter((r: any) => new Date(r.created_at) >= monthStart)
      .reduce((sum: number, r: any) => sum + Math.abs(r.amount), 0)
  } catch {
    monthlyExpense.value = 0
  }
}

async function fetchRecords() {
  loading.value = true
  try {
    const res = await api.get('/api/v1/balance/records', {
      params: {
        type: 'recharge',
        page: pagination.current,
        page_size: pagination.pageSize,
      },
    })
    dataSource.value = res.data.data.list
    pagination.total = res.data.data.total
  } catch (error: any) {
    message.error(error.response?.data?.message || '获取充值记录失败')
  } finally {
    loading.value = false
  }
}

async function submitRecharge() {
  submitting.value = true
  try {
    const res = await api.post('/api/v1/balance/recharge', {
      amount: rechargeAmount.value,
      payment_method: paymentMethod.value,
    })
    alipayQrVisible.value = false
    message.success(`充值订单创建成功，订单号：${res.data.data.order_no}`)
    setTimeout(async () => {
      message.success('充值成功！')
      await balanceStore.fetchBalance()
      await fetchMonthlyExpense()
      pagination.current = 1
      await fetchRecords()
    }, 1000)
  } catch (error: any) {
    message.error(error.response?.data?.message || '创建充值订单失败')
  } finally {
    submitting.value = false
  }
}

function handleRecharge() {
  if (!rechargeAmount.value) {
    message.warning('请选择或输入充值金额')
    return
  }
  if (paymentMethod.value === 'alipay') {
    alipayQrVisible.value = true
    return
  }
  submitRecharge()
}

function confirmAlipayPaid() {
  submitRecharge()
}

const handleTableChange: TableProps['onChange'] = (pag) => {
  if (pag) {
    pagination.current = pag.current || 1
    pagination.pageSize = pag.pageSize || 3
  }
  fetchRecords()
}

onMounted(() => {
  balanceStore.fetchBalance()
  fetchMonthlyExpense()
  fetchRecords()
})
</script>

<style scoped>
.balance-page {
  height: 100%;
  max-height: 100%;
  overflow-x: hidden;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  padding: 0;
  box-sizing: border-box;
}

.balance-top {
  flex-shrink: 0;
}

.page-header {
  margin-bottom: 16px;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0 0 4px;
  line-height: 1.3;
}

.page-subtitle {
  font-size: 13px;
  color: #888;
  margin: 0;
}

.summary-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  background: #fff;
  border: 1px solid #e8ecef;
  border-radius: 12px;
  padding: 16px 24px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.summary-label {
  font-size: 13px;
  color: #888;
  margin-bottom: 8px;
  font-weight: 500;
}

.summary-balance {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.balance-amount {
  font-size: 28px;
  font-weight: 700;
  color: #1a1a1a;
  letter-spacing: -0.5px;
}

.balance-currency {
  font-size: 14px;
  color: #888;
  font-weight: 500;
}

.summary-right {
  text-align: right;
  flex-shrink: 0;
}

.monthly-expense {
  font-size: 20px;
  font-weight: 600;
  color: #1a1a1a;
}

.section {
  margin-bottom: 14px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 10px;
}

.amount-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.amount-option {
  background: #fff;
  border: 1.5px solid #e8ecef;
  border-radius: 10px;
  padding: 12px 14px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  text-align: left;
  min-width: 0;
  width: 100%;
}

.amount-option:hover {
  border-color: #b7eb8f;
}

.amount-option.selected {
  border-color: #52c41a;
  background: #f6ffed;
}

.amount-value {
  font-size: 18px;
  font-weight: 700;
  color: #1a1a1a;
}

.amount-bonus {
  font-size: 12px;
  color: #52c41a;
  font-weight: 500;
}

.custom-option {
  justify-content: center;
}

.custom-label {
  font-size: 12px;
  color: #888;
}

.custom-input-wrap {
  display: flex;
  align-items: center;
  gap: 4px;
  width: 100%;
  min-width: 0;
}

.custom-prefix {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  flex-shrink: 0;
}

.custom-input {
  border: none;
  outline: none;
  background: transparent;
  font-size: 16px;
  color: #1a1a1a;
  width: 100%;
  min-width: 0;
  font-weight: 500;
}

.custom-input::placeholder {
  color: #bbb;
  font-weight: 400;
}

.payment-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  max-width: 340px;
  width: 100%;
}

.payment-option {
  background: #fff;
  border: 1.5px solid #e8ecef;
  border-radius: 10px;
  padding: 16px 20px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  text-align: center;
  min-width: 0;
  width: 100%;
  min-height: 72px;
}

.payment-option:hover {
  border-color: #91caff;
}

.payment-option.selected {
  border-color: #3b82f6;
  background: #f0f7ff;
}

.payment-icon {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  color: #555;
  margin-bottom: 4px;
}

.payment-icon-img {
  height: 52px;
  width: auto;
  max-width: 140px;
  object-fit: contain;
  display: block;
}

.payment-option.selected .payment-icon {
  background: #dbeafe;
  color: #3b82f6;
}

.payment-name {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
}

.payment-desc {
  font-size: 12px;
  color: #888;
}

.submit-row {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 0;
}

.submit-btn {
  border-radius: 10px;
  height: 40px;
  padding: 0 32px;
  font-size: 15px;
  font-weight: 600;
}

.alipay-qr-modal {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0 4px;
  text-align: center;
}

.alipay-qr-hint {
  margin: 0 0 8px;
  font-size: 14px;
  color: #666;
}

.alipay-qr-amount {
  margin: 0 0 16px;
  font-size: 28px;
  font-weight: 700;
  color: #1a1a1a;
  letter-spacing: -0.5px;
}

.alipay-qr-img {
  width: 240px;
  height: 240px;
  object-fit: contain;
  border-radius: 8px;
  background: #fafafa;
}

.alipay-qr-confirm {
  margin-top: 20px;
  border-radius: 10px;
  height: 40px;
  font-weight: 600;
}

.records-section {
  flex: 1 1 auto;
  min-height: 180px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin-bottom: 0;
  padding-top: 8px;
}

.table-wrapper {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-wrapper :deep(.ant-table-wrapper) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.table-wrapper :deep(.ant-spin-nested-loading),
.table-wrapper :deep(.ant-spin-container) {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.table-wrapper :deep(.ant-table) {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 12px;
}

.table-wrapper :deep(.ant-table-container) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.table-wrapper :deep(.ant-table-body) {
  flex: 1;
  min-height: 0;
  overflow-y: auto !important;
}

.table-wrapper :deep(.ant-table-thead) {
  flex-shrink: 0;
}

.table-wrapper :deep(.ant-table-thead > tr > th) {
  background: #fafafa;
  color: #888;
  font-weight: 500;
  font-size: 13px;
  border-bottom: 1px solid #f0f0f0;
}

.table-wrapper :deep(.ant-table-tbody > tr > td) {
  border-bottom: 1px solid #f5f5f5;
  font-size: 14px;
}

.record-amount {
  color: #3f8600;
  font-weight: 600;
}

.status-badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.success {
  color: #52c41a;
  background: #f6ffed;
  border: 1px solid #b7eb8f;
}

@media (max-width: 1100px) {
  .amount-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .payment-grid {
    max-width: none;
  }
}

@media (max-width: 768px) {
  .balance-page {
    padding: 0;
  }

  .page-title {
    font-size: 20px;
  }

  .summary-card {
    flex-direction: column;
    gap: 16px;
    padding: 14px 16px;
  }

  .summary-right {
    text-align: left;
  }

  .balance-amount {
    font-size: 24px;
  }

  .amount-grid,
  .payment-grid {
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  .submit-row {
    justify-content: stretch;
  }

  .submit-btn {
    width: 100%;
  }

  .records-section {
    min-height: 220px;
  }
}

@media (max-width: 480px) {
  .amount-grid,
  .payment-grid {
    grid-template-columns: 1fr;
  }
}

/* 高度不足时改为整页滚动，避免内容被裁切 */
@media (max-height: 820px) {
  .balance-page {
    height: auto;
    max-height: none;
    overflow: visible;
  }

  .records-section {
    flex: none;
    min-height: 240px;
    overflow: visible;
  }

  .table-wrapper {
    overflow: visible;
  }

  .table-wrapper :deep(.ant-table-body) {
    overflow: visible !important;
  }
}
</style>
