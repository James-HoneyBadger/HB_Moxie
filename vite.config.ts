import fs from 'fs';
import { resolve } from 'path';

import { type UserConfig, defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue';
import vuetify from 'vite-plugin-vuetify';
import Checker from 'vite-plugin-checker';
import git from 'git-rev-sync';

import packageFile from './package.json';
import thirdPartyLicenseFile from './third-party-dependencies.json';

const SRC_DIR = resolve(__dirname, './src');
const PUBLIC_DIR = resolve(__dirname, './public');
const BUILD_DIR = resolve(__dirname, './dist',);

export default defineConfig(() => {
    const licenseContent = fs.readFileSync('./LICENSE', { encoding: 'utf-8' });
    const buildUnixTime = process.env['buildUnixTime'] || '';

    const options: UserConfig = {
        root: SRC_DIR,
        publicDir: PUBLIC_DIR,
        base: './',
        define: {
            __HBMOXIE_IS_PRODUCTION__: process.env['NODE_ENV'] === 'production',
            __HBMOXIE_VERSION__: JSON.stringify(packageFile.version),
            __HBMOXIE_BUILD_UNIX_TIME__: JSON.stringify(buildUnixTime),
            __HBMOXIE_BUILD_COMMIT_HASH__: JSON.stringify(git.short()),
            __HBMOXIE_LICENSE__: JSON.stringify(licenseContent),
            __HBMOXIE_THIRD_PARTY_LICENSES__: JSON.stringify(thirdPartyLicenseFile)
        },
        plugins: [
            vue(),
            vuetify({
                styles: {
                    configFile: 'styles/desktop/configured-variables/_vuetify.scss'
                }
            }),
            Checker({
                vueTsc: true
            })
        ],
        build: {
            target: [
                'chrome91'
            ],
            outDir: BUILD_DIR,
            sourcemap: false,
            assetsInlineLimit: 0,
            emptyOutDir: true,
            rollupOptions: {
                input: {
                    desktop: resolve(SRC_DIR, 'desktop.html'),
                },
                output: {
                    assetFileNames: assetInfo => {
                        const fileExt = assetInfo.names[0]?.split('.')[1];

                        if (!fileExt) {
                            throw new Error('Invalid asset file name.');
                        }

                        let assetType = fileExt;

                        if (/png|jpe?g|gif|tiff|bmp|ico/i.test(fileExt)) {
                            assetType = 'img';
                        } else if (/ttf|woff|woff2|svg|eot/i.test(fileExt)) {
                            assetType = 'fonts';
                        }

                        return `${assetType}/[name]-[hash][extname]`;
                    },
                    chunkFileNames: 'js/[name]-[hash].js',
                    entryFileNames: 'js/[name]-[hash].js',
                    manualChunks: id => {
                        if (/[\\/]node_modules[\\/]leaflet[\\/]/i.test(id)) {
                            return 'leaflet';
                        } else if (/[\\/]node_modules[\\/](moment|moment-timezone)[\\/]/i.test(id)) {
                            return 'moment';
                        } else if (/[\\/]node_modules[\\/](vuetify|vue-router|vue3-perfect-scrollbar|perfect-scrollbar|vuedraggable|sortablejs|@mdi.*)[\\/]/i.test(id)) {
                            return 'vendor-desktop';
                        } else if (/[\\/]node_modules[\\/](echarts|zrender|tslib|resize-detector|vue-echarts)[\\/]/i.test(id)) {
                            return 'vendor-desktop';
                        } else if (/plugin-vuetify:/i.test(id)) {
                            return 'vendor-desktop';
                        } else if (/[\\/]node_modules[\\/]/i.test(id)) {
                            return 'vendor-common';
                        } else if (/[\\/]src[\\/](core|consts|models|stores)[\\/]/i.test(id)) {
                            return 'common';
                        } else if (/[\\/]src[\\/]lib[\\/](map[\\/]|ui[\\/]common|[a-zA-Z0-9-_]+\.(js|ts))/i.test(id)) {
                            return 'common';
                        } else if (/[\\/]src[\\/]components[\\/](base|common)[\\/]/i.test(id)) {
                            return 'common';
                        } else if (/[\\/]src[\\/]views[\\/]base[\\/]/i.test(id)) {
                            return 'common';
                        } else if (/[\\/]src[\\/]locales[\\/]helpers\.(js|ts)/i.test(id)) {
                            return 'common';
                        } else if (/[\\/]src[\\/]locales[\\/]/i.test(id)) {
                            return 'locales';
                        } else {
                            return null;
                        }
                    }
                },
                treeshake: true
            },
        },
        resolve: {
            alias: {
                '@': SRC_DIR,
            },
        },
        server: {
            host: '0.0.0.0',
            port: 8081,
            strictPort: true,
            proxy: {
                '/server_settings.js': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                },
                '/desktop/server_settings.js': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                },
                '/oauth2': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                },
                '/api': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                },
                '/mcp': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                },
                '/avatar': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                },
                '/pictures': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                },
                '/proxy': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                },
                '/_AMapService': {
                    target: 'http://127.0.0.1:8080/',
                    changeOrigin: true
                }
            }
        },
    };

    return options;
})
