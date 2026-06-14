<script setup lang="ts">
import { ref, onMounted } from 'vue';

// ===== State =====
const activeTab = ref('update');
const theme = ref('system');
const statusMsg = ref('就绪');
const progress = ref(0);
const isRunning = ref(false);
const logs = ref<string[]>([]);

// Modals
const showDirModal = ref(false);
const dirOptions = ref<any[]>([]);
const selectedDir = ref('');
const pendingUpdateType = ref('');
const showCustomUrlModal = ref(false);
const customUrl = ref('');

// Icons (Inline SVG)
const icons = {
  update: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/></svg>`,
  logs: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`,
  faq: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`,
  check: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>`,
  cancel: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>`,
  squirrel: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/></svg>`,
  penguin: `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"/></svg>`
};

// ===== Actions =====
const applyTheme = (mode: string) => {
  theme.value = mode;
  const actual = mode === 'system' 
    ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light') 
    : mode;
  document.documentElement.setAttribute('data-theme', actual);
};

const executeUpdate = async (type: string, dir: string) => {
  showDirModal.value = false;
  isRunning.value = true;
  statusMsg.value = `正在更新 ${type}...`;
  progress.value = 10;
  
  let urlParam = '';
  if (type.startsWith('custom&url=')) {
    urlParam = decodeURIComponent(type.split('=')[1]);
    type = 'custom';
  }
  
  try {
    const res = await (window as any).go.main.App.UpdateAction(type, dir, urlParam);
    if (res.success) {
       statusMsg.value = '更新完成！请重新部署 Rime。';
       progress.value = 100;
    } else {
       statusMsg.value = '更新失败: ' + res.error;
    }
  } catch(e: any) {
    statusMsg.value = '更新失败，请查看日志或重试';
  } finally {
    setTimeout(() => {
      isRunning.value = false;
      progress.value = 0;
    }, 2000);
  }
};

const handleApiUpdate = async (type: string) => {
  if (isRunning.value) return;
  
  try {
    const info = await (window as any).go.main.App.GetSystemInfo();
    if (info.os === 'Darwin') {
      dirOptions.value = [
        { label: '鼠须管', value: '~/Library/Rime', icon: icons.squirrel },
        { label: '小企鹅', value: '~/.local/share/fcitx5/rime', icon: icons.penguin },
        { label: '自定义目录...', value: 'custom', icon: icons.check }
      ];
      selectedDir.value = dirOptions.value[0].value;
      pendingUpdateType.value = type;
      showDirModal.value = true;
    } else if (info.os === 'Linux') {
      dirOptions.value = [
        { label: 'iBus', value: '~/.config/ibus/rime', icon: icons.squirrel },
        { label: 'Fcitx5', value: '~/.local/share/fcitx5/rime', icon: icons.penguin },
        { label: 'Flatpak', value: '~/.var/app/org.fcitx.Fcitx5/data/fcitx5/rime', icon: icons.penguin },
        { label: '自定义目录...', value: 'custom', icon: icons.check }
      ];
      selectedDir.value = dirOptions.value[0].value;
      pendingUpdateType.value = type;
      showDirModal.value = true;
    } else {
      // Windows
      executeUpdate(type, '');
    }
  } catch (e) {
    statusMsg.value = '获取系统信息失败';
  }
};

const handleDirSelect = async () => {
  if (selectedDir.value === 'custom') {
     const dir = await (window as any).go.main.App.SelectDirectory();
     if (dir) {
       executeUpdate(pendingUpdateType.value, dir);
     }
  } else {
     executeUpdate(pendingUpdateType.value, selectedDir.value);
  }
};

const openCustomUrlModal = () => {
  customUrl.value = '';
  showCustomUrlModal.value = true;
};

const confirmCustomUrl = () => {
  if(!customUrl.value) return;
  showCustomUrlModal.value = false;
  handleApiUpdate(`custom&url=${encodeURIComponent(customUrl.value)}`);
};

const copyLogs = () => {
  navigator.clipboard.writeText(logs.value.join('\n')).then(() => {
    logs.value.push('【系统提示】控制台日志已复制到剪贴板。');
  });
};

const clearLogs = () => {
  logs.value = [];
};

const openUrl = (url: string) => {
  if ((window as any).go && (window as any).go.main && (window as any).go.main.App) {
     (window as any).go.main.App.OpenUrlBrowser(url);
  }
};

