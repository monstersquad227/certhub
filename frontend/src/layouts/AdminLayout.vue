<template>
  <a-layout class="app-layout">
    <a-layout-header class="layout-header">
      <div class="layout-logo">CertHub 管理端</div>
      <div class="layout-nav">
        <a-menu mode="horizontal" theme="dark" :selected-keys="[activeKey]">
          <a-menu-item key="admin-cert" @click="go('/admin/certificates')">证书管理</a-menu-item>
        </a-menu>
      </div>
      <div class="layout-user">
        <span>{{ auth.user?.email }}</span>
        <a-button type="link" @click="onLogout" style="color: rgba(255, 255, 255, 0.85); padding: 0">退出</a-button>
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
  if (route.path.startsWith('/admin/certificates')) return 'admin-cert';
  return 'admin-cert';
});

function go(path: string) {
  router.push(path);
}

function onLogout() {
  auth.logout();
  router.push('/admin/login');
}
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
}
</style>

