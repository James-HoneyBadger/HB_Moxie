<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="d-flex align-center flex-wrap">
                        <span>{{ tt('Shared Expenses') }}</span>
                        <v-spacer />
                        <v-btn class="me-2" color="primary" :disabled="loading || submitting" @click="openNewSplit">
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
                    <v-alert v-if="submitErrorMessage" type="error" variant="tonal" class="mb-4">{{ submitErrorMessage }}</v-alert>

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
                                <v-btn size="small" variant="text" :disabled="submitting" @click="openEditSplit(entry)">
                                    {{ tt('Edit') }}
                                </v-btn>
                                <v-btn size="small" variant="text" color="error" :disabled="submitting" @click="clearSplit(entry)">
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

    <v-dialog width="640" :persistent="submitting" v-model="showSplitDialog" @after-leave="onDialogClosed">
        <v-card>
            <template #title>
                <div class="d-flex align-center">
                    <v-icon :icon="mdiAccountMultipleCheckOutline" start />
                    <span>{{ editingExistingSplits ? tt('Edit Split') : tt('Split an Expense') }}</span>
                </div>
            </template>

            <v-card-text>
                <v-alert v-if="submitErrorMessage" type="error" variant="tonal" class="mb-4">{{ submitErrorMessage }}</v-alert>

                <v-autocomplete v-if="!editableTransactionId"
                                class="mb-2"
                                density="compact" variant="outlined"
                                :items="expenseItems" item-title="title" item-value="id"
                                :loading="expensesLoading"
                                :label="tt('Transaction')"
                                :placeholder="tt('Select an expense transaction')"
                                :no-data-text="tt('No data')"
                                v-model="selectedTransactionId" />

                <div class="d-flex align-center flex-wrap mb-3" v-if="editableTransactionAmount > 0">
                    <span class="text-body-2">{{ tt('Transaction Amount') }}:</span>
                    <span class="text-body-2 font-weight-bold ms-1">{{ formatAmountInCurrency(editableTransactionAmount, editableCurrency) }}</span>
                    <v-spacer />
                    <span class="text-body-2" :class="{ 'text-error': remainingAmount < 0 }">
                        {{ tt('Unallocated') }}: {{ formatAmountInCurrency(remainingAmount, editableCurrency) }}
                    </span>
                </div>

                <v-row class="align-center" :key="row.key" v-for="row in formRows">
                    <v-col cols="12" md="5">
                        <v-text-field density="compact" variant="outlined"
                                      :label="tt('Person Name')"
                                      :disabled="submitting"
                                      v-model="row.personName" />
                    </v-col>
                    <v-col cols="12" md="5">
                        <amount-input :currency="editableCurrency"
                                      :show-currency="!!editableCurrency"
                                      :label="tt('Amount Owed')"
                                      :placeholder="tt('Amount')"
                                      :disabled="submitting"
                                      v-model="row.amount" />
                    </v-col>
                    <v-col cols="12" md="2" class="text-end">
                        <v-btn density="comfortable" icon variant="text" color="error"
                               :disabled="submitting || formRows.length <= 1"
                               @click="removeFormRow(row.key)">
                            <v-icon :icon="mdiTrashCanOutline" />
                        </v-btn>
                    </v-col>
                </v-row>

                <v-btn class="mt-2" variant="text" size="small"
                       :disabled="submitting || formRows.length >= maxSplitCount"
                       @click="addFormRow">
                    <v-icon :icon="mdiPlus" start />
                    {{ tt('Add Person') }}
                </v-btn>
            </v-card-text>

            <v-card-actions>
                <v-spacer />
                <v-btn color="default" variant="text" :disabled="submitting" @click="showSplitDialog = false">
                    {{ tt('Cancel') }}
                </v-btn>
                <v-btn color="primary" :loading="submitting" :disabled="!canSubmit" @click="submitAndClose">
                    {{ tt('Save') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';

import {
    mdiAccountGroupOutline,
    mdiAccountMultipleCheckOutline,
    mdiAccountPlusOutline,
    mdiChevronDown,
    mdiChevronUp,
    mdiFormatListBulleted,
    mdiPlus,
    mdiRefresh,
    mdiTrashCanOutline
} from '@mdi/js';

import type { SharedExpenseEntry } from '@/models/transaction_split.ts';

import { useI18n } from '@/locales/helpers.ts';
import { useSharedExpensesPageBase } from '@/views/base/insights/SharedExpensesPageBase.ts';

const { tt } = useI18n();
const {
    maxSplitCount,
    loading,
    errorMessage,
    summary,
    selectedPerson,
    selectedCurrency,
    personEntries,
    personEntriesLoading,
    editingExistingSplits,
    editableTransactionId,
    editableTransactionAmount,
    editableCurrency,
    formRows,
    submitting,
    submitErrorMessage,
    expensesLoading,
    totalOwedByCurrency,
    hasSummary,
    remainingAmount,
    canSubmit,
    expenseItems,
    categoryName,
    formatAmountInCurrency,
    formatTime,
    isPersonSelected,
    selectPerson,
    loadSummary,
    resetSplitEditor,
    startNewSplit,
    startEditSplit,
    selectTransaction,
    addFormRow,
    removeFormRow,
    submit,
    clearSplit,
    init
} = useSharedExpensesPageBase();

const showSplitDialog = ref<boolean>(false);
const selectedTransactionId = ref<string>('');

watch(selectedTransactionId, (transactionId: string) => {
    if (transactionId) {
        selectTransaction(transactionId);
    }
});

function openNewSplit(): void {
    selectedTransactionId.value = '';
    startNewSplit();
    showSplitDialog.value = true;
}

function openEditSplit(entry: SharedExpenseEntry): void {
    selectedTransactionId.value = '';
    startEditSplit(entry);
    showSplitDialog.value = true;
}

function onDialogClosed(): void {
    selectedTransactionId.value = '';
    resetSplitEditor();
}

async function submitAndClose(): Promise<void> {
    if (await submit()) {
        showSplitDialog.value = false;
    }
}

init();
</script>

<style scoped>
.shared-expense-stat {
    padding: 8px 0;
}
</style>
