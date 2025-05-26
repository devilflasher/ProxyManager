import './app.css';

import {
    GetAllProxies,
    AddProxy,
    UpdateProxy,
    DeleteProxy,
    StartProxy,
    StopProxy,
    StartAllProxies,
    StopAllProxies,
    GetStats,
    ExportConfigToFile,
    ImportConfigFromFile
} from '../wailsjs/go/main/App'

import { BrowserOpenURL } from '../wailsjs/runtime/runtime'

class ProxyManager {
    constructor() {
        this.proxies = [];
        this.selectedProxies = new Set();
        this.currentEditingProxy = null; // 添加当前编辑的代理
        this.initializeElements();
        this.bindEvents();
        this.loadProxies();
        setInterval(() => this.loadProxies(), 15000); // 改为15秒刷新一次
    }

    initializeElements() {
        this.addBtn = document.getElementById('addBtn');
        this.exportBtn = document.getElementById('exportBtn');
        this.importBtn = document.getElementById('importBtn');
        this.selectAllBtn = document.getElementById('selectAllBtn');
        this.startSelectedBtn = document.getElementById('startSelectedBtn');
        this.stopSelectedBtn = document.getElementById('stopSelectedBtn');
        this.deleteSelectedBtn = document.getElementById('deleteSelectedBtn');
        this.selectedCount = document.getElementById('selectedCount');
        this.totalCount = document.getElementById('totalCount');
        this.runningCount = document.getElementById('runningCount');
        this.proxyList = document.getElementById('proxyList');
        this.modal = document.getElementById('proxyModal');
        this.modalContent = document.querySelector('.modal-content');
        this.modalTitle = document.getElementById('modalTitle');
        this.proxyForm = document.getElementById('proxyForm');
        this.closeBtn = document.querySelector('.close');
        this.cancelBtn = document.getElementById('cancelBtn');
    }

    bindEvents() {
        this.addBtn.addEventListener('click', () => this.showAddModal());
        this.exportBtn.addEventListener('click', () => this.exportConfig());
        this.importBtn.addEventListener('click', () => this.importConfig());
        this.selectAllBtn.addEventListener('change', () => this.toggleSelectAll());
        this.startSelectedBtn.addEventListener('click', () => this.startSelectedProxies());
        this.stopSelectedBtn.addEventListener('click', () => this.stopSelectedProxies());
        this.deleteSelectedBtn.addEventListener('click', () => this.deleteSelectedProxies());
        this.closeBtn.addEventListener('click', () => this.hideModal());
        this.cancelBtn.addEventListener('click', () => this.hideModal());
        
        if (this.modalContent) {
            this.modalContent.addEventListener('click', (e) => {
                e.stopPropagation();
            });
        }
        
        this.proxyForm.addEventListener('submit', (e) => this.handleFormSubmit(e));
        
        document.addEventListener('keydown', (e) => this.handleKeyDown(e));
        
        this.bindExternalLinks();
    }

    bindExternalLinks() {
        document.addEventListener('click', (e) => {
            if (e.target.tagName === 'A' && e.target.href) {
                e.preventDefault();
                this.openExternalLink(e.target.href);
            }
        });
    }

    handleKeyDown(e) {
        if (e.key === 'Escape' && this.modal.classList.contains('show')) {
            e.preventDefault();
            e.stopPropagation();
        }
    }

    openExternalLink(url) {
        try {
            BrowserOpenURL(url);
        } catch (error) {
            console.error('打开外部链接失败:', error);
            window.open(url, '_blank');
        }
    }

    async loadProxies() {
        try {
            const [proxies, stats] = await Promise.all([GetAllProxies(), GetStats()]);
            this.proxies = proxies || [];
            this.updateProxyList();
            this.updateStats(stats);
            this.updateSelectionUI();
        } catch (error) {
            if (this.lastErrorTime && Date.now() - this.lastErrorTime < 30000) {
                return;
            }
            this.lastErrorTime = Date.now();
            console.error('加载代理失败:', error);
        }
    }

