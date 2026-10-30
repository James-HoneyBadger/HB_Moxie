<template>
    <v-dialog width="640" :persistent="submitting" v-model="showState">
        <v-card>
            <template #title>
                <div class="d-flex align-center">
                    <v-icon :icon="mdiAccountMultipleCheckOutline" start />
                    <span>{{ editingExisting ? tt('Edit Split') : tt('Split an Expense') }}</span>
                </div>
            </template>

            <v-card-text>
                <v-alert v-if="errorMessage" type="error" variant="tonal" class="mb-4">{{ errorMessage }}</v-alert>

                <v-autocomplete v-if="!transactionId"
                                class="mb-2"
                                density="compact" variant="outlined"
                                :items="expenseItems" item-title="title" item-value="id"
                                :loading="expensesLoading"
                                :label="tt('Transaction')"
                                :placeholder="tt('Select an expense transaction')"
                                :no-data-text="tt('No data')"
                                v-model="selectedTransactionId" />

                <v-btn-toggle class="mb-3" density="compact" variant="outlined"
                              :disabled="submitting || !transactionId"
                              v-model="direction">
                    <v-btn :value="SplitDirection.TheyOweMe" size="small">{{ tt('They Owe Me') }}</v-btn>
                    <v-btn :value="SplitDirection.IOweThem" size="small">{{ tt('I Owe Them') }}</v-btn>
                </v-btn-toggle>

                <div class="d-flex align-center flex-wrap mb-3" v-if="transactionAmount > 0">
                    <span class="text-body-2">{{ tt('Transaction Amount') }}:</span>
                    <span class="text-body-2 font-weight-bold ms-1">{{ formatAmountInCurrency(transactionAmount, currency) }}</span>
                    <v-spacer />
                    <span class="text-body-2" :class="{ 'text-error': remainingAmount < 0 }">
                        {{ tt('Unallocated') }}: {{ formatAmountInCurrency(remainingAmount, currency) }}
                    </span>
                </div>

                <v-row class="align-center" :key="row.key" v-for="row in rows">
                    <v-col cols="12" md="5">
                        <v-combobox density="compact" variant="outlined"
                                    :items="knownPeople" :label="tt('Person Name')"
                                    :no-data-text="tt('No data')"
                                    :disabled="submitting"
                                    v-model="row.personName" />
                    </v-col>
                    <v-col cols="12" md="5">
                        <amount-input :currency="currency"
                                      :show-currency="!!currency"
                                      :label="tt('Amount Owed')"
                                      :placeholder="tt('Amount')"
                                      :disabled="submitting"
                                      v-model="row.amount" />
                    </v-col>
                    <v-col cols="12" md="2" class="text-end">
                        <v-btn density="comfortable" icon variant="text" color="error"
                               :disabled="submitting || rows.length <= 1"
                               @click="removeRow(row.key)">
                            <v-icon :icon="mdiTrashCanOutline" />
                        </v-btn>
                    </v-col>
                </v-row>

                <div class="d-flex flex-wrap align-center">
                    <v-btn class="mt-2" variant="text" size="small"
                           :disabled="submitting || rows.length >= maxSplitCount"
                           @click="addRow">
                        <v-icon :icon="mdiPlus" start />
                        {{ tt('Add Person') }}
                    </v-btn>
                    <v-btn class="mt-2 ms-2" variant="text" size="small"
                           :disabled="submitting || !transactionId || transactionAmount <= 0 || rows.length < 1"
                           @click="splitEqually">
                        <v-icon :icon="mdiScaleBalance" start />
                        {{ tt('Split Equally') }}
                    </v-btn>
                </div>
            </v-card-text>

            <v-card-actions>
                <v-spacer />
                <v-btn color="default" variant="text" :disabled="submitting" @click="close">
                    {{ tt('Cancel') }}
                </v-btn>
                <v-btn color="primary" :loading="submitting" :disabled="!canSubmit" @click="submit">
                    {{ tt('Save') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';

import {
    mdiAccountMultipleCheckOutline,
    mdiPlus,
    mdiScaleBalance,
    mdiTrashCanOutline
} from '@mdi/js';

import type { TransactionInfoResponse, TransactionListByMaxTimeRequest } from '@/models/transaction.ts';
import { SplitDirection } from '@/models/transaction_split.ts';
import type { TransactionSplitUpdateRequest } from '@/models/transaction_split.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';

import { TransactionType } from '@/core/transaction.ts';
import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import { useI18n } from '@/locales/helpers.ts';
import services from '@/lib/services.ts';
import logger from '@/lib/logger.ts';

const maxSplitCount = 20;
const EXPENSE_PICKER_MONTHS = 12;
const EXPENSE_PICKER_COUNT = 50;

interface SplitFormRow {
    key: number;
    personName: string;
    amount: number;
}

interface ExpensePickerItem {
    readonly id: string;
    readonly title: string;
}

export interface SplitExpenseDialogOptions {
    transactionId?: string;
    transactionAmount?: number;
    currency?: string;
    defaultPersonName?: string;
    direction?: number;
}

const emit = defineEmits<{
    (e: 'saved'): void;
}>();

const { tt, formatDateTimeToLongDateTime, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();

const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();

const showState = ref<boolean>(false);
const submitting = ref<boolean>(false);
const errorMessage = ref<string>('');

const editingExisting = ref<boolean>(false);
const transactionId = ref<string>('');
const transactionAmount = ref<number>(0);
const currency = ref<string>('');
const defaultPersonName = ref<string>('');
const rows = ref<SplitFormRow[]>([]);

const selectedTransactionId = ref<string>('');
const expenses = ref<TransactionInfoResponse[]>([]);
const expensesLoading = ref<boolean>(false);

const direction = ref<number>(SplitDirection.TheyOweMe);
const knownPeople = ref<string[]>([]);

const allocatedAmount = computed<number>(() =>
    rows.value.reduce((total, row) => total + (Number.isFinite(row.amount) ? Math.abs(row.amount) : 0), 0));

const remainingAmount = computed<number>(() => transactionAmount.value - allocatedAmount.value);

const canSubmit = computed<boolean>(() => {
    if (!transactionId.value || submitting.value || remainingAmount.value < 0) {
        return false;
    }

    return rows.value.every(row => row.personName.trim().length > 0 && Math.abs(row.amount) > 0);
});

const expenseItems = computed<ExpensePickerItem[]>(() =>
    expenses.value.map(expense => ({
        id: expense.id,
        title: `${formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(expense.time))} · ${expense.comment || categoryName(expense.categoryId)} · ${formatAmountInCurrency(expense.sourceAmount, expense.sourceAccount?.currency ?? accountCurrency(expense.sourceAccountId))}`
    })));

function accountCurrency(accountId: string): string {
    return accountsStore.allAccountsMap[accountId]?.currency ?? '';
}

function categoryName(categoryId: string): string {
    return transactionCategoriesStore.allTransactionCategoriesMap[categoryId]?.name ?? '';
}

function formatAmountInCurrency(amount: number, currencyCode: string): string {
    return formatAmountToLocalizedNumeralsWithCurrency(amount, currencyCode);
}

function resetRows(personName: string, amount: number): void {
    rows.value = [{ key: 1, personName, amount }];
}

function loadExpenses(): void {
    if (expenses.value.length > 0 || expensesLoading.value) {
        return;
    }

    expensesLoading.value = true;

    const now = Math.floor(Date.now() / 1000);
    const request: TransactionListByMaxTimeRequest = {
        maxTime: 0,
        minTime: (now - EXPENSE_PICKER_MONTHS * 30 * 24 * 3600) * 1000,
        count: EXPENSE_PICKER_COUNT,
        page: 1,
        withCount: false,
        type: TransactionType.Expense,
        categoryIds: '',
        accountIds: '',
        tagFilter: '',
        amountFilter: '',
        keyword: ''
    };

    services.getTransactions(request).then(response => {
        expenses.value = response.data.result?.items ?? [];
        expensesLoading.value = false;
    }).catch(error => {
        logger.error('failed to load expense transactions', error);
        expenses.value = [];
        expensesLoading.value = false;
    });
}

function loadKnownPeople(): void {
    if (knownPeople.value.length > 0) {
        return;
    }

    services.getKnownPeople().then(response => {
        knownPeople.value = response.data.result ?? [];
    }).catch(error => {
        logger.error('failed to load known people', error);
    });
}

function splitEqually(): void {
    if (rows.value.length < 1 || transactionAmount.value <= 0) {
        return;
    }

    const share = Math.floor(transactionAmount.value / rows.value.length);
    const remainder = transactionAmount.value - share * rows.value.length;

    rows.value = rows.value.map((row, index) => ({
        ...row,
        amount: index === 0 ? share + remainder : share
    }));
}

function applySelectedTransaction(id: string): void {
    const expense = expenses.value.find(item => item.id === id);

    if (!expense) {
        return;
    }

    transactionId.value = expense.id;
    transactionAmount.value = Math.abs(expense.sourceAmount);
    currency.value = expense.sourceAccount?.currency ?? accountCurrency(expense.sourceAccountId);

    const firstRow = rows.value[0];

    if (rows.value.length === 1 && firstRow && !firstRow.personName) {
        rows.value = [{ key: 1, personName: defaultPersonName.value, amount: 0 }];
    }
}

watch(selectedTransactionId, (id: string) => {
    if (id) {
        applySelectedTransaction(id);
    }
});

function open(options?: SplitExpenseDialogOptions): void {
    const opts = options ?? {};

    editingExisting.value = !!opts.transactionId;
    errorMessage.value = '';
    submitting.value = false;
    selectedTransactionId.value = '';
    defaultPersonName.value = opts.defaultPersonName ?? '';
    direction.value = opts.direction ?? SplitDirection.TheyOweMe;

    loadKnownPeople();

    if (opts.transactionId) {
        transactionId.value = opts.transactionId;
        transactionAmount.value = Math.abs(opts.transactionAmount ?? 0);
        currency.value = opts.currency ?? '';
        resetRows(opts.defaultPersonName ?? '', 0);

        services.getTransactionSplits({ transactionId: opts.transactionId }).then(response => {
            const splits = response.data.result?.splits ?? [];

            transactionAmount.value = Math.abs(response.data.result?.transactionAmount ?? transactionAmount.value);
            currency.value = response.data.result?.currency ?? currency.value;
            direction.value = response.data.result?.direction ?? direction.value;

            // A transaction without shares is a new split, even when it was opened
            // from a specific transaction.
            editingExisting.value = splits.length > 0;

            if (editingExisting.value) {
                rows.value = splits.map((split, index) => ({
                    key: index + 1,
                    personName: split.personName,
                    amount: split.amount
                }));
            }
        }).catch(error => {
            logger.error('failed to load transaction splits', error);
            errorMessage.value = tt('Failed to load data');
        });
    } else {
        transactionId.value = '';
        transactionAmount.value = 0;
        currency.value = '';
        resetRows(defaultPersonName.value, 0);
        loadExpenses();
    }

    showState.value = true;
}

function addRow(): void {
    rows.value.push({ key: Date.now() + rows.value.length, personName: '', amount: 0 });
}

function removeRow(key: number): void {
    rows.value = rows.value.filter(row => row.key !== key);
}

function close(): void {
    showState.value = false;
    transactionId.value = '';
    transactionAmount.value = 0;
    currency.value = '';
    selectedTransactionId.value = '';
    errorMessage.value = '';
    resetRows('', 0);
}

function submit(): void {
    if (!canSubmit.value) {
        return;
    }

    const request: TransactionSplitUpdateRequest = {
        transactionId: transactionId.value,
        direction: direction.value,
        splits: rows.value.map(row => ({
            personName: row.personName.trim(),
            amount: Math.abs(row.amount)
        }))
    };

    submitting.value = true;
    errorMessage.value = '';

    services.updateTransactionSplits(request).then(() => {
        submitting.value = false;
        close();
        emit('saved');
    }).catch(error => {
        logger.error('failed to update transaction splits', error);
        errorMessage.value = tt('Failed to save data');
        submitting.value = false;
    });
}

defineExpose({
    open
});
</script>
