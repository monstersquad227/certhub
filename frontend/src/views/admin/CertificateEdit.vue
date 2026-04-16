<template>
  <div class="page-container-narrow">
    <a-card title="编辑证书" class="card-modern">
      <a-form :model="form" :rules="rules" ref="formRef" layout="vertical" class="form-container-wide" :label-col="{ span: 24 }">
      <a-form-item label="用户邮箱" name="user_email">
        <a-input v-model:value="form.user_email" placeholder="请输入用户邮箱" class="form-input" />
      </a-form-item>

      <a-form-item label="CA" name="ca">
        <a-input v-model:value="form.ca" placeholder="请输入证书颁发机构" class="form-input" />
      </a-form-item>

      <a-form-item label="Private Key" name="private_key">
        <a-textarea
          v-model:value="form.private_key"
          :rows="6"
          placeholder="请输入私钥内容（留空则不更新）"
        />
      </a-form-item>

      <a-form-item label="Public Key" name="public_key">
        <a-textarea
          v-model:value="form.public_key"
          :rows="6"
          placeholder="请输入证书内容（留空则不更新）"
        />
      </a-form-item>

      <a-form-item label="过期时间" name="expires_at">
        <a-date-picker
          v-model:value="form.expires_at"
          show-time
          format="YYYY-MM-DD HH:mm:ss"
          style="width: 100%"
          class="form-date-picker"
        />
      </a-form-item>

      <a-form-item>
        <div class="button-group">
          <a-button type="primary" :loading="submitting" @click="handleSubmit" class="form-btn">
            更新
          </a-button>
          <a-button @click="$router.push('/admin/certificates')" class="form-btn">取消</a-button>
        </div>
      </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import api from '@/utils/api'
import dayjs, { type Dayjs } from 'dayjs'

const route = useRoute()
const router = useRouter()
const formRef = ref()
const submitting = ref(false)
const loading = ref(false)

const form = reactive({
  user_email: '',
  ca: '',
  private_key: '',
  public_key: '',
  expires_at: null as Dayjs | null,
})

const rules = {
  user_email: [{ required: false }],
  ca: [{ required: false }],
  private_key: [{ required: false }],
  public_key: [{ required: false }],
  expires_at: [{ required: false }],
}

const fetchDetail = async () => {
  try {
    loading.value = true
    const id = route.params.id
    const res = await api.get(`/api/v1/certificates/${id}`)
    const cert = res.data.data.certificate
    // 注意：这里需要获取用户邮箱，但API可能不返回，需要额外查询或从列表页传入
    form.ca = cert.ca
    form.public_key = cert.public_key
    form.expires_at = dayjs(cert.expires_at)
    // private_key 需要解密，但编辑时通常不显示，留空让用户重新输入
  } catch (error: any) {
    message.error(error.response?.data?.message || '获取证书详情失败')
    router.push('/admin/certificates')
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  try {
    submitting.value = true
    const id = route.params.id
    const payload: any = {}
    if (form.user_email) {
      payload.user_email = form.user_email
    }
    if (form.ca) {
      payload.ca = form.ca
    }
    if (form.private_key) {
      payload.private_key = form.private_key
    }
    if (form.public_key) {
      payload.public_key = form.public_key
    }
    if (form.expires_at) {
      payload.expires_at = form.expires_at.toISOString()
    }
    await api.put(`/api/v1/admin/certificates/${id}`, payload)
    message.success('证书更新成功')
    router.push('/admin/certificates')
  } catch (error: any) {
    message.error(error.response?.data?.message || '更新证书失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
.form-input :deep(.ant-input) {
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  line-height: 40px;
}

.form-date-picker :deep(.ant-picker) {
  height: 40px;
  border-radius: 12px;
}

.form-btn {
  height: 40px;
  padding: 0 24px;
  border-radius: 12px;
  display: flex;
  align-items: center;
}
</style>

