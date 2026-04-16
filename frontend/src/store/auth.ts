import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import api from '@/utils/api';

export interface UserInfo {
  id: number;
  email: string;
  role: 'user' | 'admin';
  balance: number;
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('TOKEN') || null);
  
  // 从 localStorage 恢复用户信息
  const savedUser = localStorage.getItem('USER');
  const user = ref<UserInfo | null>(savedUser ? JSON.parse(savedUser) : null);

  const isLoggedIn = computed(() => !!token.value);

  function setAuth(newToken: string, newUser: UserInfo) {
    token.value = newToken;
    user.value = newUser;
    localStorage.setItem('TOKEN', newToken);
    localStorage.setItem('USER', JSON.stringify(newUser));
  }

  function logout() {
    token.value = null;
    user.value = null;
    localStorage.removeItem('TOKEN');
    localStorage.removeItem('USER');
  }

  async function login(email: string, code: string, isAdmin = false) {
    const url = isAdmin ? '/api/v1/admin/auth/login' : '/api/v1/auth/login';
    const res = await api.post(url, { email, code });
    setAuth(res.data.data.token, res.data.data.user);
  }

  async function sendCode(email: string) {
    await api.post('/api/v1/auth/send-code', { email });
  }

  return {
    token,
    user,
    isLoggedIn,
    setAuth,
    logout,
    login,
    sendCode
  };
});