    updateProxyList() {
        this.proxyList.innerHTML = '';
        if (this.proxies.length === 0) {
            this.proxyList.innerHTML = '<div style="text-align: center; padding: 2rem; color: #6b7280;"><p>暂无代理配置</p></div>';
            return;
        }
        this.proxies.forEach(proxy => {
            const proxyElement = this.createProxyElement(proxy);
            this.proxyList.appendChild(proxyElement);
        });
    }

    createProxyElement(proxy) {
        const div = document.createElement('div');
        div.className = 'proxy-item';
        const isSelected = this.selectedProxies.has(proxy.id);
        const statusClass = proxy.running ? 'status-running' : 'status-stopped';
        const statusText = proxy.running ? '运行中' : '已停止';
        const actionText = proxy.running ? '停止' : '启动';
        const actionClass = proxy.running ? 'btn-danger' : 'btn-success';
        
        div.innerHTML = `
            <div class="proxy-checkbox">
                <input type="checkbox" ${isSelected ? 'checked' : ''} data-id="${proxy.id}">
            </div>
            <div class="proxy-info">
                <div class="proxy-name" title="${proxy.name}">${proxy.name}</div>
                <div class="proxy-endpoints-compact">
                    上游: ${proxy.upstream.protocol}://${proxy.upstream.address} | 本地: ${proxy.local.protocol}://${proxy.local.listen_ip}:${proxy.local.listen_port}
                </div>
                <div class="proxy-status-inline ${statusClass}">
                    <div class="status-dot"></div>
                    <span>${statusText}</span>
                </div>
            </div>
            <div class="proxy-actions-inline">
                <button class="btn ${actionClass} toggle-btn" data-id="${proxy.id}">${actionText}</button>
                <button class="btn btn-outline edit-btn" data-id="${proxy.id}">编辑</button>
                <button class="btn btn-danger delete-btn" data-id="${proxy.id}">删除</button>
            </div>
        `;

        const checkbox = div.querySelector('input[type="checkbox"]');
        const toggleBtn = div.querySelector('.toggle-btn');
        const editBtn = div.querySelector('.edit-btn');
        const deleteBtn = div.querySelector('.delete-btn');

        checkbox.addEventListener('change', () => this.toggleProxySelection(proxy.id));
        toggleBtn.addEventListener('click', () => this.toggleProxy(proxy.id, proxy.running));
        editBtn.addEventListener('click', () => this.showEditModal(proxy));
        deleteBtn.addEventListener('click', () => this.deleteProxy(proxy.id, proxy.name));

        return div;
    }

    updateStats(stats) {
        this.totalCount.textContent = stats.total || 0;
        this.runningCount.textContent = stats.running || 0;
    }

    toggleSelectAll() {
        if (this.selectAllBtn.checked) {
            this.proxies.forEach(proxy => this.selectedProxies.add(proxy.id));
        } else {
            this.selectedProxies.clear();
        }
        this.updateProxyList();
        this.updateSelectionUI();
    }

    toggleProxySelection(id) {
        if (this.selectedProxies.has(id)) {
            this.selectedProxies.delete(id);
        } else {
            this.selectedProxies.add(id);
        }
        this.updateSelectionUI();
    }

    updateSelectionUI() {
        const selectedCount = this.selectedProxies.size;
        const totalCount = this.proxies.length;
        this.selectedCount.textContent = `已选择 ${selectedCount} 项`;
        this.selectAllBtn.checked = selectedCount === totalCount && totalCount > 0;
        this.selectAllBtn.indeterminate = selectedCount > 0 && selectedCount < totalCount;
        this.startSelectedBtn.disabled = selectedCount === 0;
        this.stopSelectedBtn.disabled = selectedCount === 0;
        this.deleteSelectedBtn.disabled = selectedCount === 0;
    }

    async startSelectedProxies() {
        const selectedIds = Array.from(this.selectedProxies);
        if (selectedIds.length === 0) return;
        for (const id of selectedIds) {
            try {
                await StartProxy(id);
            } catch (error) {
                console.error(`启动代理 ${id} 失败:`, error);
            }
        }
        await this.loadProxies();
    }

