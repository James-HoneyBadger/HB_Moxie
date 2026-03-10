<template>
    <v-row>
        <v-col cols="12">
            <v-card :class="{ 'disabled': loadingDataStatistics }">
                <template #title>
                    <div class="d-flex align-center">
                        <span>{{ tt('Data Management') }}</span>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :loading="loadingDataStatistics" @click="reloadUserDataStatistics(true)">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                    </div>
                </template>

                <v-card-text>
                    <v-row>
                        <v-col cols="6" sm="3" :key="idx" v-for="(item, idx) in [
                            {
                                title: 'Transactions',
                                count: displayDataStatistics ? displayDataStatistics.totalTransactionCount : '-',
                                icon: mdiListBoxOutline,
                                color: 'info-darken-1'
                            },
                            {
                                title: 'Transaction Pictures',
                                count: displayDataStatistics ? displayDataStatistics.totalTransactionPictureCount : '-',
                                icon: mdiImage,
                                color: 'error-darken-1'
                            },
                            {
                                title: 'Insights Explorers',
                                count: displayDataStatistics ? displayDataStatistics.totalInsightsExplorerCount : '-',
                                icon: mdiCompassOutline,
                                color: 'warning'
                            },
                            {
                                title: 'Accounts',
                                count: displayDataStatistics ? displayDataStatistics.totalAccountCount : '-',
                                icon: mdiCreditCardOutline,
                                color: 'primary'
                            },
                            {
                                title: 'Transaction Categories',
                                count: displayDataStatistics ? displayDataStatistics.totalTransactionCategoryCount : '-',
                                icon: mdiViewDashboardOutline,
                                color: 'teal'
                            },
                            {
                                title: 'Transaction Tags',
                                count: displayDataStatistics ? displayDataStatistics.totalTransactionTagCount : '-',
                                icon: mdiTagOutline,
                                color: 'grey'
                            },
                            {
                                title: 'Transaction Templates',
                                count: displayDataStatistics ? displayDataStatistics.totalTransactionTemplateCount : '-',
                                icon: mdiClipboardTextOutline,
                                color: 'secondary-darken-1'
                            },
                            {
                                title: 'Scheduled Transactions',
                                count: displayDataStatistics ? displayDataStatistics.totalScheduledTransactionCount : '-',
                                icon: mdiClipboardTextClockOutline,
                                color: 'success-darken-1'
                            }
                        ]">
                            <div class="d-flex align-center">
                                <div class="me-3">
                                    <v-avatar rounded :color="item.color" size="42" class="elevation-1">
                                        <v-icon size="24" :icon="item.icon"/>
                                    </v-avatar>
                                </div>

                                <div class="d-flex flex-column">
                                    <span class="text-caption">{{ tt(item.title) }}</span>
                                    <v-skeleton-loader class="skeleton-no-margin pt-2 pb-2" type="text" style="width: 60px" :loading="true" v-if="loadingDataStatistics"></v-skeleton-loader>
                                    <span class="text-xl" v-if="!loadingDataStatistics">{{ item.count }}</span>
                                </div>
                            </div>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" v-if="isDataExportingEnabled()">
            <v-card :class="{ 'disabled': exportingData }" :title="tt('Export Data')">
                <v-card-text>
                    <span class="text-body-1">{{ tt('Export all transaction data to file.') }}&nbsp;{{ tt('It may take a long time, please wait for a few minutes.') }}</span>
                </v-card-text>

                <v-card-text class="d-flex flex-wrap gap-4">
                    <v-btn-group variant="elevated" density="comfortable" color="primary">
                        <v-btn :disabled="loadingDataStatistics || exportingData || !dataStatistics || !dataStatistics.totalTransactionCount || dataStatistics.totalTransactionCount === '0'">
                            {{ tt('Export Data') }}
                            <v-progress-circular indeterminate size="22" class="ms-2" v-if="exportingData"></v-progress-circular>
                            <v-menu activator="parent">
                                <v-list :disabled="loadingDataStatistics || exportingData || !dataStatistics || !dataStatistics.totalTransactionCount || dataStatistics.totalTransactionCount === '0'">
                                    <v-list-item @click="exportData('csv')">
                                        <v-list-item-title>{{ tt('CSV (Comma-separated values) File') }}</v-list-item-title>
                                    </v-list-item>
                                    <v-list-item @click="exportData('tsv')">
                                        <v-list-item-title>{{ tt('TSV (Tab-separated values) File') }}</v-list-item-title>
                                    </v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                    </v-btn-group>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" v-if="isBackupEnabled()">
            <v-card :class="{ 'disabled': backupOperationInProgress }">
                <template #title>
                    <div class="d-flex align-center">
                        <v-icon :icon="mdiArchiveOutline" class="me-2" />
                        <span>{{ tt('Backups & Archives') }}</span>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :loading="loadingBackups" @click="reloadBackupList()">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                    </div>
                </template>

                <v-card-text>
                    <span class="text-body-1">{{ tt('Create a full backup of your database, storage files, and configuration. Backups can be downloaded or restored.') }}</span>
                </v-card-text>

                <v-card-text class="pt-0">
                    <v-row class="mb-2">
                        <v-col cols="12" md="6">
                            <v-text-field
                                autocomplete="current-password"
                                type="password"
                                variant="underlined"
                                :disabled="backupOperationInProgress"
                                :placeholder="tt('Password (required for create/restore)')"
                                v-model="backupPassword"
                            />
                        </v-col>
                    </v-row>
                </v-card-text>

                <v-card-text class="d-flex flex-wrap gap-4 pt-0">
                    <v-btn color="primary" :disabled="!backupPassword || backupOperationInProgress" @click="createNewBackup">
                        <v-icon :icon="mdiPlus" class="me-1" />
                        {{ tt('Create Backup') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="creatingBackup"></v-progress-circular>
                    </v-btn>
                </v-card-text>

                <v-card-text v-if="loadingBackups" class="text-center">
                    <v-progress-circular indeterminate />
                </v-card-text>

                <v-card-text v-if="!loadingBackups && backupList.length === 0">
                    <span class="text-body-2 text-grey">{{ tt('No backups found.') }}</span>
                </v-card-text>

                <v-card-text v-if="!loadingBackups && backupList.length > 0" class="pt-0">
                    <v-table density="comfortable">
                        <thead>
                            <tr>
                                <th>{{ tt('File Name') }}</th>
                                <th>{{ tt('Size') }}</th>
                                <th>{{ tt('Created') }}</th>
                                <th class="text-right">{{ tt('Actions') }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="backup in backupList" :key="backup.fileName">
                                <td>
                                    <v-icon :icon="mdiArchiveOutline" size="16" class="me-1" />
                                    {{ backup.fileName }}
                                </td>
                                <td>{{ formatFileSize(backup.fileSize) }}</td>
                                <td>{{ formatBackupDate(backup.createdAt) }}</td>
                                <td class="text-right">
                                    <v-btn density="compact" variant="text" color="primary" size="small"
                                           :disabled="backupOperationInProgress" @click="downloadExistingBackup(backup.fileName)">
                                        <v-icon :icon="mdiDownload" size="18" />
                                        <v-tooltip activator="parent">{{ tt('Download') }}</v-tooltip>
                                    </v-btn>
                                    <v-btn density="compact" variant="text" color="warning" size="small"
                                           :disabled="!backupPassword || backupOperationInProgress" @click="restoreFromBackup(backup.fileName)">
                                        <v-icon :icon="mdiRestore" size="18" />
                                        <v-tooltip activator="parent">{{ tt('Restore') }}</v-tooltip>
                                    </v-btn>
                                    <v-btn density="compact" variant="text" color="error" size="small"
                                           :disabled="backupOperationInProgress" @click="deleteExistingBackup(backup.fileName)">
                                        <v-icon :icon="mdiDeleteOutline" size="18" />
                                        <v-tooltip activator="parent">{{ tt('Delete') }}</v-tooltip>
                                    </v-btn>
                                </td>
                            </tr>
                        </tbody>
                    </v-table>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12">
            <v-card :class="{ 'disabled': clearingData }">
                <template #title>
                    <span class="text-error">{{ tt('Danger Zone') }}</span>
                </template>

                <v-card-text class="py-0">
                    <span class="text-body-1 text-error">
                        <v-icon class="mt-n1" :icon="mdiAlert"/>
                        {{ tt('You CANNOT undo this action. "Clear All Transactions" will clear all your transactions data, and "Clear All Data" will clear your accounts, categories, tags and transactions data. Please enter your current password to confirm.') }}
                    </span>
                </v-card-text>

                <v-card-text class="pb-0">
                    <v-row class="mb-3">
                        <v-col cols="12" md="6">
                            <v-text-field
                                autocomplete="current-password"
                                ref="currentPasswordInput"
                                type="password"
                                variant="underlined"
                                color="error"
                                :disabled="loadingDataStatistics || clearingData"
                                :placeholder="tt('Current Password')"
                                v-model="currentPasswordForClearData"
                            />
                        </v-col>
                    </v-row>
                </v-card-text>

                <v-card-text class="d-flex flex-wrap gap-4">
                    <v-btn color="error" :disabled="loadingDataStatistics || !currentPasswordForClearData || clearingData">
                        {{ tt('Clear User Data') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="clearingData"></v-progress-circular>
                        <v-menu activator="parent">
                            <v-list :disabled="loadingDataStatistics || !currentPasswordForClearData || clearingData">
                                <v-list-item @click="clearAllTransactions">
                                    <v-list-item-title>{{ tt('Clear All Transactions') }}</v-list-item-title>
                                </v-list-item>
                                <v-list-item @click="clearAllData">
                                    <v-list-item-title>{{ tt('Clear All Data') }}</v-list-item-title>
                                </v-list-item>
                            </v-list>
                        </v-menu>
                    </v-btn>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useDataManagementPageBase } from '@/views/base/users/DataManagementPageBase.ts';

import { useRootStore } from '@/stores/index.ts';
import { useUserStore } from '@/stores/user.ts';

import { isEquals } from '@/lib/common.ts';
import { isDataExportingEnabled, isBackupEnabled } from '@/lib/server_settings.ts';
import { startDownloadFile } from '@/lib/ui/common.ts';

import type { BackupListResponse } from '@/models/backup.ts';

import {
    mdiRefresh,
    mdiListBoxOutline,
    mdiCreditCardOutline,
    mdiImage,
    mdiCompassOutline,
    mdiViewDashboardOutline,
    mdiTagOutline,
    mdiClipboardTextOutline,
    mdiClipboardTextClockOutline,
    mdiAlert,
    mdiArchiveOutline,
    mdiPlus,
    mdiDownload,
    mdiRestore,
    mdiDeleteOutline
} from '@mdi/js';

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const { dataStatistics, displayDataStatistics, getExportFileName } = useDataManagementPageBase();

const rootStore = useRootStore();
const userStore = useUserStore();

const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const loadingDataStatistics = ref<boolean>(true);
const exportingData = ref<boolean>(false);
const currentPasswordForClearData = ref<string>('');
const clearingData = ref<boolean>(false);

const backupList = ref<BackupListResponse[]>([]);
const loadingBackups = ref<boolean>(false);
const creatingBackup = ref<boolean>(false);
const backupPassword = ref<string>('');
const backupOperationInProgress = ref<boolean>(false);

function reloadUserDataStatistics(force: boolean): void {
    loadingDataStatistics.value = true;

    userStore.getUserDataStatistics().then(dataStatisticsResponse => {
        if (force) {
            if (isEquals(dataStatistics.value, dataStatisticsResponse)) {
                snackbar.value?.showMessage('Data is up to date');
            } else {
                snackbar.value?.showMessage('Data has been updated');
            }
        }

        dataStatistics.value = dataStatisticsResponse;
        loadingDataStatistics.value = false;
    }).catch(error => {
        loadingDataStatistics.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function exportData(fileType: string): void {
    if (exportingData.value) {
        return;
    }

    exportingData.value = true;

    userStore.getExportedUserData(fileType).then(data => {
        startDownloadFile(getExportFileName(fileType), data);
        exportingData.value = false;
    }).catch(error => {
        exportingData.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function clearAllTransactions(): void {
    if (!currentPasswordForClearData.value) {
        snackbar.value?.showMessage('Current password cannot be blank');
        return;
    }

    if (clearingData.value) {
        return;
    }

    confirmDialog.value?.open('Are you sure you want to clear all transactions?', { color: 'error' }).then(() => {
        clearingData.value = true;

        rootStore.clearAllUserTransactions({
            password: currentPasswordForClearData.value
        }).then(() => {
            clearingData.value = false;
            currentPasswordForClearData.value = '';

            snackbar.value?.showMessage('All transactions has been cleared');
            reloadUserDataStatistics(false);
        }).catch(error => {
            clearingData.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function clearAllData(): void {
    if (!currentPasswordForClearData.value) {
        snackbar.value?.showMessage('Current password cannot be blank');
        return;
    }

    if (clearingData.value) {
        return;
    }

    confirmDialog.value?.open('Are you sure you want to clear all data?', { color: 'error' }).then(() => {
        clearingData.value = true;

        rootStore.clearAllUserData({
            password: currentPasswordForClearData.value
        }).then(() => {
            clearingData.value = false;
            currentPasswordForClearData.value = '';

            snackbar.value?.showMessage('All user data has been cleared');
            reloadUserDataStatistics(false);
        }).catch(error => {
            clearingData.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function formatFileSize(bytes: string | number): string {
    const n = typeof bytes === 'string' ? parseInt(bytes, 10) : bytes;
    if (!n || n === 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(n) / Math.log(1024));
    return (n / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + ' ' + units[i];
}

function formatBackupDate(timestamp: string | number): string {
    const t = typeof timestamp === 'string' ? parseInt(timestamp, 10) : timestamp;
    if (!t) return '';
    return new Date(t * 1000).toLocaleString();
}

function reloadBackupList(): void {
    loadingBackups.value = true;

    userStore.getBackupList().then(list => {
        backupList.value = list;
        loadingBackups.value = false;
    }).catch(error => {
        loadingBackups.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function createNewBackup(): void {
    if (!backupPassword.value || creatingBackup.value || backupOperationInProgress.value) {
        return;
    }

    creatingBackup.value = true;
    backupOperationInProgress.value = true;

    userStore.createBackup({ password: backupPassword.value }).then(() => {
        creatingBackup.value = false;
        backupOperationInProgress.value = false;
        backupPassword.value = '';

        snackbar.value?.showMessage('Backup created successfully');
        reloadBackupList();
    }).catch(error => {
        creatingBackup.value = false;
        backupOperationInProgress.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function downloadExistingBackup(fileName: string): void {
    if (backupOperationInProgress.value) {
        return;
    }

    backupOperationInProgress.value = true;

    userStore.downloadBackup(fileName).then(data => {
        startDownloadFile(fileName, data);
        backupOperationInProgress.value = false;
    }).catch(error => {
        backupOperationInProgress.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function restoreFromBackup(fileName: string): void {
    if (!backupPassword.value || backupOperationInProgress.value) {
        return;
    }

    confirmDialog.value?.open('Are you sure you want to restore from this backup? Current data will be overwritten.', { color: 'warning' }).then(() => {
        backupOperationInProgress.value = true;

        userStore.restoreBackup({ fileName, password: backupPassword.value }).then(() => {
            backupOperationInProgress.value = false;
            backupPassword.value = '';

            snackbar.value?.showMessage('Backup restored successfully. Please restart the application.');
        }).catch(error => {
            backupOperationInProgress.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function deleteExistingBackup(fileName: string): void {
    if (backupOperationInProgress.value) {
        return;
    }

    confirmDialog.value?.open('Are you sure you want to delete this backup?', { color: 'error' }).then(() => {
        backupOperationInProgress.value = true;

        userStore.deleteBackup({ fileName }).then(() => {
            backupOperationInProgress.value = false;

            snackbar.value?.showMessage('Backup deleted successfully');
            reloadBackupList();
        }).catch(error => {
            backupOperationInProgress.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

reloadUserDataStatistics(false);

if (isBackupEnabled()) {
    reloadBackupList();
}
</script>
