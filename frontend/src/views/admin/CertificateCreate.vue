<template>
  <div class="page-container-narrow">
    <a-card title="新增证书" class="card-modern">
      <a-form :model="form" :rules="rules" ref="formRef" layout="vertical" class="form-container-wide" :label-col="{ span: 24 }">
      <a-form-item label="用户邮箱" name="user_email">
        <a-input v-model:value="form.user_email" placeholder="请输入用户邮箱" class="form-input" />
      </a-form-item>

      <a-form-item label="域名" name="domain">
        <a-input v-model:value="form.domain" placeholder="请输入域名" class="form-input" />
      </a-form-item>

      <a-form-item label="CA" name="ca">
        <a-input v-model:value="form.ca" placeholder="请输入证书颁发机构" class="form-input" />
      </a-form-item>

      <a-form-item label="Private Key" name="private_key">
        <a-textarea
          v-model:value="form.private_key"
          :rows="6"
          placeholder="请输入私钥内容"
        />
      </a-form-item>

      <a-form-item label="Public Key" name="public_key">
        <a-textarea
          v-model:value="form.public_key"
          :rows="6"
          placeholder="请输入证书内容"
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

      <a-form-item label="申请时间" name="created_at">
        <a-date-picker
          v-model:value="form.created_at"
          show-time
          format="YYYY-MM-DD HH:mm:ss"
          style="width: 100%"
          class="form-date-picker"
        />
        <div class="text-hint">可选，默认为当前时间</div>
      </a-form-item>

      <a-form-item>
        <div class="button-group">
          <a-button type="primary" :loading="submitting" @click="handleSubmit" class="form-btn">
            创建
          </a-button>
          <a-button @click="$router.push('/admin/certificates')" class="form-btn">取消</a-button>
        </div>
      </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import api from '@/utils/api'
import dayjs, { type Dayjs } from 'dayjs'

const router = useRouter()
const formRef = ref()
const submitting = ref(false)

const form = reactive({
  user_email: '',
  domain: '',
  ca: '',
  private_key: '',
  public_key: '',
  expires_at: null as Dayjs | null,
  created_at: null as Dayjs | null,
})

const rules = {
  user_email: [{ required: true, message: '请输入用户邮箱', trigger: 'blur' }],
  domain: [{ required: true, message: '请输入域名', trigger: 'blur' }],
  ca: [{ required: true, message: '请输入CA', trigger: 'blur' }],
  private_key: [{ required: true, message: '请输入私钥', trigger: 'blur' }],
  public_key: [{ required: true, message: '请输入证书内容', trigger: 'blur' }],
  expires_at: [{ required: true, message: '请选择过期时间', trigger: 'change' }],
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    if (!form.expires_at) {
      message.error('请选择过期时间')
      return
    }
    submitting.value = true
    const payload: any = {
      user_email: form.user_email,
      domain: form.domain,
      ca: form.ca,
      private_key: form.private_key,
      public_key: form.public_key,
      expires_at: form.expires_at.toISOString(),
    }
    if (form.created_at) {
      payload.created_at = form.created_at.toISOString()
    }
    await api.post('/api/v1/admin/certificates', payload)
    message.success('证书创建成功')
    router.push('/admin/certificates')
  } catch (error: any) {
    message.error(error.response?.data?.message || '创建证书失败')
  } finally {
    submitting.value = false
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

