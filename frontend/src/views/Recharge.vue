<template>
  <div class="page-container">
    <a-card title="账户充值" class="card-modern recharge-card">
      <a-form :model="form" :rules="rules" ref="formRef" layout="vertical" class="form-container" :label-col="{ span: 24 }">
        <a-form-item label="充值金额" name="amount">
          <a-input-number
            v-model:value="form.amount"
            :min="1"
            :precision="2"
            style="width: 100%"
            placeholder="请输入充值金额"
            class="amount-input"
          />
          <div class="amount-buttons">
            <a-button @click="setAmount(10)" class="amount-btn">10元</a-button>
            <a-button @click="setAmount(50)" class="amount-btn">50元</a-button>
            <a-button @click="setAmount(100)" class="amount-btn">100元</a-button>
            <a-button @click="setAmount(200)" class="amount-btn">200元</a-button>
            <a-button @click="setAmount(500)" class="amount-btn">500元</a-button>
          </div>
        </a-form-item>

        <a-form-item label="支付方式" name="payment_method">
          <a-radio-group v-model:value="form.payment_method" class="payment-group">
            <a-radio value="alipay"> 支付宝 </a-radio>
            <a-radio value="wechat"> 微信支付 </a-radio>
          </a-radio-group>
        </a-form-item>

        <a-form-item>
          <div class="button-group">
            <a-button type="primary" :loading="submitting" @click="handleSubmit" class="submit-btn">
              立即充值
            </a-button>
            <a-button @click="$router.push('/balance')" class="back-btn">返回</a-button>
          </div>
        </a-form-item>
      </a-form>

      <a-alert
        message="提示"
        description="当前版本为模拟支付流程，充值将直接成功。后续版本将接入真实支付渠道。"
        type="info"
        show-icon
        class="alert-spaced"
      />
    </a-card>
  </div>
</template>

<style scoped>
.recharge-card {
  max-width: 600px;
  margin: 0 auto;
}

.form-container {
  max-width: 100%;
}

.amount-input :deep(.ant-input-number-input) {
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  line-height: 40px;
}

.amount-input :deep(.ant-input-number) {
  width: 100%;
}

.amount-buttons {
  margin-top: 12px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.amount-btn {
  border-radius: 12px;
  height: 36px;
  padding: 0 16px;
  transition: all 0.3s;
}

.amount-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.payment-group {
  padding: 8px 0;
}

.payment-group :deep(.ant-radio-wrapper) {
  font-size: 16px;
  margin-right: 24px;
}

.button-group {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.submit-btn,
.back-btn {
  border-radius: 12px;
  height: 40px;
  padding: 0 32px;
  font-weight: 500;
  display: flex;
  align-items: center;
}

.alert-spaced {
  margin-top: 24px;
}
</style>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import api from '@/utils/api'

const router = useRouter()
const formRef = ref()
const submitting = ref(false)

const form = reactive({
  amount: undefined as number | undefined,
  payment_method: 'alipay',
})

const rules = {
  amount: [{ required: true, message: '请输入充值金额', trigger: 'blur' }],
  payment_method: [{ required: true, message: '请选择支付方式', trigger: 'change' }],
}

const setAmount = (amount: number) => {
  form.amount = amount
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    const res = await api.post('/api/v1/balance/recharge', form)
    message.success(`充值订单创建成功，订单号：${res.data.order_no}`)
    // 模拟支付成功，直接更新余额
    setTimeout(() => {
      message.success('充值成功！')
      router.push('/balance')
    }, 1000)
  } catch (error: any) {
    message.error(error.response?.data?.message || '创建充值订单失败')
  } finally {
    submitting.value = false
  }
}
</script>