onMounted(() => {
  applyTheme('system');
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (theme.value === 'system') applyTheme('system');
  });

  // 绑定Wails事件机制接收日志和进度
  if ((window as any).runtime) {
    (window as any).runtime.EventsOn("log", (msg: string) => {
      logs.value.push(msg);
      setTimeout(() => {
        const container = document.querySelector('.logs-container');
        if (container) {
          container.scrollTop = container.scrollHeight;
        }
      }, 50);
    });

    (window as any).runtime.EventsOn("progress", (data: any) => {
      progress.value = data.percentage;
      if (data.percentage > 0 && data.percentage < 100) {
        statusMsg.value = `下载中... ${data.percentage.toFixed(1)}% (${data.details})`;
      }
    });
  }
});
</script>

<template>
  <div class="app-layout">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <h1>Oh My Rime</h1>
        <p>配置管理客户端</p>
      </div>

      <nav class="sidebar-nav">
        <button :class="['nav-btn', { active: activeTab === 'update' }]" @click="activeTab = 'update'">
          <span class="icon" v-html="icons.update"></span> 方案更新
        </button>
        <button :class="['nav-btn', { active: activeTab === 'logs' }]" @click="activeTab = 'logs'">
          <span class="icon" v-html="icons.logs"></span> 运行日志
        </button>
        <button :class="['nav-btn', { active: activeTab === 'faq' }]" @click="activeTab = 'faq'">
          <span class="icon" v-html="icons.faq"></span> 常见问题
        </button>
      </nav>

      <div class="sidebar-footer">
        <div class="theme-switcher">
          <label>主题模式</label>
          <div class="segment-control">
            <button :class="{ active: theme === 'light' }" @click="applyTheme('light')">明亮</button>
            <button :class="{ active: theme === 'dark' }" @click="applyTheme('dark')">暗黑</button>
            <button :class="{ active: theme === 'system' }" @click="applyTheme('system')">系统</button>
          </div>
        </div>
        <div class="version">v2.0.0</div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="main-content">
      <transition name="fade" mode="out-in">
        <!-- Tab: Update -->
        <div v-if="activeTab === 'update'" class="tab-view" key="update">
          <div class="cards-grid">
            <div class="card">
              <h2>薄荷方案 (Mint Scheme)</h2>
              <p>备份当前配置并替换为主方案。一键同步并覆盖最受欢迎的薄荷 Rime 基础输入法配置。</p>
              <button class="btn primary" @click="handleApiUpdate('main')" :disabled="isRunning">立即下载更新</button>
            </div>
            
            <div class="card">
              <h2>万象模型 (WanXiang Model)</h2>
              <p>搭载先进的万象中文语料模型 gram，大幅扩充和增强对多音字、错拼和长句联想的智能匹配率。</p>
              <button class="btn primary" @click="handleApiUpdate('model')" :disabled="isRunning">立即下载更新</button>
            </div>

            <div class="card">
              <h2>万象词库 (WanXiang Dict)</h2>
              <p>快速覆盖升级精简版（Lite）词库，为日常高频输入补充海量互联网热词与流行词。</p>
              <button class="btn primary" @click="handleApiUpdate('dict')" :disabled="isRunning">立即下载更新</button>
            </div>

            <div class="card">
              <h2>自定义更新 (Custom Link)</h2>
              <p>允许您手动粘贴特定的 zip 文件或模型 gram 文件直链，执行无缝的本地方案覆盖升级。</p>
              <button class="btn secondary" @click="openCustomUrlModal" :disabled="isRunning">自定义文件更新</button>
            </div>
          </div>

          <!-- Status Monitor -->
          <div class="card status-card">
            <h3>任务监控</h3>
            <div class="status-row">
              <span class="label">当前状态:</span>
              <span class="value">{{ statusMsg }}</span>
            </div>
            <div class="progress-track" v-if="isRunning || progress > 0">
              <div class="progress-fill" :style="{ width: progress + '%' }"></div>
            </div>
          </div>
        </div>

        <!-- Tab: Logs -->
        <div v-else-if="activeTab === 'logs'" class="tab-view logs-view" key="logs">
          <div class="logs-header">
            <h2>控制台日志记录</h2>
            <div class="actions">
              <button class="btn secondary" @click="copyLogs">复制日志</button>
              <button class="btn secondary" @click="clearLogs">清空日志</button>
            </div>
          </div>
          <div class="logs-container">
            <div class="log-line" v-for="(log, i) in logs" :key="i">{{ log }}</div>
            <div v-if="logs.length === 0" class="empty-logs">操作日志将在这里显示...</div>
          </div>
        </div>

        <!-- Tab: FAQ -->
        <div v-else-if="activeTab === 'faq'" class="tab-view faq-view" key="faq">
          <div class="card faq-intro">
            <h2>项目信息 & 文档</h2>
            <p>本客户端是薄荷输入法（Oh My Rime）的配套桌面更新工具。简化了手动克隆仓库、重命名备份和解压覆盖等复杂操作。</p>
            <div class="actions">
              <button class="btn secondary" @click="openUrl('https://space.bilibili.com/355567627')">关注作者 Bilibili</button>
              <button class="btn secondary" @click="openUrl('https://www.mintimate.cc')">打开薄荷文档</button>
            </div>
          </div>

          <h3 class="faq-title">常见问题与解答 (FAQ)</h3>
          <div class="accordion">
            <details class="accordion-item" open>
              <summary>1. 更新成功后，为什么我的输入法没有变化？</summary>
              <div class="content">Rime 输入法的所有配置更新在写入目录后，必须手动执行“重新部署”（Deploy / Sync）才能加载最新配置。请右键输入法托盘图标，选择“重新部署”。</div>
            </details>
            <details class="accordion-item">
              <summary>2. 更新会弄丢我自己写过的自定义配置吗？</summary>
              <div class="content">不会。本工具在执行更新时，会在目标 Rime 配置目录下把您当前的目录整体打包备份成一个带有时间戳的 ZIP 文件夹。</div>
            </details>
            <details class="accordion-item">
              <summary>3. 下载总是报错超时、或者速度极慢怎么办？</summary>
              <div class="content">您可以稍后重试，或者在“方案更新”中使用自定义链接功能，指定其他可用国内 CDN 镜像地址进行更新。</div>
            </details>
          </div>
        </div>
      </transition>
    </main>

    <!-- Dir Selection Modal -->
    <transition name="fade">
      <div class="modal-overlay" v-if="showDirModal">
        <div class="modal-card">
          <h3>选择配置目录</h3>
          <p class="modal-desc">请选择您使用的输入法框架:</p>
          <div class="dir-options">
            <button 
              v-for="opt in dirOptions" 
              :key="opt.value"
              :class="['btn secondary dir-btn', { active: selectedDir === opt.value }]"
              @click="selectedDir = opt.value"
            >
              <span class="icon" v-html="opt.icon"></span> {{ opt.label }}
            </button>
          </div>
          <div class="preview-path">
            目标路径: <code>{{ selectedDir }}</code>
          </div>
          <div class="modal-actions">
            <button class="btn secondary" @click="showDirModal = false">
               <span class="icon" v-html="icons.cancel"></span> 取消
            </button>
            <button class="btn primary" @click="handleDirSelect">
              <span class="icon" v-html="icons.check"></span> 确定
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- Custom URL Modal -->
    <transition name="fade">
      <div class="modal-overlay" v-if="showCustomUrlModal">
        <div class="modal-card">
          <h3>自定义更新</h3>
          <p class="modal-desc">
            • ZIP 文件 => 更新方案包<br>
            • GRAM 文件 => 更新模型文件
          </p>
          <div class="input-group">
            <label>请输入文件 URL:</label>
            <input type="text" v-model="customUrl" placeholder="https://..." />
          </div>
          <div class="modal-actions">
            <button class="btn secondary" @click="showCustomUrlModal = false">
               <span class="icon" v-html="icons.cancel"></span> 取消
            </button>
            <button class="btn primary" @click="confirmCustomUrl">
              <span class="icon" v-html="icons.check"></span> 确定
            </button>
          </div>
        </div>
      </div>
    </transition>

  </div>
