<template>
  <div class="page-container balance-page">
    <a-card title="余额管理" class="card-modern balance-card-container">
      <div class="balance-content">
        <div class="balance-section">
          <div class="balance-card">
            <div class="balance-header">
              <div class="balance-info">
                <div class="balance-label">当前余额</div>
                <div class="balance-value">¥ {{ balance.toFixed(2) }}</div>
              </div>
              <a-button type="primary" @click="$router.push('/balance/recharge')" class="recharge-btn">
                充值
              </a-button>
            </div>
          </div>
        </div>

        <a-divider class="divider-spaced" />

        <div class="records-section">
          <div class="section-title">余额变动记录</div>
          <div class="filter-bar">
            <a-select
              v-model:value="filterType"
              placeholder="类型"
              class="filter-select"
              allow-clear
              @change="handleSearch"
            >
              <a-select-option value="recharge">充值</a-select-option>
              <a-select-option value="consume">消费</a-select-option>
            </a-select>
            <a-button @click="handleReset" class="action-btn">重置</a-button>
          </div>

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
          <template v-if="column.key === 'type'">
            <a-tag :color="record.type === 'recharge' ? 'green' : 'red'">
              {{ record.type === 'recharge' ? '充值' : '消费' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'amount'">
            <span :style="{ color: record.amount > 0 ? '#3f8600' : '#cf1322' }">
              {{ record.amount > 0 ? '+' : '' }}{{ record.amount.toFixed(2) }} 元
            </span>
          </template>
          <template v-else-if="column.key === 'payment_method'">
            <span v-if="record.payment_method">
              {{ record.payment_method === 'alipay' ? '支付宝' : '微信' }}
            </span>
            <span v-else>-</span>
        </template>
      </template>
            </a-table>
          </div>
        </div>
      </div>
    </a-card>
  </div>
</template>

<style scoped>
.balance-page {
  max-height: calc(100vh - 112px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.balance-card-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.balance-card-container :deep(.ant-card-body) {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
  padding: 24px;
}

.balance-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.balance-section {
  flex-shrink: 0;
  margin-bottom: 20px;
}

.balance-card {
  background: linear-gradient(135deg, #f6f8fb 0%, #ffffff 100%);
  border: 1px solid #e8ecef;
  border-radius: 16px;
  padding: 20px 24px;
}

.balance-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.balance-info {
  flex: 1;
}

.balance-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
  font-weight: 500;
}

.balance-value {
  font-size: 32px;
  font-weight: 700;
  color: #3f8600;
  letter-spacing: -0.5px;
  margin: 0;
}

.recharge-btn {
  border-radius: 12px;
  height: 40px;
  padding: 0 32px;
  font-weight: 500;
  display: flex;
  align-items: center;
}

.divider-spaced {
  flex-shrink: 0;
  margin: 16px 0;
}

.records-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.filter-bar {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
  flex-shrink: 0;
}

.filter-select {
  width: 200px;
}

.filter-select :deep(.ant-select-selector) {
  border-radius: 12px;
  height: 40px;
  display: flex;
  align-items: center;
}

.filter-select :deep(.ant-select-selection-item),
.filter-select :deep(.ant-select-selection-placeholder) {
  line-height: 40px;
}

.action-btn {
  border-radius: 12px;
  height: 40px;
  padding: 0 20px;
  display: flex;
  align-items: center;
}

.table-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.table-wrapper :deep(.ant-table-wrapper) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.table-wrapper :deep(.ant-table) {
  flex: 1;
  display: flex;
  flex-direction: column;
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
  overflow-x: auto;
}

.table-wrapper :deep(.ant-table-thead) {
  flex-shrink: 0;
}

.table-wrapper :deep(.ant-pagination) {
  margin: 16px 0 0 0 !important;
  padding: 0;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

@media (max-height: 820px) {
  .balance-page {
    max-height: none;
    overflow: auto;
  }

  .balance-card-container {
    min-height: 0;
  }

  .records-section {
    overflow: visible;
  }

  .table-wrapper {
    min-height: 280px;
  }
}
</style>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import api from '@/utils/api'
import type { TableColumnsType, TableProps } from 'ant-design-vue'

const balance = ref(0)
const loading = ref(false)
const filterType = ref('')
const dataSource = ref([])
const tableScroll = { x: 900 }

const columns: TableColumnsType = [
  {
    title: '类型',
    key: 'type',
    width: 100,
  },
  {
    title: '金额',
    key: 'amount',
    width: 150,
  },
  {
    title: '支付方式',
    key: 'payment_method',
    width: 120,
  },
  {
    title: '订单号',
    dataIndex: 'order_no',
    key: 'order_no',
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description',
  },
  {
    title: '时间',
    dataIndex: 'created_at',
    key: 'created_at',
  },
]

const pagination = reactive({
  current: 1,
  pageSize: 5,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  hideOnSinglePage: false,
})

const fetchBalance = async () => {
  try {
    const res = await api.get('/api/v1/balance')
    balance.value = res.data.data.balance
  } catch (error: any) {
    message.error(error.response?.data?.message || '获取余额失败')
  }
}

const fetchRecords = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
    }
    if (filterType.value) {
      params.type = filterType.value
    }
    const res = await api.get('/api/v1/balance/records', { params })
    dataSource.value = res.data.data.list
    pagination.total = res.data.data.total
  } catch (error: any) {
    message.error(error.response?.data?.message || '获取余额记录失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  fetchRecords()
}

const handleReset = () => {
  filterType.value = ''
  pagination.current = 1
  fetchRecords()
}

const handleTableChange: TableProps['onChange'] = (pag) => {
  if (pag) {
    pagination.current = pag.current || 1
    pagination.pageSize = pag.pageSize || 10
  }
  fetchRecords()
}

onMounted(() => {
  fetchBalance()
  fetchRecords()
})
</script>

