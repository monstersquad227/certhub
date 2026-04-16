<template>
  <div class="login-container">
    <div class="login-bg" aria-hidden="true" />
    <a-card class="login-card" :bordered="false">
      <h2 class="login-title">Certhub Platform</h2>
      <a-form :model="form" layout="vertical" :label-col="{ span: 24 }" class="login-form">
        <a-form-item label="邮箱">
          <a-input 
            v-model:value="form.email" 
            placeholder="请输入您的邮箱" 
            size="large"
            class="login-input"
          />
        </a-form-item>
        <a-form-item label="验证码">
          <div class="verify-code-group">
            <a-input
              v-model:value="form.code"
              placeholder="输入验证码"
              size="large"
              class="verify-code-input"
            />
            <a-button
              size="large"
              type="primary"
              :loading="sending"
              :disabled="countdown > 0"
              @click="onSendCode"
              class="send-code-btn"
            >
              {{ countdown > 0 ? `${countdown}秒` : '发送验证码' }}
            </a-button>
          </div>
        </a-form-item>
        <a-form-item class="login-btn-item">
          <a-button type="primary" size="large" block :loading="loggingIn" @click="onLogin" class="login-btn">
            登录
          </a-button>
          <p class="login-hint">未注册的邮箱将自动创建账号</p>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { message } from 'ant-design-vue';
import { useAuthStore } from '@/store/auth';

const auth = useAuthStore();
const router = useRouter();

const form = reactive({
  email: '',
  code: ''
});

const sending = ref(false);
const loggingIn = ref(false);
const countdown = ref(0);
let timer: number | null = null;

function startCountdown() {
  countdown.value = 60;
  timer && window.clearInterval(timer);
  timer = window.setInterval(() => {
    countdown.value -= 1;
    if (countdown.value <= 0 && timer) {
      window.clearInterval(timer);
      timer = null;
    }
  }, 1000);
}

async function onSendCode() {
  if (!form.email) {
    message.error('请输入邮箱');
    return;
  }
  try {
    sending.value = true;
    await auth.sendCode(form.email);
    message.success('验证码已发送，请检查邮箱');
    startCountdown();
  } catch (e: any) {
    message.error(e?.response?.data?.message || '发送失败');
  } finally {
    sending.value = false;
  }
}

async function onLogin() {
  if (!form.email || !form.code) {
    message.error('请输入邮箱和验证码');
    return;
  }
  try {
    loggingIn.value = true;
    await auth.login(form.email, form.code, false);
    message.success('登录成功');
    router.push('/certificates');
  } catch (e: any) {
    message.error(e?.response?.data?.message || '登录失败');
  } finally {
    loggingIn.value = false;
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.login-bg {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background: center / cover no-repeat;
  background-image: url('../assets/login-bg.jpg');
}

.login-bg::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.28);
}

.login-card {
  width: 100%;
  max-width: 520px;
  padding: 32px 44px;
  background: transparent;
  border: none;
  box-shadow: none;
  border-radius: 20px;
  position: relative;
  z-index: 1;
}

.login-card :deep(.ant-card-body) {
  padding: 0;
  background: transparent;
}

.login-title {
  text-align: center;
  margin: 0 0 26px 0;
  font-size: 36px;
  font-weight: 700;
  background: linear-gradient(135deg, #ffffff 0%, rgba(255, 255, 255, 0.82) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: 1px;
  position: relative;
}

.login-title::after {
  content: '';
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  width: 60px;
  height: 3px;
  background: linear-gradient(90deg, transparent, #3b82f6, transparent);
  border-radius: 2px;
}

.login-form :deep(.ant-form-item) {
  margin-bottom: 14px;
}

.login-form :deep(.ant-form-item-label > label) {
  font-weight: 500;
  color: rgba(255, 255, 255, 0.92);
  font-size: 14px;
  margin-bottom: 2px;
  display: block;
}

/* class 与 ant-input 在同一 input 上，需使用 .login-input.ant-input */
.login-input.ant-input,
.verify-code-input.ant-input {
  box-sizing: border-box;
  height: 44px;
  min-height: 44px;
  line-height: 44px;
  border-radius: 12px;
  border: 1.5px solid rgba(255, 255, 255, 0.28);
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
  font-size: 14px;
  padding: 0 16px;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    box-shadow 0.2s ease;
}

.verify-code-input.ant-input {
  width: 100%;
}

.login-input.ant-input::placeholder,
.verify-code-input.ant-input::placeholder {
  color: rgba(255, 255, 255, 0.45);
}

.login-input.ant-input:hover,
.verify-code-input.ant-input:hover {
  border-color: rgba(255, 255, 255, 0.4);
  background: rgba(255, 255, 255, 0.16);
}

.login-input.ant-input:focus,
.login-input.ant-input:focus-visible,
.verify-code-input.ant-input:focus,
.verify-code-input.ant-input:focus-visible {
  border-color: #3b82f6;
  background: rgba(255, 255, 255, 0.18);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.22);
  outline: none;
}

.verify-code-group {
  display: flex;
  gap: 12px;
  align-items: stretch;
}

.send-code-btn {
  height: 44px;
  min-width: 150px;
}

.send-code-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-btn-item {
  margin-top: 20px;
  margin-bottom: 0;
}

.login-hint {
  margin: 12px 0 0;
  text-align: center;
  font-size: 13px;
  line-height: 1.5;
  color: rgba(255, 255, 255, 0.65);
}

.login-btn {
  height: 44px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .login-card {
    padding: 26px 20px;
    margin: 20px;
    border-radius: 16px;
  }

  .login-title {
    font-size: 28px;
    margin-bottom: 22px;
  }

  .verify-code-group {
    flex-direction: column;
    gap: 12px;
  }

  .send-code-btn {
    width: 100%;
    min-width: auto;
  }

  .login-form :deep(.ant-form-item) {
    margin-bottom: 14px;
  }
}

/* 加载图标在深色背景上保持可读 */
.login-btn :deep(.ant-btn-loading-icon),
.send-code-btn :deep(.ant-btn-loading-icon) {
  color: rgba(255, 255, 255, 0.8);
}
</style>

