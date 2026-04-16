<template>
  <a-layout class="app-layout">
    <a-layout-header class="layout-header">
      <div class="layout-logo">CertHub</div>
      <div class="layout-nav">
        <a-menu mode="horizontal" theme="dark" :selected-keys="[activeKey]">
          <a-menu-item key="certificates" @click="go('/certificates')">证书列表</a-menu-item>
          <a-menu-item key="balance" @click="go('/balance')">余额中心</a-menu-item>
        </a-menu>
      </div>
      <div class="layout-user">
        <span class="user-email">{{ auth.user?.email }}</span>
        <a-button type="link" @click="onLogout" class="logout-btn">退出</a-button>
      </div>
    </a-layout-header>
    <a-layout-content class="layout-content">
      <router-view />
    </a-layout-content>
  </a-layout>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/store/auth';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const activeKey = computed(() => {
  if (route.path.startsWith('/balance')) return 'balance';
  return 'certificates';
});

function go(path: string) {
  router.push(path);
}

function onLogout() {
  auth.logout();
  router.push('/login');
}
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  background: #f5f7fa;
}

.layout-header {
  background: linear-gradient(135deg, #1a1f3a 0%, #0a0e27 100%);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding: 0 40px;
  height: 80px;
  line-height: 80px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.layout-logo {
  font-weight: 900;
  font-size: 32px;
  letter-spacing: 2px;
  background: linear-gradient(135deg, #ffffff 0%, rgba(255, 255, 255, 0.95) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  line-height: 80px;
}

.layout-nav {
  flex: 1;
  margin-left: 40px;
}

.layout-nav :deep(.ant-menu) {
  height: 80px;
  line-height: 80px;
  background: transparent;
  border: none;
}

.layout-nav :deep(.ant-menu-item) {
  font-size: 18px;
  font-weight: 600;
  padding: 0 32px;
  height: 80px;
  line-height: 80px;
  margin: 0 4px;
  border-radius: 8px;
  transition: all 0.3s;
}

.layout-nav :deep(.ant-menu-item:hover) {
  background: rgba(255, 255, 255, 0.1);
}

.layout-nav :deep(.ant-menu-item-selected) {
  background: rgba(255, 255, 255, 0.15);
}

.layout-user {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-email {
  color: rgba(255, 255, 255, 0.95);
  font-size: 17px;
  font-weight: 500;
}

.logout-btn {
  color: rgba(255, 255, 255, 0.9);
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 17px;
  font-weight: 500;
  transition: all 0.2s;
  height: auto;
}

.logout-btn:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.2);
}

.layout-content {
  padding: var(--spacing-xl);
  background: #f5f7fa;
  min-height: calc(100vh - 80px);
}
</style>

