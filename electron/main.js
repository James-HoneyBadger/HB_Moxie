const { app, BrowserWindow, protocol, shell, net: electronNet, session, Tray, Menu, globalShortcut, Notification, nativeImage, ipcMain } = require('electron');
const path = require('path');

const { adoptLegacyConfigFile, migrateLegacyUserDataDir } = require('./userDataLayout.js');
const fs = require('fs');
const fsPromises = fs.promises;
const { spawn } = require('child_process');
const net = require('net');

app.setName('Honey Ledger');

let mainWindow = null;
let backendProcess = null;
let backendPort = 0;
let backendUrl = '';
let tray = null;
let isQuitting = false;
let appUpdater = null;

const CUSTOM_SCHEME = 'app';
const CUSTOM_HOST = 'honey-ledger';

// API paths that should be proxied to the Go backend
const PROXY_PATH_PREFIXES = [
    '/api/',
    '/server_settings.js',
    '/avatar/',
    '/pictures/',
    '/proxy/',
    '/qrcode/',
    '/mcp',
    '/oauth2/',
    '/healthz.json',
    '/_AMapService/'
];

// MIME type mapping for local file serving
const MIME_TYPES = {
    '.html': 'text/html',
    '.js': 'text/javascript',
    '.css': 'text/css',
    '.json': 'application/json',
    '.png': 'image/png',
    '.jpg': 'image/jpeg',
    '.jpeg': 'image/jpeg',
    '.gif': 'image/gif',
    '.svg': 'image/svg+xml',
    '.ico': 'image/x-icon',
    '.woff': 'font/woff',
    '.woff2': 'font/woff2',
    '.ttf': 'font/ttf',
    '.eot': 'application/vnd.ms-fontobject',
    '.map': 'application/json',
    '.webp': 'image/webp'
};

function getBackendBinaryName() {
    return process.platform === 'win32' ? 'honey-ledger.exe' : 'honey-ledger';
}

function getBackendPath() {
    if (app.isPackaged) {
        const packagedRoot = path.join(process.resourcesPath, 'backend');
        const packagedCandidates = [
            path.join(packagedRoot, getBackendBinaryName()),
            path.join(packagedRoot, 'honey-ledger'),
            path.join(packagedRoot, 'honey-ledger.exe'),
            path.join(packagedRoot, 'hbmoxie'),
            path.join(packagedRoot, 'hbmoxie.exe')
        ];

        for (const candidate of packagedCandidates) {
            if (fs.existsSync(candidate)) {
                return candidate;
            }
        }

        return packagedCandidates[0];
    }

    const developmentRoot = path.join(__dirname, '..');
    const developmentCandidates = [
        path.join(developmentRoot, getBackendBinaryName()),
        path.join(developmentRoot, 'honey-ledger_bin'),
        path.join(developmentRoot, 'honey-ledger'),
        path.join(developmentRoot, 'honey-ledger_bin.exe'),
        path.join(developmentRoot, 'hbmoxie_bin'),
        path.join(developmentRoot, 'hbmoxie')
    ];

    for (const candidate of developmentCandidates) {
        if (fs.existsSync(candidate)) {
            return candidate;
        }
    }

    return developmentCandidates[0];
}

function ensureDirectory(directoryPath) {
    fs.mkdirSync(directoryPath, { recursive: true });
}

function copyFileIfMissing(sourcePath, destinationPath) {
    if (fs.existsSync(destinationPath)) {
        return;
    }

    ensureDirectory(path.dirname(destinationPath));
    fs.copyFileSync(sourcePath, destinationPath);
}

/**
 * Create a writable per-user data directory and seed it with the packaged
 * configuration and templates. Application resources are read-only on
 * AppImage, macOS and Windows installations.
 */
