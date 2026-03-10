const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('hbmoxie', {
    platform: process.platform,
    isDesktopApp: true,
    // The app:// protocol handler proxies API calls transparently,
    // so the Vue app just needs to know the origin for API calls
    apiBaseUrl: 'app://hbmoxie',

    // Desktop notifications — send from renderer to main process
    showNotification: (title, body) => {
        ipcRenderer.send('show-notification', { title, body });
    },

    // Listen for quick-add transaction shortcut from main process
    onQuickAddTransaction: (callback) => {
        ipcRenderer.on('quick-add-transaction', () => callback());
    }
});
