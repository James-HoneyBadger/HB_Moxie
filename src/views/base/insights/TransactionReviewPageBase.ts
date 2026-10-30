import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';

import { TransactionReviewDismissalKind } from '@/models/transaction_review.ts';
import type {
    TransactionReviewDismissRequest,
    TransactionReviewDuplicateGroup,
    TransactionReviewResponse,
    TransactionReviewSubscription,
    TransactionReviewUnusual
} from '@/models/transaction_review.ts';

import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import services from '@/lib/services.ts';
import logger from '@/lib/logger.ts';

const REVIEW_PERIOD_OPTIONS = [3, 6, 12, 24];

// Occurrences of each cadence in a year, used to estimate the monthly cost.
const CADENCE_PER_YEAR: Record<string, number> = {
    weekly: 52,
    monthly: 12,
    quarterly: 4,
    yearly: 1
};

export function useTransactionReviewPageBase() {
    const { tt, formatDateTimeToLongDateTime, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();

    const accountsStore = useAccountsStore();
    const transactionCategoriesStore = useTransactionCategoriesStore();

    const loading = ref<boolean>(false);
    const dismissing = ref<boolean>(false);
    const months = ref<number>(12);
    const review = ref<TransactionReviewResponse | null>(null);
    const errorMessage = ref<string>('');

    const duplicateGroups = computed<TransactionReviewDuplicateGroup[]>(() => review.value?.duplicates ?? []);
    const unusualItems = computed<TransactionReviewUnusual[]>(() => review.value?.unusual ?? []);
    const subscriptions = computed<TransactionReviewSubscription[]>(() => review.value?.subscriptions ?? []);

    const hasFindings = computed<boolean>(() =>
        duplicateGroups.value.length > 0 || unusualItems.value.length > 0 || subscriptions.value.length > 0);

    const periodOptions = computed<{ title: string; value: number }[]>(() =>
        REVIEW_PERIOD_OPTIONS.map(option => ({ title: `${option} ${tt('Months')}`, value: option })));

    // Subscriptions can be billed from accounts in different currencies, so the
    // estimate is reported per currency instead of being summed blindly.
    const estimatedMonthlyCosts = computed<{ currency: string; amount: number }[]>(() => {
        const totals: Record<string, number> = {};

        for (const subscription of subscriptions.value) {
            const perYear = CADENCE_PER_YEAR[subscription.cadence];

            if (!perYear) {
                continue;
            }

            const currency = accountCurrency(subscription.accountId);
            totals[currency] = (totals[currency] ?? 0) + (subscription.averageAmount * perYear) / 12;
        }

        return Object.entries(totals)
            .map(([currency, amount]) => ({ currency, amount: Math.round(amount) }))
            .sort((a, b) => b.amount - a.amount);
    });

    const reviewPeriod = computed<string>(() => {
        if (!review.value) {
            return '';
        }

        const start = formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(review.value.startTime));
        const end = formatDateTimeToLongDateTime(parseDateTimeFromUnixTime(review.value.endTime));

        return `${start} - ${end}`;
    });

    function accountName(accountId: string): string {
        return accountsStore.allAccountsMap[accountId]?.name ?? '';
    }

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

    function cadenceName(cadence: string): string {
        switch (cadence) {
            case 'weekly':
                return tt('Weekly');
            case 'monthly':
                return tt('Monthly');
            case 'quarterly':
                return tt('Quarterly');
            case 'yearly':
                return tt('Yearly');
            default:
                return cadence;
        }
    }

    function dismiss(kind: TransactionReviewDismissalKind, key: string): Promise<boolean> {
        const request: TransactionReviewDismissRequest = { kind, key };

        dismissing.value = true;

        return services.dismissTransactionReviewFinding(request).then(() => {
            dismissing.value = false;
            reload();
            return true;
        }).catch(error => {
            logger.error('failed to dismiss review finding', error);
            errorMessage.value = tt('Failed to save data');
            dismissing.value = false;
            return false;
        });
    }

    function reload(): void {
        loading.value = true;
        errorMessage.value = '';

        services.getTransactionReview({ months: months.value }).then(response => {
            review.value = response.data.result;
            loading.value = false;
        }).catch(error => {
            logger.error('failed to load transaction review', error);
            errorMessage.value = tt('Failed to load data');
            loading.value = false;
        });
    }

    function init(): void {
        reload();
    }

    return {
        // constants
        reviewPeriodOptions: REVIEW_PERIOD_OPTIONS,
        // states
        loading,
        dismissing,
        months,
        review,
        errorMessage,
        // computed states
        duplicateGroups,
        unusualItems,
        subscriptions,
        hasFindings,
        periodOptions,
        estimatedMonthlyCosts,
        reviewPeriod,
        // functions
        accountName,
        accountCurrency,
        categoryName,
        formatAmount,
        formatAmountInCurrency,
        formatTime,
        cadenceName,
        dismiss,
        reload,
        init
    };
}