function prepareUserDataDir() {
    const userDataDir = app.getPath('userData');
    const resourceRoot = path.join(process.resourcesPath, 'backend');

    migrateLegacyUserDataDir(userDataDir);

    for (const directoryName of ['data', 'log', 'storage', 'conf', 'templates']) {
        ensureDirectory(path.join(userDataDir, directoryName));
    }

    const configPath = adoptLegacyConfigFile(userDataDir);

    const sourceConfigPath = path.join(resourceRoot, 'conf', 'honey-ledger.ini');
    if (fs.existsSync(sourceConfigPath)) {
        copyFileIfMissing(sourceConfigPath, configPath);
    }

    const sourceTemplatesPath = path.join(resourceRoot, 'templates');
    const destinationTemplatesPath = path.join(userDataDir, 'templates');

    if (fs.existsSync(sourceTemplatesPath)) {
        fs.cpSync(sourceTemplatesPath, destinationTemplatesPath, { recursive: true, force: false });
    }

    return userDataDir;
}

function getBackendWorkingDir() {
    if (app.isPackaged) {
        return prepareUserDataDir();
    }
    return path.join(__dirname, '..');
}

function getStaticRootPath() {
    if (app.isPackaged) {
        return path.join(process.resourcesPath, 'backend', 'public');
    }
    return path.join(__dirname, '..', 'dist');
}

/**
 * Find a free port on localhost
 */
function findFreePort() {
    return new Promise((resolve, reject) => {
        const srv = net.createServer();
        srv.listen(0, '127.0.0.1', () => {
            const port = srv.address().port;
            srv.close(() => resolve(port));
        });
        srv.on('error', reject);
    });
}

function startBackend(port) {
    const backendPath = getBackendPath();
    const workingDir = getBackendWorkingDir();

    console.log(`[Honey Ledger] Starting backend: ${backendPath}`);
    console.log(`[Honey Ledger] Working directory: ${workingDir}`);
    console.log(`[Honey Ledger] Backend port: ${port}`);

    backendProcess = spawn(backendPath, ['server', 'run'], {
        cwd: workingDir,
        env: {
            ...process.env,
            HONEY_WORK_DIR: workingDir,
            HONEY_SERVER_HTTP_ADDR: '127.0.0.1',
            HONEY_SERVER_HTTP_PORT: String(port)
        },
        stdio: ['ignore', 'pipe', 'pipe']
    });

    backendProcess.stdout.on('data', (data) => {
        console.log(`[Honey Ledger Backend] ${data.toString().trim()}`);
    });

    backendProcess.stderr.on('data', (data) => {
        console.error(`[Honey Ledger Backend Error] ${data.toString().trim()}`);
    });

    backendProcess.on('close', (code) => {
        console.log(`[Honey Ledger] Backend process exited with code ${code}`);
        backendProcess = null;
    });

    backendProcess.on('error', (err) => {
        console.error(`[Honey Ledger] Failed to start backend:`, err);
    });
}

function waitForBackend(port, maxRetries = 60, interval = 500) {
    return new Promise((resolve, reject) => {
        let attempts = 0;
        const check = () => {
            attempts++;
            const socket = new net.Socket();
            socket.setTimeout(300);
            socket.on('connect', () => {
                socket.destroy();
                resolve();
            });
            socket.on('timeout', () => {
                socket.destroy();
                if (attempts >= maxRetries) {
                    reject(new Error('Backend did not start in time'));
                } else {
                    setTimeout(check, interval);
                }
            });
            socket.on('error', () => {
                socket.destroy();
                if (attempts >= maxRetries) {
                    reject(new Error('Backend did not start in time'));
                } else {
                    setTimeout(check, interval);
                }
            });
            socket.connect(port, '127.0.0.1');
        };
        check();
    });
}

/**
 * Determine whether a request URL should be proxied to the Go backend
 */
function shouldProxyToBackend(urlPath) {
    for (const prefix of PROXY_PATH_PREFIXES) {
        if (urlPath === prefix || urlPath.startsWith(prefix)) {
            return true;
        }
    }
    return false;
}

/**
 * Register the custom app:// protocol to serve local files and proxy API calls
 */