    async stopSelectedProxies() {
        const selectedIds = Array.from(this.selectedProxies);
        if (selectedIds.length === 0) return;
        for (const id of selectedIds) {
            try {
                await StopProxy(id);
            } catch (error) {
                console.error(`停止代理 ${id} 失败:`, error);
            }
        }
        await this.loadProxies();
    }

    async deleteSelectedProxies() {
        const selectedIds = Array.from(this.selectedProxies);
        if (selectedIds.length === 0) return;
        
        const selectedNames = selectedIds.map(id => {
            const proxy = this.proxies.find(p => p.id === id);
            return proxy ? proxy.name : id;
        });
        
        if (!confirm(`确定要删除选中的 ${selectedIds.length} 个代理吗？\n${selectedNames.join(', ')}`)) return;
        
        for (const id of selectedIds) {
            try {
                await DeleteProxy(id);
                this.selectedProxies.delete(id);
            } catch (error) {
                console.error(`删除代理 ${id} 失败:`, error);
            }
        }
        await this.loadProxies();
    }

    async toggleProxy(id, isRunning) {
        try {
            if (isRunning) {
                await StopProxy(id);
            } else {
                await StartProxy(id);
            }
            await this.loadProxies();
        } catch (error) {
            console.error('切换代理状态失败:', error);
        }
    }

    async deleteProxy(id, name) {
        if (!confirm(`确定要删除代理 "${name}" 吗？`)) return;
        try {
            await DeleteProxy(id);
            this.selectedProxies.delete(id);
            await this.loadProxies();
        } catch (error) {
            console.error('删除代理失败:', error);
        }
    }

    showAddModal() {
        this.modalTitle.textContent = '添加新代理';
        this.currentEditingProxy = null;
        this.proxyForm.reset();
        document.getElementById('localIP').value = '127.0.0.1';
        document.getElementById('proxyEnabled').checked = true;
        this.showModal();
    }

    showEditModal(proxy) {
        this.modalTitle.textContent = '编辑代理';
        this.currentEditingProxy = proxy;
        document.getElementById('proxyName').value = proxy.name;
        document.getElementById('upstreamProtocol').value = proxy.upstream.protocol;
        document.getElementById('upstreamAddress').value = proxy.upstream.address;
        document.getElementById('upstreamUsername').value = proxy.upstream.username || '';
        document.getElementById('upstreamPassword').value = proxy.upstream.password || '';
        document.getElementById('localProtocol').value = proxy.local.protocol;
        document.getElementById('localIP').value = proxy.local.listen_ip;
        document.getElementById('localPort').value = proxy.local.listen_port;
        document.getElementById('proxyEnabled').checked = proxy.enabled;
        this.showModal();
    }

    showModal() {
        this.modal.classList.add('show');
    }

    hideModal() {
        this.modal.classList.remove('show');
        this.currentEditingProxy = null;
    }

    async handleFormSubmit(e) {
        e.preventDefault();
        try {
            const formData = new FormData(this.proxyForm);
            const proxy = {
                id: this.currentEditingProxy ? this.currentEditingProxy.id : '',
                name: formData.get('name'),
                upstream: {
                    protocol: formData.get('upstream.protocol'),
                    address: formData.get('upstream.address'),
                    username: formData.get('upstream.username') || '',
                    password: formData.get('upstream.password') || '',
                    auth_method: 'basic'
                },
                local: {
                    protocol: formData.get('local.protocol'),
                    listen_ip: formData.get('local.listen_ip'),
                    listen_port: parseInt(formData.get('local.listen_port'))
                },
                enabled: document.getElementById('proxyEnabled').checked
            };
            
            if (this.currentEditingProxy) {
                await UpdateProxy(proxy);
            } else {
                await AddProxy(proxy);
            }
            
            this.hideModal();
            await this.loadProxies();
        } catch (error) {
            console.error('保存代理失败:', error);
        }
    }

    async exportConfig() {
        try {
            await ExportConfigToFile();
        } catch (error) {
            console.error('导出配置失败:', error);
        }
    }

    importConfig() {
        this.importConfigFromFile();
    }

    async importConfigFromFile() {
        try {
            await ImportConfigFromFile();
            await this.loadProxies();
        } catch (error) {
            console.error('导入配置失败:', error);
        }
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new ProxyManager();
});
