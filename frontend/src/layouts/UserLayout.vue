<template>
  <a-layout class="app-layout">
    <a-layout-sider class="layout-sider" :width="240">
      <div class="sider-inner">
        <div class="sider-top">
          <div class="layout-logo">CertHub</div>
          <a-menu
            mode="inline"
            theme="dark"
            :selected-keys="[activeKey]"
            class="sider-menu"
          >
            <a-menu-item key="certificates" @click="go('/certificates')">
              <template #icon><FileProtectOutlined /></template>
              证书列表
            </a-menu-item>
            <a-menu-item key="balance" @click="go('/balance')">
              <template #icon><WalletOutlined /></template>
              余额中心
            </a-menu-item>
          </a-menu>
        </div>

        <div class="sider-bottom">
          <div class="user-section">
            <span class="user-email">{{ auth.user?.email }}</span>
            <a-button type="link" @click="onLogout" class="logout-btn">退出</a-button>
          </div>
        </div>
      </div>
    </a-layout-sider>

    <a-layout class="main-layout">
      <a-layout-header class="content-header" />
      <a-layout-content class="layout-content layout-content-fill">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { FileProtectOutlined, WalletOutlined } from '@ant-design/icons-vue';
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
}

.layout-sider {
  background: linear-gradient(180deg, #1a1f3a 0%, #0a0e27 100%) !important;
  box-shadow: 4px 0 20px rgba(0, 0, 0, 0.15);
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 100;
}

.layout-sider :deep(.ant-layout-sider-children) {
  display: flex;
  flex-direction: column;
}

.sider-inner {
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 24px 0 20px;
}

.sider-top {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.layout-logo {
  font-weight: 900;
  font-size: 24px;
  letter-spacing: 1px;
  color: #fff;
  padding: 0 24px 24px;
  line-height: 1.2;
}

.sider-menu {
  background: transparent;
  border-inline-end: none !important;
}

.sider-menu :deep(.ant-menu-item) {
  margin: 4px 12px;
  width: calc(100% - 24px);
  border-radius: 8px;
  font-size: 15px;
  font-weight: 500;
}

.sider-menu :deep(.ant-menu-item-selected) {
  background: rgba(255, 255, 255, 0.15) !important;
}

.sider-bottom {
  flex-shrink: 0;
  padding: 16px 16px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.user-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 0 8px;
}

.user-email {
  color: rgba(255, 255, 255, 0.75);
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.logout-btn {
  color: rgba(255, 255, 255, 0.65);
  padding: 0;
  height: auto;
  font-size: 13px;
  flex-shrink: 0;
}

.logout-btn:hover {
  color: #fff;
}

.main-layout {
  margin-left: 240px;
  width: calc(100% - 240px);
  min-height: 100vh;
  background: #f5f7fa;
  display: flex;
  flex-direction: column;
}

.content-header {
  height: 56px;
  line-height: 56px;
  padding: 0;
  background: #fff;
  border-bottom: 1px solid #e8ecef;
  flex-shrink: 0;
}

.layout-content {
  background: #f5f7fa;
  flex: 1;
  min-height: 0;
}
</style>