function registerProtocol() {
    protocol.handle(CUSTOM_SCHEME, async (request) => {
        const url = new URL(request.url);
        let urlPath = decodeURIComponent(url.pathname);

        // Proxy API requests to the Go backend
        if (shouldProxyToBackend(urlPath)) {
            const backendRequestUrl = `${backendUrl}${urlPath}${url.search}`;
            try {
                const response = await electronNet.fetch(backendRequestUrl, {
                    method: request.method,
                    headers: request.headers,
                    body: request.method !== 'GET' && request.method !== 'HEAD' ? request.body : undefined,
                    duplex: request.method !== 'GET' && request.method !== 'HEAD' ? 'half' : undefined
                });
                return response;
            } catch (err) {
                console.error(`[Honey Ledger] Proxy error for ${urlPath}:`, err);
                return new Response(JSON.stringify({ error: 'Backend unavailable' }), {
                    status: 502,
                    headers: { 'Content-Type': 'application/json' }
                });
            }
        }

        // Serve local static files
        const staticRoot = getStaticRootPath();

        // Map /desktop → desktop.html, / → desktop.html (Electron-only)
        if (urlPath === '/desktop' || urlPath === '/' || urlPath === '') {
            urlPath = '/desktop.html';
        }

        // Resolve file path
        const filePath = path.join(staticRoot, urlPath);

        // Security: prevent path traversal and sibling-directory prefix bypasses
        const relativeFilePath = path.relative(staticRoot, filePath);
        if (relativeFilePath.startsWith('..') || path.isAbsolute(relativeFilePath)) {
            return new Response('Forbidden', { status: 403 });
        }

        try {
            const data = await fsPromises.readFile(filePath);
            const ext = path.extname(filePath).toLowerCase();
            const mimeType = MIME_TYPES[ext] || 'application/octet-stream';

            return new Response(data, {
                status: 200,
                headers: { 'Content-Type': mimeType }
            });
        } catch (err) {
            // If file not found, try serving desktop.html for SPA routing
            if (err.code === 'ENOENT') {
                try {
                    const indexPath = path.join(staticRoot, 'desktop.html');
                    const indexData = await fsPromises.readFile(indexPath);
                    return new Response(indexData, {
                        status: 200,
                        headers: { 'Content-Type': 'text/html' }
                    });
                } catch {
                    return new Response('Not Found', { status: 404 });
                }
            }
            return new Response('Internal Server Error', { status: 500 });
        }
    });
}

function createTray() {
    const iconPath = path.join(__dirname, 'assets', 'icon.png');
    let trayIcon;

    try {
        trayIcon = nativeImage.createFromPath(iconPath);
        // Resize for tray (16x16 on most platforms, 22x22 on Linux)
        trayIcon = trayIcon.resize({ width: 22, height: 22 });
    } catch (err) {
        console.error('[Honey Ledger] Failed to load tray icon:', err);
        return;
    }

    tray = new Tray(trayIcon);
    tray.setToolTip('Honey Ledger');

    const contextMenu = Menu.buildFromTemplate([
        {
            label: 'Show Honey Ledger',
            click: () => {
                if (mainWindow) {
                    mainWindow.show();
                    mainWindow.focus();
                }
            }
        },
        {
            label: 'Quick Add Transaction',
            click: () => {
                if (mainWindow) {
                    mainWindow.show();
                    mainWindow.focus();
                    mainWindow.webContents.send('quick-add-transaction');
                }
            }
        },
        { type: 'separator' },
        ...(appUpdater
            ? [
                  {
                      label: 'Check for Updates',
                      click: () => {
                          appUpdater.checkForUpdates().catch(() => {});
                      }
                  },
                  { type: 'separator' }
              ]
            : []),
        {
            label: 'Quit',
            click: () => {
                isQuitting = true;
                app.quit();
            }
        }
    ]);

    tray.setContextMenu(contextMenu);

    tray.on('click', () => {
        if (mainWindow) {
            if (mainWindow.isVisible()) {
                mainWindow.focus();
            } else {
                mainWindow.show();
                mainWindow.focus();
            }
        }
    });
}