</template>

<style scoped>
/* Same CSS as before */
.app-layout {
  display: flex;
  width: 100%;
  height: 100%;
}

.sidebar {
  width: 260px;
  background-color: var(--surface-color);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  padding: 24px;
  box-shadow: 2px 0 8px rgba(0,0,0,0.02);
  z-index: 10;
}

.sidebar-header h1 {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  text-align: center;
  margin-bottom: 4px;
}
.sidebar-header p {
  font-size: 13px;
  color: var(--text-secondary);
  text-align: center;
  font-style: italic;
  margin-bottom: 32px;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.nav-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: transparent;
  border: none;
  border-radius: var(--radius);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition);
}

.nav-btn .icon {
  width: 20px;
  height: 20px;
  opacity: 0.7;
}

.nav-btn:hover {
  background-color: var(--bg-color);
  transform: translateX(4px);
}

.nav-btn.active {
  background-color: var(--primary);
  color: white;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.nav-btn.active .icon {
  opacity: 1;
}

.sidebar-footer {
  margin-top: auto;
  border-top: 1px solid var(--border-color);
  padding-top: 24px;
}

.theme-switcher label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  display: block;
  margin-bottom: 8px;
}

.segment-control {
  display: flex;
  background: var(--bg-color);
  border-radius: var(--radius);
  padding: 4px;
}

