import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '@/utils/api';

export const useBalanceStore = defineStore('balance', () => {
  const balance = ref<number | null>(null);
  const loading = ref(false);

  async function fetchBalance() {
    loading.value = true;
    try {
      const res = await api.get('/api/v1/balance');
      balance.value = res.data.data.balance;
    } catch {
      // keep last known balance on transient errors
    } finally {
      loading.value = false;
    }
  }

  return {
    balance,
    loading,
    fetchBalance,
  };
});