function registerGlobalShortcuts() {
    // Ctrl+Shift+M (or Cmd+Shift+M on macOS) — Show/hide Honey Ledger
    globalShortcut.register('CommandOrControl+Shift+M', () => {
        if (mainWindow) {
            if (mainWindow.isVisible() && mainWindow.isFocused()) {
                mainWindow.hide();
            } else {
                mainWindow.show();
                mainWindow.focus();
            }
        }
    });

    // Ctrl+Shift+N (or Cmd+Shift+N on macOS) — Quick add transaction
    globalShortcut.register('CommandOrControl+Shift+N', () => {
        if (mainWindow) {
            mainWindow.show();
            mainWindow.focus();
            mainWindow.webContents.send('quick-add-transaction');
        }
    });
}

function showNotification(title, body) {
    if (Notification.isSupported()) {
        const notification = new Notification({
            title: title,
            body: body,
            icon: path.join(__dirname, 'assets', 'icon.png')
        });
        notification.on('click', () => {
            if (mainWindow) {
                mainWindow.show();
                mainWindow.focus();
            }
        });
        notification.show();
    }
}

/**
 * Check GitHub releases in the background and install a downloaded update on
 * quit. Only the AppImage build can replace itself: .deb packages are updated
 * by the system package manager, and macOS self-update requires a signed and
 * notarized bundle, which this project does not ship yet.
 */
function setupAutoUpdate() {
    if (!app.isPackaged || process.env.HONEY_DISABLE_UPDATE_CHECK === '1' || process.env.RATEL_DISABLE_UPDATE_CHECK === '1') {
        return null;
    }

    const isAppImage = process.platform === 'linux' && !!process.env.APPIMAGE;

    if (!isAppImage && process.platform !== 'win32') {
        console.log('[Honey Ledger] Auto-update is not available for this installation type');
        return null;
    }

    let autoUpdater;

    try {
        ({ autoUpdater } = require('electron-updater'));
    } catch (err) {
        console.warn('[Honey Ledger] Auto-update unavailable:', err?.message ?? err);
        return null;
    }

    autoUpdater.autoDownload = true;
    autoUpdater.autoInstallOnAppQuit = true;

    // Update problems must never affect the running application.
    autoUpdater.on('error', (err) => {
        console.warn('[Honey Ledger] Update check failed:', err?.message ?? err);
    });

    autoUpdater.on('update-downloaded', (info) => {
        console.log(`[Honey Ledger] Update ${info.version} downloaded`);
        showNotification(
            'Honey Ledger update ready',
            `Version ${info.version} will be installed the next time you quit.`
        );
    });

    // Failures arrive through the 'error' event above; swallowing the rejection
    // here avoids logging the same failure twice.
    autoUpdater.checkForUpdates().catch(() => {});

    return autoUpdater;
}