.segment-control button {
  flex: 1;
  background: transparent;
  border: none;
  padding: 6px 0;
  font-size: 12px;
  color: var(--text-secondary);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.segment-control button.active {
  background: var(--surface-color);
  color: var(--text-primary);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.version {
  text-align: center;
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 16px;
  font-family: monospace;
}

.main-content {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
  position: relative;
}

.tab-view {
  display: flex;
  flex-direction: column;
  gap: 24px;
  height: 100%;
}

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.card {
  background: var(--surface-color);
  border-radius: var(--radius);
  padding: 24px;
  box-shadow: var(--card-shadow);
  border: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  transition: transform 0.3s, box-shadow 0.3s;
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px -4px rgb(0 0 0 / 0.1);
}

.card h2 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 12px;
}

.card h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}

.card p {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 24px;
  flex: 1;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: var(--radius);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all var(--transition);
}

.btn .icon {
  width: 18px;
  height: 18px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn.primary {
  background-color: var(--primary);
  color: white;
}

.btn.primary:not(:disabled):hover {
  background-color: var(--primary-hover);
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.btn.secondary {
  background-color: var(--surface-color);
  color: var(--text-primary);
  border-color: var(--border-color);
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

.btn.secondary:hover {
  background-color: var(--bg-color);
}

.status-card {
  margin-top: auto;
}

.status-row {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
}

.status-row .label {
  font-weight: 600;
  color: var(--text-secondary);
}

.status-row .value {
  font-weight: 500;
}

.progress-track {
  width: 100%;
  height: 8px;
  background-color: var(--bg-color);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: var(--primary);
  border-radius: 4px;
  transition: width 0.3s ease;
}

/* Logs */
.logs-view {
  display: flex;
  flex-direction: column;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logs-container {
  flex: 1;
  background: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  padding: 16px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 13px;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.05);
}

.empty-logs {
  color: var(--text-secondary);
  font-style: italic;
}

/* FAQ */
.faq-view .actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

.faq-title {
  margin-top: 8px;
  font-size: 18px;
  font-weight: 600;
}

.accordion {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.accordion-item {
  background: var(--surface-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  overflow: hidden;
}

.accordion-item summary {
  padding: 16px;
  font-weight: 600;
  cursor: pointer;
  background: var(--surface-color);
  transition: background var(--transition);
}

.accordion-item summary:hover {
  background: var(--bg-color);
}

.accordion-item .content {
  padding: 0 16px 16px;
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.6;
}

/* Modals */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--modal-overlay);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  animation: fadeInScale 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-card {
  background: var(--surface-color);
  border-radius: 16px;
  padding: 32px;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
  border: 1px solid var(--border-color);
}

.modal-card h3 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 12px;
  text-align: center;
}

.modal-desc {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 24px;
  line-height: 1.6;
}

.dir-options {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.dir-btn {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px;
  border-radius: var(--radius);
}

.dir-btn .icon {
  width: 32px;
  height: 32px;
}

.dir-btn.active {
  background-color: var(--primary);
  color: white;
  border-color: var(--primary);
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.preview-path {
  background: var(--bg-color);
  padding: 12px;
  border-radius: var(--radius);
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 32px;
  text-align: center;
  border: 1px dashed var(--border-color);
}

.preview-path code {
  color: var(--primary);
  font-weight: 600;
  display: block;
  margin-top: 4px;
}

.input-group {
  margin-bottom: 32px;
}

.input-group label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
}

.input-group input {
  width: 100%;
  padding: 12px 16px;
  border-radius: var(--radius);
  border: 1px solid var(--border-color);
  background: var(--bg-color);
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.input-group input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.modal-actions {
  display: flex;
  gap: 16px;
  justify-content: flex-end;
}

.modal-actions .btn {
  flex: 1;
}
</style>
