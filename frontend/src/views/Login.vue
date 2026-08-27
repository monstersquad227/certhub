<template>
  <div class="login-container">
    <section class="login-brand-panel" aria-hidden="true" />

    <section class="login-form-panel">
      <div class="login-form-wrap">
        <h2 class="login-title">邮箱登录</h2>

        <a-form :model="form" layout="vertical" class="login-form">
          <a-form-item label="邮箱">
            <a-input
              v-model:value="form.email"
              placeholder="you@company.com"
              size="large"
              class="login-input"
            />
          </a-form-item>
          <a-form-item label="验证码">
            <div class="verify-code-group">
              <a-input
                v-model:value="form.code"
                placeholder="6 位数字验证码"
                size="large"
                class="verify-code-input"
                :maxlength="6"
              />
              <a-button
                size="large"
                :loading="sending"
                :disabled="countdown > 0"
                @click="onSendCode"
                class="send-code-btn"
              >
                {{ countdown > 0 ? `${countdown}秒` : '获取验证码' }}
              </a-button>
            </div>
          </a-form-item>
          <a-form-item class="login-btn-item">
            <a-button type="primary" size="large" block :loading="loggingIn" @click="onLogin" class="login-btn">
              登录
            </a-button>
          </a-form-item>
        </a-form>

        <p class="login-terms">
          登录即代表同意
          <a href="#" class="terms-link" @click.prevent>服务条款</a>
          与
          <a href="#" class="terms-link" @click.prevent>隐私政策</a>
        </p>
      </div>
    </section>
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
  width: 100%;
  height: 100vh;
  min-height: 100vh;
  display: flex;
  margin: 0;
  padding: 0;
  overflow: hidden;
}

/* 左侧品牌区 60% */
.login-brand-panel {
  flex: 0 0 60%;
  width: 60%;
  height: 100%;
  min-height: 100vh;
  background: center / cover no-repeat url('../assets/login-bg.jpg');
}

/* 右侧功能区 40% */
.login-form-panel {
  flex: 0 0 40%;
  width: 40%;
  height: 100%;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 40px;
  background: #ffffff;
}

.login-form-wrap {
  width: 100%;
  max-width: 380px;
}

.login-title {
  margin: 0 0 32px;
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  line-height: 1.3;
}

.login-form :deep(.ant-form-item) {
  margin-bottom: 20px;
}

.login-form :deep(.ant-form-item-label > label) {
  font-weight: 500;
  color: #374151;
  font-size: 14px;
  height: auto;
}

.login-input.ant-input,
.verify-code-input.ant-input {
  box-sizing: border-box;
  height: 44px;
  min-height: 44px;
  border-radius: 8px;
  border: 1px solid #d1d5db;
  background: #ffffff;
  color: #111827;
  font-size: 14px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.login-input.ant-input::placeholder,
.verify-code-input.ant-input::placeholder {
  color: #9ca3af;
}

.login-input.ant-input:hover,
.verify-code-input.ant-input:hover {
  border-color: #9ca3af;
}

.login-input.ant-input:focus,
.login-input.ant-input:focus-visible,
.verify-code-input.ant-input:focus,
.verify-code-input.ant-input:focus-visible {
  border-color: #2563eb;
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.15);
  outline: none;
}

.verify-code-group {
  display: flex;
  gap: 12px;
  align-items: stretch;
}

.verify-code-input.ant-input {
  flex: 1;
  min-width: 0;
}

.send-code-btn {
  height: 44px;
  min-width: 120px;
  flex-shrink: 0;
  border-radius: 8px;
  border: 1px solid #d1d5db;
  background: #ffffff;
  color: #2563eb;
  font-weight: 500;
  box-shadow: none;
}

.send-code-btn:hover:not(:disabled) {
  color: #1d4ed8;
  border-color: #2563eb;
  background: #eff6ff;
}

.send-code-btn:disabled {
  color: #9ca3af;
  border-color: #e5e7eb;
  background: #f9fafb;
}

.login-btn-item {
  margin-top: 8px;
  margin-bottom: 0;
}

.login-btn {
  height: 44px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 15px;
  background: #2563eb;
  border-color: #2563eb;
}

.login-btn:hover {
  background: #1d4ed8;
  border-color: #1d4ed8;
}

.login-terms {
  margin: 24px 0 0;
  text-align: center;
  font-size: 13px;
  color: #9ca3af;
  line-height: 1.6;
}

.terms-link {
  color: #2563eb;
  text-decoration: none;
}

.terms-link:hover {
  text-decoration: underline;
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    height: 100vh;
  }

  .login-brand-panel {
    flex: 6;
    width: 100%;
    height: auto;
    min-height: 0;
    background-position: center;
    background-size: cover;
  }

  .login-form-panel {
    flex: 4;
    width: 100%;
    height: auto;
    min-height: 0;
    padding: 32px 24px;
  }

  .verify-code-group {
    flex-direction: column;
  }

  .send-code-btn {
    width: 100%;
    min-width: auto;
  }
}
</style>
