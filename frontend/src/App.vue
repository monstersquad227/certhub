<template>
  <router-view />
</template>

<script setup lang="ts">
</script>

<style>
html,
body,
#app {
  height: 100%;
  margin: 0;
  min-height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

* {
  box-sizing: border-box;
}

/* Ant Design Vue component overrides for better visual consistency */
.ant-card {
  border-radius: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border: 1px solid #f0f0f0;
}

.ant-card-head {
  border-bottom: 1px solid #f0f0f0;
  font-weight: 500;
  padding: 16px 24px;
}

.ant-card-body {
  padding: 24px;
}

.ant-table-wrapper .ant-table {
  border-radius: 12px;
}

.ant-btn {
  border-radius: 12px;
  font-weight: 500;
  height: 40px;
  padding: 0 20px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.ant-btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: none;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.25);
}

.ant-btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.35);
  transform: translateY(-1px);
}

.ant-btn-lg {
  height: 48px;
  padding: 0 24px;
  font-size: 16px;
}

.ant-btn-sm {
  height: 32px;
  padding: 0 16px;
  font-size: 14px;
  border-radius: 8px;
}

.ant-input,
.ant-select-selector,
.ant-picker {
  border-radius: 12px;
  border-color: #d9d9d9;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.ant-input:hover,
.ant-select-selector:hover,
.ant-picker:hover {
  border-color: #40a9ff;
}

.ant-input:focus,
.ant-input-focused,
.ant-select-focused .ant-select-selector,
.ant-picker-focused {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

/* Smooth transitions for better UX */
.ant-card,
.ant-btn,
.ant-input,
.ant-select-selector {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Focus states */
.ant-input:focus,
.ant-input-focused,
.ant-select-focused .ant-select-selector {
  box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}
</style>

