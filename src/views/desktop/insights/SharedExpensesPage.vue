<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="d-flex align-center flex-wrap">
                        <span>{{ tt('Shared Expenses') }}</span>
                        <v-spacer />
                        <v-btn class="me-2" color="primary" :disabled="loading" @click="openNewSplit">
                            <v-icon :icon="mdiAccountPlusOutline" start />
                            {{ tt('Split an Expense') }}
                        </v-btn>
                        <v-btn color="default" variant="text" :disabled="loading" @click="loadSummary">
                            <v-icon :icon="mdiRefresh" start />
                            {{ tt('Refresh') }}
                        </v-btn>
                    </div>
                </template>

                <v-progress-linear v-if="loading" indeterminate />

                <v-card-text>
                    <div class="text-body-2 mb-2">
                        {{ tt('Track how much other people owe you by splitting expense transactions with them.') }}
                    </div>

                    <v-alert v-if="errorMessage" type="error" variant="tonal" class="mb-4">{{ errorMessage }}</v-alert>

                    <v-row v-if="!loading && hasSummary">
                        <v-col cols="12" md="6" :key="item.currency" v-for="item in totalOwedByCurrency">
                            <div class="shared-expense-stat">
                                <div class="text-h4">{{ formatAmountInCurrency(item.amount, item.currency) }}</div>
                                <div class="text-body-2">{{ tt('Total Owed') }} ({{ item.currency }})</div>
                            </div>
                        </v-col>
                    </v-row>

                    <div v-if="!loading && !hasSummary" class="text-body-2 text-medium-emphasis">
                        {{ tt('No shared expenses recorded yet') }}
                    </div>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="d-flex align-center">
                        <v-icon :icon="mdiAccountGroupOutline" start />
                        <span>{{ tt('People') }}</span>
                    </div>
                </template>
                <v-card-text>
                    <v-table v-if="hasSummary" density="compact">
                        <thead>
                        <tr>
                            <th>{{ tt('Name') }}</th>
                            <th>{{ tt('Currency') }}</th>
                            <th class="text-end">{{ tt('Amount Owed') }}</th>
                            <th class="text-end">{{ tt('Entries') }}</th>
                            <th>{{ tt('Last Expense') }}</th>
                            <th></th>
                        </tr>
                        </thead>
                        <tbody>
                        <tr class="cursor-pointer"
                            :key="item.personName + item.currency"
                            v-for="item in summary"
                            @click="selectPerson(item)">
                            <td>{{ item.personName }}</td>
                            <td>{{ item.currency }}</td>
                            <td class="text-end">{{ formatAmountInCurrency(item.total, item.currency) }}</td>
                            <td class="text-end">{{ item.entryCount }}</td>
                            <td>{{ formatTime(item.lastTime) }}</td>
                            <td class="text-end">
                                <v-icon :icon="isPersonSelected(item) ? mdiChevronUp : mdiChevronDown" size="20" />
                            </td>
                        </tr>
                        </tbody>
                    </v-table>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" v-if="selectedPerson">
            <v-card>
                <template #title>
                    <div class="d-flex align-center flex-wrap">
                        <v-icon :icon="mdiFormatListBulleted" start />
                        <span>{{ selectedPerson }}</span>
                        <v-spacer />
                        <v-chip size="small" variant="tonal">{{ selectedCurrency }}</v-chip>
                    </div>
                </template>

                <v-progress-linear v-if="personEntriesLoading" indeterminate />

                <v-card-text>
                    <div v-if="!personEntriesLoading && personEntries.length < 1" class="text-body-2 text-medium-emphasis">
                        {{ tt('No shared expenses recorded yet') }}
                    </div>

                    <v-table v-if="personEntries.length > 0" density="compact">
                        <thead>
                        <tr>
                            <th>{{ tt('Date') }}</th>
                            <th>{{ tt('Category') }}</th>
                            <th class="text-end">{{ tt('Transaction Amount') }}</th>
                            <th class="text-end">{{ tt('Amount Owed') }}</th>
                            <th>{{ tt('Description') }}</th>
                            <th></th>
                        </tr>
                        </thead>
                        <tbody>
                        <tr :key="entry.transactionId" v-for="entry in personEntries">
                            <td>{{ formatTime(entry.time) }}</td>
                            <td>{{ categoryName(entry.categoryId) }}</td>
                            <td class="text-end">{{ formatAmountInCurrency(entry.transactionAmount, entry.currency) }}</td>
                            <td class="text-end">{{ formatAmountInCurrency(entry.splitAmount, entry.currency) }}</td>
                            <td class="text-truncate">{{ entry.comment }}</td>
                            <td class="text-end text-no-wrap">
                                <v-btn size="small" variant="text" :disabled="clearing" @click="openEditSplit(entry)">
                                    {{ tt('Edit') }}
                                </v-btn>
                                <v-btn size="small" variant="text" color="error" :disabled="clearing" @click="clearSplit(entry)">
                                    {{ tt('Clear') }}
                                </v-btn>
                            </td>
                        </tr>
                        </tbody>
                    </v-table>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>

    <split-expense-dialog ref="splitDialog" @saved="loadSummary" />
</template>

<script setup lang="ts">
import { useTemplateRef } from 'vue';

import {
    mdiAccountGroupOutline,
    mdiAccountPlusOutline,
    mdiChevronDown,
    mdiChevronUp,
    mdiFormatListBulleted,
    mdiRefresh
} from '@mdi/js';

import SplitExpenseDialog from './dialogs/SplitExpenseDialog.vue';

import type { SharedExpenseEntry } from '@/models/transaction_split.ts';

import { useI18n } from '@/locales/helpers.ts';
import { useSharedExpensesPageBase } from '@/views/base/insights/SharedExpensesPageBase.ts';

const { tt } = useI18n();
const {
    loading,
    errorMessage,
    summary,
    selectedPerson,
    selectedCurrency,
    personEntries,
    personEntriesLoading,
    clearing,
    totalOwedByCurrency,
    hasSummary,
    categoryName,
    formatAmountInCurrency,
    formatTime,
    isPersonSelected,
    selectPerson,
    loadSummary,
    clearSplit,
    init
} = useSharedExpensesPageBase();

const splitDialog = useTemplateRef<InstanceType<typeof SplitExpenseDialog>>('splitDialog');

function openNewSplit(): void {
    splitDialog.value?.open({ defaultPersonName: selectedPerson.value });
}

function openEditSplit(entry: SharedExpenseEntry): void {
    splitDialog.value?.open({
        transactionId: entry.transactionId,
        transactionAmount: entry.transactionAmount,
        currency: entry.currency,
        defaultPersonName: selectedPerson.value
    });
}

init();
</script>

<style scoped>
.shared-expense-stat {
    padding: 8px 0;
}
</style>
