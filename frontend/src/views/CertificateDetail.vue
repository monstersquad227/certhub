<template>
  <div class="page-container">
    <a-card title="证书详情" class="card-modern" v-if="certificate">
      <a-descriptions :column="2" bordered>
      <a-descriptions-item label="域名">{{ certificate.domain }}</a-descriptions-item>
      <a-descriptions-item label="证书类型">
        <a-tag :color="getCertTypeColor(certificate.cert_type)">
          {{ getCertTypeText(certificate.cert_type) }}
        </a-tag>
      </a-descriptions-item>
      <a-descriptions-item label="申请时间">
        {{ formatDate(certificate.created_at) }}
      </a-descriptions-item>
      <a-descriptions-item label="过期时间">
        {{ formatDate(certificate.expires_at) }}
      </a-descriptions-item>
      <a-descriptions-item label="证书状态">
        <a-tag :color="getStatusColor(certificate.status)">
          {{ getStatusText(certificate.status) }}
        </a-tag>
      </a-descriptions-item>
      <a-descriptions-item label="CA">
        <span v-if="!isCaError(certificate.ca)">{{ certificate.ca }}</span>
        <a-alert v-else type="error" :message="extractErrorMessage(certificate)" show-icon style="margin-top: 8px" />
      </a-descriptions-item>
      </a-descriptions>

      <a-divider class="divider-spaced" />

      <a-card v-if="certificate.status !== 'failed' && certificate.status !== 'pending'" title="证书内容 (Public Key)" size="small" class="card-modern content-section">
        <a-typography-paragraph>
          <pre class="cert-content">{{ certificate.public_key }}</pre>
        </a-typography-paragraph>
        <a-button type="link" @click="handleDownloadPublicKey">下载证书</a-button>
      </a-card>
      
      <a-card v-if="certificate.status === 'pending'" title="证书生成中" size="small" class="card-modern content-section">
      <a-alert
        message="证书正在生成中"
        description="您的证书正在处理中，请稍后刷新页面查看。通常需要几分钟时间。"
        type="info"
        show-icon
      />
      </a-card>

      <a-card v-if="certificate.status !== 'failed' && certificate.status !== 'pending'" title="私钥 (Private Key)" size="small" class="card-modern content-section">
        <a-alert
          message="安全提示"
          description="私钥是敏感信息，下载前需要验证身份。"
          type="warning"
          show-icon
          class="alert-spaced"
        />
      <a-button type="primary" @click="handleDownloadPrivateKey">下载私钥</a-button>
    </a-card>

    <a-modal
      v-model:open="verifyModalVisible"
      title="身份验证"
      @ok="handleVerifyAndDownload"
      :confirmLoading="verifying"
      okText="确定"
      cancelText="取消"
    >
      <a-form :model="verifyForm" layout="vertical">
        <a-form-item label="验证码" name="code">
          <a-input
            v-model:value="verifyForm.code"
            placeholder="请输入6位验证码"
            :maxlength="6"
            class="verify-input"
          />
          <div style="margin-top: 8px">
            <a-button type="link" size="small" @click="handleSendVerifyCode" :loading="sendingCode">
              发送验证码
            </a-button>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import api from '@/utils/api'
import { useAuthStore } from '@/store/auth'
import { getStatusColor, getStatusText, getCertTypeColor, getCertTypeText, extractErrorMessage, isCaError, formatDate } from '@/utils/certificate'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const certificate = ref<any>(null)
const verifyModalVisible = ref(false)
const verifying = ref(false)
const sendingCode = ref(false)
const verifyForm = reactive({
  code: '',
})


const fetchDetail = async () => {
  try {
    const id = route.params.id
    const res = await api.get(`/api/v1/certificates/${id}`)
    certificate.value = res.data.data.certificate
  } catch (error: any) {
    message.error(error.response?.data?.message || '获取证书详情失败')
    router.push('/certificates')
  }
}

const handleDownloadPublicKey = () => {
  if (!certificate.value) return
  const blob = new Blob([certificate.value.public_key], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${certificate.value.domain}.crt`
  a.click()
  URL.revokeObjectURL(url)
  message.success('证书下载成功')
}

const handleDownloadPrivateKey = () => {
  verifyModalVisible.value = true
}

const handleSendVerifyCode = async () => {
  const email = auth.user?.email
  if (!email) {
    message.error('用户信息获取失败，请重新登录')
    return
  }
  try {
    sendingCode.value = true
    await api.post('/api/v1/auth/send-code', { email })
    message.success('验证码已发送')
  } catch (error: any) {
    message.error(error.response?.data?.message || '发送验证码失败')
  } finally {
    sendingCode.value = false
  }
}

const handleVerifyAndDownload = async () => {
  if (!verifyForm.code || verifyForm.code.length !== 6) {
    message.error('请输入6位验证码')
    return
  }
  try {
    verifying.value = true
    const id = route.params.id
    const res = await api.post(`/api/v1/certificates/${id}/download-private-key`, {
      code: verifyForm.code,
    })
    const blob = new Blob([res.data.data.private_key], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${certificate.value?.domain || 'private'}.key`
    a.click()
    URL.revokeObjectURL(url)
    verifyModalVisible.value = false
    verifyForm.code = ''
    message.success('私钥下载成功')
  } catch (error: any) {
    message.error(error.response?.data?.message || '验证失败')
  } finally {
    verifying.value = false
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
.cert-content {
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 100px;
  overflow-y: auto;
  background: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.6;
  margin: 0;
  font-family: 'Courier New', monospace;
}

.cert-content::-webkit-scrollbar {
  width: 8px;
}

.cert-content::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.cert-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.cert-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.verify-input :deep(.ant-input) {
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  line-height: 40px;
}
</style>

