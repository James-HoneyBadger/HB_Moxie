import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';

import type { TransactionInfoResponse, TransactionListByMaxTimeRequest } from '@/models/transaction.ts';
import type {
    SharedExpenseEntry,
    SharedExpenseSummaryItem,
    TransactionSplitUpdateRequest
} from '@/models/transaction_split.ts';

import { TransactionType } from '@/core/transaction.ts';
import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import services from '@/lib/services.ts';
import logger from '@/lib/logger.ts';

const EXPENSE_PICKER_MONTHS = 12;
const EXPENSE_PICKER_COUNT = 50;

export interface TransactionSplitFormRow {
    key: number;
    personName: string;
    amount: number;
}

export interface ShareableExpenseItem {
    readonly id: string;
    readonly title: string;
}

function emptyListRequest(minTime: number): TransactionListByMaxTimeRequest {
    return {
        maxTime: 0,
        minTime,
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
}

export function useSharedExpensesPageBase() {
    const { tt, formatDateTimeToLongDateTime, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();

    const accountsStore = useAccountsStore();
    const transactionCategoriesStore = useTransactionCategoriesStore();

    const loading = ref<boolean>(false);
    const errorMessage = ref<string>('');

    const summary = ref<SharedExpenseSummaryItem[]>([]);
    const selectedPerson = ref<string>('');
    const selectedCurrency = ref<string>('');
    const personEntries = ref<SharedExpenseEntry[]>([]);
    const personEntriesLoading = ref<boolean>(false);

    const editingExistingSplits = ref<boolean>(false);
    const editableTransactionId = ref<string>('');
    const editableTransactionAmount = ref<number>(0);
    const editableCurrency = ref<string>('');
    const formRows = ref<TransactionSplitFormRow[]>([]);
    const submitting = ref<boolean>(false);
    const submitErrorMessage = ref<string>('');

    const expenses = ref<TransactionInfoResponse[]>([]);
    const expensesLoading = ref<boolean>(false);

    const totalOwedByCurrency = computed<{ currency: string; amount: number }[]>(() => {
        const totals: Record<string, number> = {};

        for (const item of summary.value) {
            totals[item.currency] = (totals[item.currency] ?? 0) + item.total;
        }

        return Object.entries(totals)
            .map(([currency, amount]) => ({ currency, amount }))
            .sort((a, b) => b.amount - a.amount);
    });

    const hasSummary = computed<boolean>(() => summary.value.length > 0);

    const allocatedAmount = computed<number>(() =>
        formRows.value.reduce((total, row) => total + (Number.isFinite(row.amount) ? Math.abs(row.amount) : 0), 0));

    const remainingAmount = computed<number>(() => editableTransactionAmount.value - allocatedAmount.value);

    const canSubmit = computed<boolean>(() => {
        if (!editableTransactionId.value || submitting.value) {
            return false;
        }

        if (remainingAmount.value < 0) {
            return false;
        }

        return formRows.value.every(row => row.personName.trim().length > 0 && Math.abs(row.amount) > 0);
    });

    const expenseItems = computed<ShareableExpenseItem[]>(() =>
        expenses.value.map(expense => ({
            id: expense.id,
            title: `${formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(expense.time))} · ${expense.comment || categoryName(expense.categoryId)} · ${formatAmountToLocalizedNumeralsWithCurrency(expense.sourceAmount, expense.sourceAccount?.currency ?? accountCurrency(expense.sourceAccountId))}`
        })));

    function accountCurrency(accountId: string): string {
        return accountsStore.allAccountsMap[accountId]?.currency ?? '';
    }

    function categoryName(categoryId: string): string {
        return transactionCategoriesStore.allTransactionCategoriesMap[categoryId]?.name ?? '';
    }

    function formatAmount(amount: number, accountId: string): string {
        return formatAmountToLocalizedNumeralsWithCurrency(amount, accountCurrency(accountId));
    }

    function formatAmountInCurrency(amount: number, currency: string): string {
        return formatAmountToLocalizedNumeralsWithCurrency(amount, currency);
    }

    function formatTime(unixTime: number): string {
        return formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(unixTime));
    }

    function isPersonSelected(item: SharedExpenseSummaryItem): boolean {
        return selectedPerson.value === item.personName && selectedCurrency.value === item.currency;
    }

    function selectPerson(item: SharedExpenseSummaryItem): void {
        if (isPersonSelected(item)) {
            selectedPerson.value = '';
            selectedCurrency.value = '';
            personEntries.value = [];
            return;
        }

        selectedPerson.value = item.personName;
        selectedCurrency.value = item.currency;
        loadPersonEntries();
    }

    function loadPersonEntries(): void {
        if (!selectedPerson.value) {
            personEntries.value = [];
            return;
        }

        personEntriesLoading.value = true;

        services.getSharedExpensesByPerson({ personName: selectedPerson.value }).then(response => {
            personEntries.value = response.data.result ?? [];
            personEntriesLoading.value = false;
        }).catch(error => {
            logger.error('failed to load shared expenses of person', error);
            personEntries.value = [];
            personEntriesLoading.value = false;
        });
    }

    function loadSummary(): void {
        loading.value = true;
        errorMessage.value = '';

        services.getSharedExpenseSummary().then(response => {
            summary.value = response.data.result ?? [];
            loading.value = false;

            if (selectedPerson.value) {
                loadPersonEntries();
            }
        }).catch(error => {
            logger.error('failed to load shared expense summary', error);
            errorMessage.value = tt('Failed to load data');
            loading.value = false;
        });
    }

    function loadExpenses(): void {
        if (expenses.value.length > 0 || expensesLoading.value) {
            return;
        }

        expensesLoading.value = true;

        const now = Math.floor(Date.now() / 1000);
        const minTime = (now - EXPENSE_PICKER_MONTHS * 30 * 24 * 3600) * 1000;

        services.getTransactions(emptyListRequest(minTime)).then(response => {
            expenses.value = response.data.result?.items ?? [];
            expensesLoading.value = false;
        }).catch(error => {
            logger.error('failed to load expense transactions', error);
            expenses.value = [];
            expensesLoading.value = false;
        });
    }

    function prepareSplitEditor(): void {
        loadSummary();
        loadExpenses();
    }

    function startNewSplit(): void {
        editingExistingSplits.value = false;
        editableTransactionId.value = '';
        editableTransactionAmount.value = 0;
        editableCurrency.value = '';
        submitErrorMessage.value = '';
        formRows.value = [{ key: 1, personName: '', amount: 0 }];
        prepareSplitEditor();
    }

    function resetSplitEditor(): void {
        editingExistingSplits.value = false;
        editableTransactionId.value = '';
        editableTransactionAmount.value = 0;
        editableCurrency.value = '';
        submitErrorMessage.value = '';
        formRows.value = [{ key: 1, personName: '', amount: 0 }];
    }

    function startEditSplit(entry: SharedExpenseEntry): void {
        editingExistingSplits.value = true;
        editableTransactionId.value = entry.transactionId;
        editableTransactionAmount.value = Math.abs(entry.transactionAmount);
        editableCurrency.value = entry.currency;
        submitErrorMessage.value = '';
        formRows.value = [{ key: 1, personName: selectedPerson.value, amount: entry.splitAmount }];

        services.getTransactionSplits({ transactionId: entry.transactionId }).then(response => {
            const splits = response.data.result?.splits ?? [];

            if (splits.length > 0) {
                formRows.value = splits.map((split, index) => ({
                    key: index + 1,
                    personName: split.personName,
                    amount: split.amount
                }));
            }
        }).catch(error => {
            logger.error('failed to load transaction splits', error);
        });

        prepareSplitEditor();
    }

    function selectTransaction(transactionId: string): void {
        const expense = expenses.value.find(item => item.id === transactionId);

        if (!expense) {
            return;
        }

        editableTransactionId.value = expense.id;
        editableTransactionAmount.value = Math.abs(expense.sourceAmount);
        editableCurrency.value = accountCurrency(expense.sourceAccountId);

        const firstRow = formRows.value[0];

        if (formRows.value.length === 1 && firstRow && !firstRow.personName) {
            formRows.value = [{ key: 1, personName: selectedPerson.value, amount: 0 }];
        }
    }

    function addFormRow(): void {
        formRows.value.push({ key: Date.now() + formRows.value.length, personName: '', amount: 0 });
    }

    function removeFormRow(key: number): void {
        formRows.value = formRows.value.filter(row => row.key !== key);
    }

    function submit(): Promise<boolean> {
        if (!canSubmit.value) {
            return Promise.resolve(false);
        }

        const request: TransactionSplitUpdateRequest = {
            transactionId: editableTransactionId.value,
            splits: formRows.value.map(row => ({
                personName: row.personName.trim(),
                amount: Math.abs(row.amount)
            }))
        };

        submitting.value = true;
        submitErrorMessage.value = '';

        return services.updateTransactionSplits(request).then(() => {
            submitting.value = false;
            editableTransactionId.value = '';
            formRows.value = [{ key: 1, personName: selectedPerson.value, amount: 0 }];
            loadSummary();
            return true;
        }).catch(error => {
            logger.error('failed to update transaction splits', error);
            submitErrorMessage.value = tt('Failed to save data');
            submitting.value = false;
            return false;
        });
    }

    function clearSplit(entry: SharedExpenseEntry): void {
        submitting.value = true;
        submitErrorMessage.value = '';

        services.updateTransactionSplits({ transactionId: entry.transactionId, splits: [] }).then(() => {
            submitting.value = false;
            loadSummary();
        }).catch(error => {
            logger.error('failed to clear transaction splits', error);
            submitErrorMessage.value = tt('Failed to save data');
            submitting.value = false;
        });
    }

    function init(): void {
        accountsStore.loadAllAccounts({ force: false }).finally(() => loadSummary());
        transactionCategoriesStore.loadAllCategories({ force: false }).finally(() => loadSummary());
        loadExpenses();
    }

    return {
        // constants
        maxSplitCount: 20,
        // states
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
        expenses,
        expensesLoading,
        // computed states
        totalOwedByCurrency,
        hasSummary,
        allocatedAmount,
        remainingAmount,
        canSubmit,
        expenseItems,
        // functions
        accountCurrency,
        categoryName,
        formatAmount,
        formatAmountInCurrency,
        formatTime,
        isPersonSelected,
        selectPerson,
        loadSummary,
        loadExpenses,
        prepareSplitEditor,
        resetSplitEditor,
        startNewSplit,
        startEditSplit,
        selectTransaction,
        addFormRow,
        removeFormRow,
        submit,
        clearSplit,
        init
    };
}
