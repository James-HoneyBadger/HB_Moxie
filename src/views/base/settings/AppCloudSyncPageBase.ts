import { ref, computed } from 'vue';

import { useSettingsStore } from '@/stores/setting.ts';

import { keysIfValueEquals } from '@/core/base.ts';
import type { ApplicationCloudSetting } from '@/core/setting.ts';

export interface CategorizedApplicationCloudSettingItems {
    readonly categoryName: string;
    readonly categorySubName?: string;
    readonly items: ApplicationCloudSettingItem[];
}

export interface ApplicationCloudSettingItem {
    readonly settingKey: string;
    readonly relatedSettingKeys?: string[];
    readonly settingName: string;
}

export const ALL_APPLICATION_CLOUD_SETTINGS: CategorizedApplicationCloudSettingItems[] = [
    {
        categoryName: 'Basic Settings',
        items: [
            { settingKey: 'showAccountBalance', settingName: 'Show Account Balance' },
            { settingKey: 'autoUpdateExchangeRatesData', settingName: 'Auto-update Exchange Rates Data' }
        ]
    },
    {
        categoryName: 'Navigation Bar',
        items: [
            { settingKey: 'showAddTransactionButtonInDesktopNavbar', settingName: 'Show Add Transaction Button' }
        ]
    },
    {
        categoryName: 'Overview Page',
        items: [
            { settingKey: 'showAmountInHomePage', settingName: 'Show Amount' },
            { settingKey: 'timezoneUsedForStatisticsInHomePage', settingName: 'Timezone Used for Statistics' },
            { settingKey: 'overviewAccountFilterInHomePage', settingName: 'Accounts Included in Overview Statistics' },
            { settingKey: 'overviewTransactionCategoryFilterInHomePage', settingName: 'Transaction Categories Included in Overview Statistics' }
        ]
    },
    {
        categoryName: 'Transaction List Page',
        items: [
            { settingKey: 'itemsCountInTransactionListPage', settingName: 'Transactions Per Page' },
            { settingKey: 'showTotalAmountInTransactionListPage', settingName: 'Show Monthly Total Amount' },
            { settingKey: 'showTagInTransactionListPage', settingName: 'Show Transaction Tags' }
        ]
    },
    {
        categoryName: 'Transaction Edit Page',
        items: [
            { settingKey: 'autoSaveTransactionDraft', settingName: 'Automatically Save Draft' },
            { settingKey: 'autoGetCurrentGeoLocation', settingName: 'Automatically Add Geolocation' }
        ]
    },
    {
        categoryName: 'Import Transaction Dialog',
        items: [
            { settingKey: 'rememberLastSelectedFileTypeInImportTransactionDialog', relatedSettingKeys: ['lastSelectedFileTypeInImportTransactionDialog'], settingName: 'Remember Last Selected File Type' }
        ]
    },
    {
        categoryName: 'Insights Explorer Page',
        items: [
            { settingKey: 'insightsExplorerDefaultDateRangeType', settingName: 'Default Date Range' },
            { settingKey: 'showTagInInsightsExplorerPage', settingName: 'Show Transaction Tags' }
        ]
    },
    {
        categoryName: 'Account List Page',
        items: [
            { settingKey: 'totalAmountExcludeAccountIds', settingName: 'Accounts Included in Total' },
            { settingKey: 'accountCategoryOrders', settingName: 'Account Category Order' },
            { settingKey: 'hideCategoriesWithoutAccounts', settingName: 'Hide Categories Without Accounts' }
        ]
    },
    {
        categoryName: 'Exchange Rates Data Page',
        items: [
            { settingKey: 'currencySortByInExchangeRatesPage', settingName: 'Sort by' }
        ]
    },
    {
        categoryName: 'Data Cache',
        items: [
            { settingKey: 'exchangeRatesDataCacheExpiration', settingName: 'Cache Expiration for Exchange Rates Data' }
        ]
    },
    {
        categoryName: 'Statistics Settings',
        categorySubName: 'Common Settings',
        items: [
            { settingKey: 'statistics.defaultChartDataType', settingName: 'Default Chart Data Type' },
            { settingKey: 'statistics.defaultTimezoneType', settingName: 'Timezone Used for Date Range' },
            { settingKey: 'statistics.defaultAccountFilter', settingName: 'Default Account Filter' },
            { settingKey: 'statistics.defaultTransactionCategoryFilter', settingName: 'Default Transaction Category Filter' },
            { settingKey: 'statistics.defaultSortingType', settingName: 'Default Sort Order' }
        ]
    },
    {
        categoryName: 'Statistics Settings',
        categorySubName: 'Categorical Analysis Settings',
        items: [
            { settingKey: 'statistics.defaultCategoricalChartType', settingName: 'Default Chart Type' },
            { settingKey: 'statistics.defaultCategoricalChartDataRangeType', settingName: 'Default Date Range' }
        ]
    },
    {
        categoryName: 'Statistics Settings',
        categorySubName: 'Trend Analysis Settings',
        items: [
            { settingKey: 'statistics.defaultTrendChartType', settingName: 'Default Chart Type' },
            { settingKey: 'statistics.defaultTrendChartDataRangeType', settingName: 'Default Date Range' }
        ]
    },
    {
        categoryName: 'Statistics Settings',
        categorySubName: 'Asset Trends Settings',
        items: [
            { settingKey: 'statistics.defaultAssetTrendsChartType', settingName: 'Default Chart Type' },
            { settingKey: 'statistics.defaultAssetTrendsChartDataRangeType', settingName: 'Default Date Range' }
        ]
    }
];

