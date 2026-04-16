<template>
  <div class="page-container-narrow">
    <a-card title="证书申请" class="card-modern">
      <a-form :model="form" :rules="rules" ref="formRef" layout="vertical" :label-col="{ span: 24 }">
        <a-form-item label="域名" name="domain">
          <a-input
            v-model:value="form.domain"
            placeholder="请输入域名，如 example.com 或 *.example.com"
            :disabled="step > 1"
            class="form-input"
          />
          <div class="text-hint">
            支持单域名（如 example.com）和泛域名（如 *.example.com）
          </div>
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            :loading="generatingDNS"
            @click="handleGenerateDNS"
            :disabled="step > 1"
            class="form-btn"
          >
            生成 DNS 记录
          </a-button>
        </a-form-item>
      </a-form>

      <a-divider v-if="dnsRecord" class="divider-spaced" />

      <div v-if="dnsRecord" class="content-section">
        <a-alert
          message="DNS 配置说明"
          description="请前往您的域名服务商（如阿里云、腾讯云等）添加以下 DNS 解析记录，配置完成后点击「生成证书」按钮。"
          type="info"
          show-icon
          class="alert-spaced"
        />

        <a-card title="DNS 记录信息" size="small" class="card-modern content-section">
          <a-descriptions :column="1" bordered>
            <a-descriptions-item label="记录类型">
              <a-tag color="blue">{{ dnsRecord.type }}</a-tag>
            </a-descriptions-item>
            <a-descriptions-item label="记录名称">
              <a-typography-text copyable>{{ dnsRecord.name }}</a-typography-text>
            </a-descriptions-item>
            <a-descriptions-item label="记录值">
              <a-typography-text copyable>{{ dnsRecord.value }}</a-typography-text>
            </a-descriptions-item>
          </a-descriptions>
        </a-card>

        <div class="button-group" style="margin-top: var(--spacing-lg)">
          <a-button
            type="primary"
            :loading="generatingCert"
            @click="handleGenerateCert"
            class="form-btn"
          >
            生成证书
          </a-button>
          <a-button @click="handleReset" class="form-btn">重新申请</a-button>
        </div>
      </div>

    <a-modal
      v-model:open="successModalVisible"
      title="证书申请成功"
      :footer="null"
      @ok="handleViewCert"
    >
      <p>证书申请成功！</p>
      <a-button type="primary" @click="handleViewCert" style="margin-top: 16px">
        查看证书详情
      </a-button>
    </a-modal>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import api from '@/utils/api'

const DNS_RECORD_KEY = 'CERT_DNS_RECORD'

const router = useRouter()
const formRef = ref()
const step = ref(1)
const generatingDNS = ref(false)
const generatingCert = ref(false)
const successModalVisible = ref(false)
const createdCertId = ref<number | null>(null)

const form = reactive({
  domain: '',
})

// 页面加载时恢复之前保存的 DNS 记录
onMounted(() => {
  const saved = localStorage.getItem(DNS_RECORD_KEY)
  if (saved) {
    try {
      const data = JSON.parse(saved)
      if (data.domain && data.dnsRecord) {
        form.domain = data.domain
        dnsRecord.value = data.dnsRecord
        step.value = 2
      }
    } catch (e) {
      localStorage.removeItem(DNS_RECORD_KEY)
    }
  }
})

const rules = {
  domain: [
    { required: true, message: '请输入域名', trigger: 'blur' },
    {
      validator: (_rule: any, value: string) => {
        if (!value) return Promise.resolve()
        const trimmed = value.trim()
        if (trimmed.startsWith('*.')) {
          const domain = trimmed.substring(2)
          if (!domain.includes('.') || domain.includes(' ')) {
            return Promise.reject('泛域名格式错误，应为 *.example.com')
          }
        } else {
          if (!trimmed.includes('.') || trimmed.includes(' ')) {
            return Promise.reject('域名格式错误')
          }
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
}

const dnsRecord = ref<{
  type: string
  name: string
  value: string
} | null>(null)

const handleGenerateDNS = async () => {
  try {
    await formRef.value.validate()
    generatingDNS.value = true
    const res = await api.post('/api/v1/certificates/generate-dns', {
      domain: form.domain.trim(),
    })
    dnsRecord.value = res.data.data.dns_record
    step.value = 2
    // 保存到 localStorage
    localStorage.setItem(DNS_RECORD_KEY, JSON.stringify({
      domain: form.domain.trim(),
      dnsRecord: dnsRecord.value,
      createdAt: Date.now()
    }))
    message.success('DNS 记录生成成功')
  } catch (error: any) {
    message.error(error.response?.data?.message || '生成 DNS 记录失败')
  } finally {
    generatingDNS.value = false
  }
}

const handleGenerateCert = async () => {
  if (!dnsRecord.value) return
  try {
    generatingCert.value = true
    const res = await api.post('/api/v1/certificates/generate', {
      domain: form.domain.trim(),
      dns_record: dnsRecord.value,
    })
    
    // 检查是否是异步处理
    if (res.data.data.status === 'processing') {
      message.info('证书正在生成中，请稍后在证书列表查看')
      // 清除保存的 DNS 记录
      localStorage.removeItem(DNS_RECORD_KEY)
      router.push('/certificates')
    } else {
      createdCertId.value = res.data.data.certificate.id
      successModalVisible.value = true
      message.success('证书生成成功')
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || '证书生成失败')
  } finally {
    generatingCert.value = false
  }
}

const handleReset = () => {
  form.domain = ''
  dnsRecord.value = null
  step.value = 1
  formRef.value?.resetFields()
  // 清除保存的 DNS 记录
  localStorage.removeItem(DNS_RECORD_KEY)
}

const handleViewCert = () => {
  successModalVisible.value = false
  // 清除保存的 DNS 记录
  localStorage.removeItem(DNS_RECORD_KEY)
  if (createdCertId.value) {
    router.push(`/certificates/${createdCertId.value}`)
  } else {
    router.push('/certificates')
  }
}
</script>

<style scoped>
.form-input :deep(.ant-input) {
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  line-height: 40px;
}

.form-btn {
  height: 40px;
  padding: 0 24px;
  border-radius: 12px;
  display: flex;
  align-items: center;
}

.form-btn:disabled,
.form-btn.disabled {
  background-color: #d9d9d9 !important;
  border-color: #d9d9d9 !important;
  color: rgba(0, 0, 0, 0.25) !important;
  cursor: not-allowed !important;
  opacity: 1 !important;
}

/* 针对 Ant Design 按钮的禁用状态 */
:deep(.ant-btn-primary.form-btn:disabled),
:deep(.ant-btn-primary.form-btn.disabled),
:deep(.form-btn.ant-btn-primary:disabled),
:deep(.form-btn.ant-btn-primary.disabled),
:deep(.ant-btn-primary:disabled.form-btn),
:deep(.ant-btn-primary.disabled.form-btn) {
  background: #d9d9d9 !important;
  background-color: #d9d9d9 !important;
  border-color: #d9d9d9 !important;
  color: rgba(0, 0, 0, 0.25) !important;
  cursor: not-allowed !important;
  opacity: 1 !important;
}

:deep(.ant-btn-primary.form-btn:disabled:hover),
:deep(.ant-btn-primary.form-btn.disabled:hover),
:deep(.form-btn.ant-btn-primary:disabled:hover),
:deep(.form-btn.ant-btn-primary.disabled:hover) {
  background: #d9d9d9 !important;
  background-color: #d9d9d9 !important;
  border-color: #d9d9d9 !important;
  color: rgba(0, 0, 0, 0.25) !important;
}
</style>