function createMainWindow() {
    // Enforce Content Security Policy
    session.defaultSession.webRequest.onHeadersReceived((details, callback) => {
        callback({
            responseHeaders: {
                ...details.responseHeaders,
                'Content-Security-Policy': [
                    "default-src 'self' app://honey-ledger; " +
                    "script-src 'self' app://honey-ledger; " +
                    "style-src 'self' 'unsafe-inline' app://honey-ledger; " +
                    "img-src 'self' app://honey-ledger data: blob:; " +
                    "font-src 'self' app://honey-ledger; " +
                    "connect-src 'self' app://honey-ledger; " +
                    "object-src 'none'; " +
                    "base-uri 'self'"
                ]
            }
        });
    });

    // Deny all permission requests (camera, microphone, geolocation, etc.)
    session.defaultSession.setPermissionRequestHandler((webContents, permission, callback) => {
        callback(false);
    });

    mainWindow = new BrowserWindow({
        width: 1400,
        height: 900,
        minWidth: 800,
        minHeight: 600,
        title: 'Honey Ledger',
        icon: path.join(__dirname, 'assets', 'icon.png'),
        show: false,
        webPreferences: {
            nodeIntegration: false,
            contextIsolation: true,
            sandbox: true,
            preload: path.join(__dirname, 'preload.js')
        }
    });

    mainWindow.loadURL(`${CUSTOM_SCHEME}://${CUSTOM_HOST}/desktop`);

    mainWindow.webContents.on('did-fail-load', (event, errorCode, errorDescription, validatedURL) => {
        console.error(`[Honey Ledger] Failed to load: ${validatedURL} - ${errorDescription} (${errorCode})`);
    });

    mainWindow.webContents.setWindowOpenHandler(({ url }) => {
        if (url.startsWith('http://') || url.startsWith('https://')) {
            shell.openExternal(url);
        }
        return { action: 'deny' };
    });

    // Prevent navigation away from the app origin
    mainWindow.webContents.on('will-navigate', (event, navigationUrl) => {
        const parsedUrl = new URL(navigationUrl);
        if (parsedUrl.protocol !== `${CUSTOM_SCHEME}:`) {
            event.preventDefault();
            if (navigationUrl.startsWith('http://') || navigationUrl.startsWith('https://')) {
                shell.openExternal(navigationUrl);
            }
        }
    });

    mainWindow.once('ready-to-show', () => {
        mainWindow.show();
        mainWindow.focus();
    });

    // Minimize to tray instead of closing (unless explicitly quitting)
    mainWindow.on('close', (event) => {
        if (!isQuitting) {
            event.preventDefault();
            mainWindow.hide();
        }
    });

    mainWindow.on('closed', () => {
        mainWindow = null;
    });
}

function stopBackend() {
    if (backendProcess) {
        console.log('[Honey Ledger] Stopping backend...');
        backendProcess.kill('SIGTERM');
        setTimeout(() => {
            if (backendProcess) {
                backendProcess.kill('SIGKILL');
            }
        }, 3000);
    }
}

// Handle uncaught errors gracefully
process.on('uncaughtException', (error) => {
    console.error('[Honey Ledger] Uncaught exception:', error);
});
process.on('unhandledRejection', (reason) => {
    console.error('[Honey Ledger] Unhandled rejection:', reason);
});

// Register the custom scheme as privileged (must happen before app ready)
protocol.registerSchemesAsPrivileged([
    {
        scheme: CUSTOM_SCHEME,
        privileges: {
            standard: true,
            secure: true,
            supportFetchAPI: true,
            corsEnabled: true,
            stream: true
        }
    }
]);

const hasSingleInstanceLock = app.requestSingleInstanceLock();

if (!hasSingleInstanceLock) {
    app.quit();
} else {
    app.on('second-instance', () => {
        if (mainWindow) {
            if (mainWindow.isMinimized()) {
                mainWindow.restore();
            }

            mainWindow.show();
            mainWindow.focus();
        }
    });
}

app.whenReady().then(async () => {
    if (!hasSingleInstanceLock) {
        return;
    }

    // Find a free port and start the backend
    backendPort = await findFreePort();
    backendUrl = `http://127.0.0.1:${backendPort}`;

    console.log(`[Honey Ledger] Using port ${backendPort}`);

    // Register the custom protocol handler
    registerProtocol();

    // Start the Go backend on the chosen port
    startBackend(backendPort);

    try {
        await waitForBackend(backendPort);
        console.log('[Honey Ledger] Backend is ready');

        // Check for a newer release before building the tray menu so the
        // "Check for Updates" entry is only shown where it can work.
        appUpdater = setupAutoUpdate();

        // Create the system tray
        createTray();

        // Register global keyboard shortcuts
        registerGlobalShortcuts();

        // Set up IPC handlers for renderer → main communication
        ipcMain.on('show-notification', (event, { title, body }) => {
            showNotification(title, body);
        });

        createMainWindow();
    } catch (err) {
        console.error('[Honey Ledger] Backend failed to start:', err);
        app.quit();
    }
});

app.on('window-all-closed', () => {
    // On macOS, keep the app in the tray; on other platforms,
    // this event won't fire since we intercept close → hide
});

app.on('before-quit', () => {
    isQuitting = true;
    globalShortcut.unregisterAll();
    stopBackend();
});

app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
        createMainWindow();
    }
});
