import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';

import type {
    SharedExpenseEntry,
    SharedExpenseSummaryItem,
    SharedExpenseSettlementRequest,
    TransactionSplitRenamePersonRequest,
    TransactionSplitUpdateRequest
} from '@/models/transaction_split.ts';

import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import services from '@/lib/services.ts';
import logger from '@/lib/logger.ts';

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
    const clearing = ref<boolean>(false);
    const settling = ref<boolean>(false);

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

    function categoryName(categoryId: string): string {
        return transactionCategoriesStore.allTransactionCategoriesMap[categoryId]?.name ?? '';
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

    function settleSharedExpense(request: SharedExpenseSettlementRequest): Promise<boolean> {
        settling.value = true;
        errorMessage.value = '';

        return services.settleSharedExpense(request).then(() => {
            settling.value = false;
            loadSummary();
            return true;
        }).catch(error => {
            logger.error('failed to record settlement', error);
            errorMessage.value = tt('Failed to save data');
            settling.value = false;
            return false;
        });
    }

    function deleteSettlement(settlementId: string): Promise<boolean> {
        settling.value = true;
        errorMessage.value = '';

        return services.deleteSharedExpenseSettlement({ settlementId }).then(() => {
            settling.value = false;
            loadSummary();
            return true;
        }).catch(error => {
            logger.error('failed to delete settlement', error);
            errorMessage.value = tt('Failed to save data');
            settling.value = false;
            return false;
        });
    }

    function renamePerson(oldName: string, newName: string): Promise<boolean> {
        const request: TransactionSplitRenamePersonRequest = { oldName, newName };

        return services.renameSharedExpensePerson(request).then(() => {
            if (selectedPerson.value === oldName) {
                selectedPerson.value = newName;
            }

            loadSummary();
            return true;
        }).catch(error => {
            logger.error('failed to rename person', error);
            errorMessage.value = tt('Failed to save data');
            return false;
        });
    }

    function clearSplit(entry: SharedExpenseEntry): void {
        const request: TransactionSplitUpdateRequest = { transactionId: entry.transactionId, splits: [] };

        clearing.value = true;
        errorMessage.value = '';

        services.updateTransactionSplits(request).then(() => {
            clearing.value = false;
            loadSummary();
        }).catch(error => {
            logger.error('failed to clear transaction splits', error);
            errorMessage.value = tt('Failed to save data');
            clearing.value = false;
        });
    }

    function init(): void {
        accountsStore.loadAllAccounts({ force: false }).finally(() => loadSummary());
        transactionCategoriesStore.loadAllCategories({ force: false }).finally(() => loadSummary());
    }

    return {
        // states
        loading,
        errorMessage,
        summary,
        selectedPerson,
        selectedCurrency,
        personEntries,
        personEntriesLoading,
        clearing,
        settling,
        // computed states
        totalOwedByCurrency,
        hasSummary,
        // functions
        categoryName,
        formatAmountInCurrency,
        formatTime,
        isPersonSelected,
        selectPerson,
        loadSummary,
        loadPersonEntries,
        settleSharedExpense,
        deleteSettlement,
        renamePerson,
        clearSplit,
        init
    };
}
