<template>
  <div class="page-container">
    <a-card class="card-modern" :bordered="false">
      <div class="filter-bar">
        <a-input
          v-model:value="searchDomain"
          placeholder="搜索域名"
          class="filter-input"
          @pressEnter="handleSearch"
        />
        <a-select
          v-model:value="filterStatus"
          placeholder="证书状态"
          class="filter-select"
          allow-clear
          @change="handleSearch"
        >
          <a-select-option value="pending">生成中</a-select-option>
          <a-select-option value="valid">有效</a-select-option>
          <a-select-option value="expiring">即将过期</a-select-option>
          <a-select-option value="expired">已过期</a-select-option>
          <a-select-option value="failed">生成失败</a-select-option>
        </a-select>
        <a-button @click="handleSearch" class="action-btn">搜索</a-button>
        <a-button @click="handleReset" class="action-btn">重置</a-button>
        <a-button type="primary" @click="$router.push('/certificates/apply')" class="apply-btn">
          申请证书
        </a-button>
      </div>

      <div class="table-wrapper">
        <a-table
          :columns="columns"
          :data-source="dataSource"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'cert_type'">
          <a-tag :color="getCertTypeColor(record.cert_type)">
            {{ getCertTypeText(record.cert_type) }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'status'">
          <a-tag :color="getStatusColor(record.status)">
            {{ getStatusText(record.status) }}
          </a-tag>
          <a-tooltip v-if="record.status === 'failed' && extractErrorMessage(record)">
            <template #title>
              <div style="max-width: 300px; word-wrap: break-word">
                {{ extractErrorMessage(record) }}
              </div>
            </template>
            <QuestionCircleOutlined style="margin-left: 4px; color: #ff4d4f" />
          </a-tooltip>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="handleView(record.id)">
              查看详情
            </a-button>
          </a-space>
        </template>
      </template>
        </a-table>
      </div>
    </a-card>
  </div>
</template>

<style scoped>
.filter-bar {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: nowrap;
  max-width: 100%;
}

.filter-input {
  width: 200px;
  flex-shrink: 0;
  height: 40px;
}

.filter-input :deep(.ant-input) {
  border-radius: 12px;
  height: 40px;
  display: flex;
  align-items: center;
  line-height: 40px;
}

.filter-select {
  width: 200px;
  height: 40px;
  flex-shrink: 0;
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
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.apply-btn {
  border-radius: 12px;
  height: 40px;
  font-weight: 500;
  padding: 0 20px;
  margin-left: auto;
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.table-wrapper {
  margin-top: 0;
  width: 100%;
  overflow-x: auto;
}

:deep(.ant-btn-link) {
  border-radius: 8px;
  padding: 0 8px;
}

@media (max-width: 1200px) {
  .filter-bar {
    flex-wrap: wrap;
  }
  
  .apply-btn {
    margin-left: 0;
    width: 100%;
  }
}
</style>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { QuestionCircleOutlined } from '@ant-design/icons-vue'
import api from '@/utils/api'
import type { TableColumnsType, TableProps } from 'ant-design-vue'
import { getStatusColor, getStatusText, getCertTypeColor, getCertTypeText, extractErrorMessage } from '@/utils/certificate'

const router = useRouter()
const loading = ref(false)
const searchDomain = ref('')
const filterStatus = ref('')
const dataSource = ref([])

const columns: TableColumnsType = [
  {
    title: '域名',
    dataIndex: 'domain',
    key: 'domain',
  },
  {
    title: '证书类型',
    key: 'cert_type',
  },
  {
    title: '申请时间',
    dataIndex: 'created_at',
    key: 'created_at',
  },
  {
    title: '过期时间',
    dataIndex: 'expires_at',
    key: 'expires_at',
  },
  {
    title: '状态',
    key: 'status',
  },
  {
    title: '操作',
    key: 'action',
    width: 120,
  },
]

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
})

const fetchData = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
    }
    if (searchDomain.value) {
      params.domain = searchDomain.value
    }
    if (filterStatus.value) {
      params.status = filterStatus.value
    }
    const res = await api.get('/api/v1/certificates', { params })
    dataSource.value = res.data.data.list
    pagination.total = res.data.data.total
  } catch (error: any) {
    message.error(error.response?.data?.message || '获取证书列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  fetchData()
}

const handleReset = () => {
  searchDomain.value = ''
  filterStatus.value = ''
  pagination.current = 1
  fetchData()
}

const handleTableChange: TableProps['onChange'] = (pag) => {
  if (pag) {
    pagination.current = pag.current || 1
    pagination.pageSize = pag.pageSize || 10
  }
  fetchData()
}

const handleView = (id: number) => {
  router.push(`/certificates/${id}`)
}

onMounted(() => {
  fetchData()
})
</script>