export function useAppCloudSyncBase() {
    const settingsStore = useSettingsStore();

    const loading = ref<boolean>(false);
    const enabling = ref<boolean>(false);
    const disabling = ref<boolean>(false);
    const enabledApplicationCloudSettings = ref<Record<string, boolean>>(Object.assign({}, settingsStore.syncedAppSettings));

    const isEnableCloudSync = computed<boolean>(() => settingsStore.enableApplicationCloudSync);

    const hasEnabledApplicationCloudSettings = computed<boolean>(() => {
        for (const _ of keysIfValueEquals(enabledApplicationCloudSettings.value, true)) {
            return true;
        }

        return false;
    });

    const enabledApplicationCloudSettingKeys = computed<string[]>(() => {
        const keys: string[] = [];

        for (const key of keysIfValueEquals(enabledApplicationCloudSettings.value, true)) {
            keys.push(key);
        }

        return keys;
    });

    function isAllSettingsSelected(categorizedItems: CategorizedApplicationCloudSettingItems): boolean {
        for (const item of categorizedItems.items) {
            if (!enabledApplicationCloudSettings.value[item.settingKey]) {
                return false;
            }
        }

        return true;
    }

    function hasSettingSelectedButNotAllChecked(categorizedItems: CategorizedApplicationCloudSettingItems): boolean {
        let checkedCount = 0;

        for (const item of categorizedItems.items) {
            if (!enabledApplicationCloudSettings.value[item.settingKey]) {
                checkedCount++;
            }
        }

        return checkedCount > 0 && checkedCount < categorizedItems.items.length;
    }

    function updateSettingsSelected(categorizedItems: CategorizedApplicationCloudSettingItems, value: boolean): void {
        for (const item of categorizedItems.items) {
            enabledApplicationCloudSettings.value[item.settingKey] = value;

            if (item.relatedSettingKeys) {
                for (const relatedKey of item.relatedSettingKeys) {
                    enabledApplicationCloudSettings.value[relatedKey] = value;
                }
            }
        }
    }

    function updateSettingSelected(settingItem: ApplicationCloudSettingItem, value: boolean): void {
        enabledApplicationCloudSettings.value[settingItem.settingKey] = value;

        if (settingItem.relatedSettingKeys) {
            for (const relatedKey of settingItem.relatedSettingKeys) {
                enabledApplicationCloudSettings.value[relatedKey] = value;
            }
        }
    }

    function selectAllSettings(): void {
        for (const categorizedItems of ALL_APPLICATION_CLOUD_SETTINGS) {
            for (const item of categorizedItems.items) {
                enabledApplicationCloudSettings.value[item.settingKey] = true;

                if (item.relatedSettingKeys) {
                    for (const relatedKey of item.relatedSettingKeys) {
                        enabledApplicationCloudSettings.value[relatedKey] = true;
                    }
                }
            }
        }
    }

    function selectNoneSettings(): void {
        for (const categorizedItems of ALL_APPLICATION_CLOUD_SETTINGS) {
            for (const item of categorizedItems.items) {
                enabledApplicationCloudSettings.value[item.settingKey] = false;

                if (item.relatedSettingKeys) {
                    for (const relatedKey of item.relatedSettingKeys) {
                        enabledApplicationCloudSettings.value[relatedKey] = false;
                    }
                }
            }
        }
    }

    function selectInvertSettings(): void {
        for (const categorizedItems of ALL_APPLICATION_CLOUD_SETTINGS) {
            for (const item of categorizedItems.items) {
                const newValue = !enabledApplicationCloudSettings.value[item.settingKey];
                enabledApplicationCloudSettings.value[item.settingKey] = newValue;

                if (item.relatedSettingKeys) {
                    for (const relatedKey of item.relatedSettingKeys) {
                        enabledApplicationCloudSettings.value[relatedKey] = newValue;
                    }
                }
            }
        }
    }

    function setUserApplicationCloudSettings(settings: ApplicationCloudSetting[] | false) {
        if (settings && settings.length > 0) {
            settingsStore.setApplicationSettingsFromCloudSettings(settings);

            for (const setting of settings) {
                if (setting && setting.settingKey) {
                    enabledApplicationCloudSettings.value[setting.settingKey] = true;
                }
            }
        } else {
            settingsStore.setApplicationSettingsFromCloudSettings(undefined);
            enabledApplicationCloudSettings.value = {};
        }
    }

    return {
        // constants
        ALL_APPLICATION_CLOUD_SETTINGS,
        // states
        loading,
        enabling,
        disabling,
        enabledApplicationCloudSettings,
        // computed states
        isEnableCloudSync,
        hasEnabledApplicationCloudSettings,
        enabledApplicationCloudSettingKeys,
        // functions
        isAllSettingsSelected,
        hasSettingSelectedButNotAllChecked,
        updateSettingsSelected,
        updateSettingSelected,
        selectAllSettings,
        selectNoneSettings,
        selectInvertSettings,
        setUserApplicationCloudSettings
    };
}
