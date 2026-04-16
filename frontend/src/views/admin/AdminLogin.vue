<template>
  <div class="login-container">
    <a-card title="管理员登录" class="login-card" :bordered="false">
      <a-form :model="form" :rules="rules" ref="formRef" layout="vertical" :label-col="{ span: 24 }">
        <a-form-item label="管理员邮箱" name="email">
          <a-input
            v-model:value="form.email"
            placeholder="请输入管理员邮箱"
            size="large"
            @pressEnter="handleSendCode"
          />
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            :loading="sendingCode"
            @click="handleSendCode"
            block
            size="large"
          >
            获取验证码
          </a-button>
        </a-form-item>

        <a-form-item label="验证码" name="code">
          <a-input
            v-model:value="form.code"
            placeholder="请输入6位验证码"
            size="large"
            :maxlength="6"
            @pressEnter="handleLogin"
          />
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            :loading="logging"
            @click="handleLogin"
            block
            size="large"
          >
            登录
          </a-button>
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
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()
const sendingCode = ref(false)
const logging = ref(false)

const form = reactive({
  email: '',
  code: '',
})

const rules = {
  email: [
    { required: true, message: '请输入管理员邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' },
  ],
}

const handleSendCode = async () => {
  try {
    await formRef.value.validateFields(['email'])
    sendingCode.value = true
    await api.post('/api/v1/auth/send-code', { email: form.email })
    message.success('验证码已发送')
  } catch (error: any) {
    message.error(error.response?.data?.message || '发送验证码失败')
  } finally {
    sendingCode.value = false
  }
}

const handleLogin = async () => {
  try {
    await formRef.value.validate()
    logging.value = true
    const res = await api.post('/api/v1/admin/auth/login', form)
    authStore.setAuth(res.data.data.token, res.data.data.user)
    message.success('登录成功')
    router.push('/admin/certificates')
  } catch (error: any) {
    message.error(error.response?.data?.message || '登录失败')
  } finally {
    logging.value = false
  }
}
</script>

<style scoped>
/* Styles are now in common.css */
</style>

