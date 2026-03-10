<template>
    <v-app>
        <router-view />
    </v-app>
    <v-snackbar class="cursor-pointer" color="notification-background" location="top"
                :multi-line="true" :timeout="-1" :close-on-content-click="true" v-model="showNotification">
        <v-tooltip activator="parent">{{ tt('Click to close') }}</v-tooltip>
        <div class="d-inline-flex">
            <img alt="logo" class="notification-logo" :src="APPLICATION_LOGO_PATH" />
            <span class="ms-2">{{ tt('global.app.title') }}</span>
        </div>
        <div>
            {{ currentNotificationContent }}
        </div>
    </v-snackbar>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';

import { useTheme } from 'vuetify';
import { useI18n } from '@/locales/helpers.ts';

import { useRouter } from 'vue-router';

import { useRootStore } from '@/stores/index.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useDesktopPageStore } from '@/stores/desktopPage.ts';

import { APPLICATION_LOGO_PATH } from '@/consts/asset.ts';
import { ThemeType } from '@/core/theme.ts';
import { initMapProvider } from '@/lib/map/index.ts';
import { getSystemTheme, setExpenseAndIncomeAmountColor } from '@/lib/ui/common.ts';

const { tt, getCurrentLanguageInfo, initLocale } = useI18n();

const theme = useTheme();

const rootStore = useRootStore();
const settingsStore = useSettingsStore();
const userStore = useUserStore();
const exchangeRatesStore = useExchangeRatesStore();
const desktopPageStore = useDesktopPageStore();
const router = useRouter();

const showNotification = ref<boolean>(false);

const currentNotificationContent = computed<string | null>(() => rootStore.currentNotification);

onMounted(() => {
    document.addEventListener('DOMContentLoaded', () => {
        const languageInfo = getCurrentLanguageInfo();
        initMapProvider(languageInfo?.alternativeLanguageTag);
    });
});

watch(currentNotificationContent, (newValue) => {
    showNotification.value = !!newValue;
});

if (settingsStore.appSettings.theme === ThemeType.Light) {
    theme.change(ThemeType.Light);
} else if (settingsStore.appSettings.theme === ThemeType.Dark) {
    theme.change(ThemeType.Dark);
} else {
    theme.change(getSystemTheme());
}

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function (e) {
    if (settingsStore.appSettings.theme === 'auto') {
        if (e.matches) {
            theme.change(ThemeType.Dark);
        } else {
            theme.change(ThemeType.Light);
        }
    }
});

const localeDefaultSettings = initLocale(userStore.currentUserLanguage, settingsStore.appSettings.timeZone);
settingsStore.updateLocalizedDefaultSettings(localeDefaultSettings);

setExpenseAndIncomeAmountColor(userStore.currentUserExpenseAmountColor, userStore.currentUserIncomeAmountColor);

// Auto-auth is handled by the router guard (ensureAuthenticated).
// Just initialize exchange rates on load.
exchangeRatesStore.removeExpiredExchangeRates(true);
exchangeRatesStore.autoUpdateExchangeRatesData();

// Listen for quick-add transaction from Electron global shortcut / tray menu
if (window.hbmoxie?.onQuickAddTransaction) {
    window.hbmoxie.onQuickAddTransaction(() => {
        desktopPageStore.setShowAddTransactionDialogInTransactionList();
        router.push('/transaction/list');
    });
}
</script>

<style>
.notification-logo {
    width: 1.2rem;
    height: 1.2rem;
}
</style>
