<template>
  <div class="page-container-narrow">
    <a-card title="DNS 记录生成" class="card-modern">
      <a-form :model="form" :rules="rules" ref="formRef" layout="vertical" :label-col="{ span: 24 }">
        <a-form-item label="域名" name="domain">
          <a-input
            v-model:value="form.domain"
            placeholder="请输入域名，如 example.com 或 *.example.com"
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
          description="请前往您的域名服务商（如阿里云、腾讯云等）添加以下 DNS 解析记录。"
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

        <a-alert
          message="配置步骤"
          type="success"
          class="alert-spaced"
        >
          <template #description>
            <ol style="margin: var(--spacing-sm) 0; padding-left: 20px">
              <li>登录您的域名管理平台（如阿里云、腾讯云、Cloudflare 等）</li>
              <li>找到 DNS 解析设置页面</li>
              <li>添加一条新的 TXT 记录</li>
              <li>记录名称填写：<code>{{ dnsRecord.name }}</code></li>
              <li>记录值填写：<code>{{ dnsRecord.value }}</code></li>
              <li>TTL 使用默认值即可</li>
              <li>保存后等待 DNS 解析生效（通常几分钟内）</li>
            </ol>
          </template>
        </a-alert>

        <div class="button-group" style="margin-top: var(--spacing-lg)">
          <a-button
            type="primary"
            @click="handleGenerateAgain"
            class="form-btn"
          >
            重新生成
          </a-button>
          <a-button @click="handleGoToApply" class="form-btn">
            申请证书
          </a-button>
        </div>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import api from '@/utils/api'

const router = useRouter()
const formRef = ref()
const generatingDNS = ref(false)

const form = reactive({
  domain: '',
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
    message.success('DNS 记录生成成功')
  } catch (error: any) {
    message.error(error.response?.data?.message || '生成 DNS 记录失败')
  } finally {
    generatingDNS.value = false
  }
}

const handleGenerateAgain = () => {
  form.domain = ''
  dnsRecord.value = null
  formRef.value?.resetFields()
}

const handleGoToApply = () => {
  router.push('/certificates/apply')
}
</script>

<style scoped>
code {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

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

