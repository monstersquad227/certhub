<template>
  <div class="page-container">
    <a-card title="证书管理" class="card-modern">
      <div class="filter-bar">
        <a-input
          v-model:value="searchUserEmail"
          placeholder="用户邮箱"
          class="filter-input"
          allow-clear
        />
        <a-input
          v-model:value="searchDomain"
          placeholder="域名"
          class="filter-input"
          allow-clear
        />
        <a-select
          v-model:value="filterStatus"
          placeholder="证书状态"
          class="filter-select"
          allow-clear
        >
          <a-select-option value="valid">有效</a-select-option>
          <a-select-option value="expiring">即将过期</a-select-option>
          <a-select-option value="expired">已过期</a-select-option>
        </a-select>
        <a-range-picker v-model:value="dateRange" class="filter-range-picker" />
        <a-button type="primary" @click="handleSearch" class="action-btn">搜索</a-button>
        <a-button @click="handleReset" class="action-btn">重置</a-button>
        <a-button type="primary" @click="$router.push('/admin/certificates/create')" class="action-btn">
          新增证书
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
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="handleEdit(record.id)">
              编辑
            </a-button>
            <a-popconfirm title="确定要删除这个证书吗？" @confirm="handleDelete(record.id)">
              <a-button type="link" size="small" danger>删除</a-button>
            </a-popconfirm>
          </a-space>
        </template>
      </template>
        </a-table>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import api from '@/utils/api'
import type { TableColumnsType, TableProps } from 'ant-design-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { getStatusColor, getStatusText, getCertTypeColor, getCertTypeText } from '@/utils/certificate'

const router = useRouter()
const loading = ref(false)
const searchUserEmail = ref('')
const searchDomain = ref('')
const filterStatus = ref('')
const dateRange = ref<[Dayjs, Dayjs] | null>(null)
const dataSource = ref([])

const columns: TableColumnsType = [
  {
    title: '证书ID',
    dataIndex: 'id',
    key: 'id',
    width: 100,
  },
  {
    title: '用户邮箱',
    dataIndex: 'user_email',
    key: 'user_email',
  },
  {
    title: '域名',
    dataIndex: 'domain',
    key: 'domain',
  },
  {
    title: '证书类型',
    key: 'cert_type',
    width: 120,
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
    width: 120,
  },
  {
    title: '操作',
    key: 'action',
    width: 150,
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
    if (searchUserEmail.value) {
      params.user_email = searchUserEmail.value
    }
    if (searchDomain.value) {
      params.domain = searchDomain.value
    }
    if (filterStatus.value) {
      params.status = filterStatus.value
    }
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_time = dateRange.value[0].toISOString()
      params.end_time = dateRange.value[1].toISOString()
    }
    const res = await api.get('/api/v1/admin/certificates', { params })
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
  searchUserEmail.value = ''
  searchDomain.value = ''
  filterStatus.value = ''
  dateRange.value = null
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

const handleEdit = (id: number) => {
  router.push(`/admin/certificates/${id}/edit`)
}

const handleDelete = async (id: number) => {
  try {
    await api.delete(`/api/v1/admin/certificates/${id}`)
    message.success('删除成功')
    fetchData()
  } catch (error: any) {
    message.error(error.response?.data?.message || '删除失败')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.filter-bar {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.filter-input {
  width: 200px;
  flex-shrink: 0;
}

.filter-input :deep(.ant-input) {
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  line-height: 40px;
}

.filter-select {
  width: 150px;
  flex-shrink: 0;
}

.filter-select :deep(.ant-select-selector) {
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
}

.filter-select :deep(.ant-select-selection-item),
.filter-select :deep(.ant-select-selection-placeholder) {
  line-height: 40px;
}

.filter-range-picker :deep(.ant-picker) {
  height: 40px;
  border-radius: 12px;
}

.action-btn {
  border-radius: 12px;
  height: 40px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
</style>

