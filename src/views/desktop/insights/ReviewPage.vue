<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="d-flex align-center flex-wrap">
                        <span>{{ tt('Transaction Review') }}</span>
                        <v-spacer />
                        <v-select class="review-period-select"
                                  density="compact" variant="outlined" hide-details
                                  :items="periodOptions" v-model="months"
                                  @update:model-value="reload" />
                        <v-btn class="ms-3" color="default" variant="text" :disabled="loading" @click="reload">
                            <v-icon :icon="mdiRefresh" start />
                            {{ tt('Refresh') }}
                        </v-btn>
                    </div>
                </template>

                <v-progress-linear v-if="loading" indeterminate />

                <v-card-text>
                    <div class="text-body-2 mb-2">
                        {{ tt('Checks recent transactions for possible duplicate charges, unusual spending and recurring subscriptions.') }}
                    </div>
                    <div class="text-caption text-medium-emphasis mb-5" v-if="reviewPeriod">
                        {{ tt('Review Period') }}: {{ reviewPeriod }}
                    </div>

                    <v-alert v-if="errorMessage" type="error" variant="tonal" class="mb-4">{{ errorMessage }}</v-alert>

                    <v-row v-if="!loading && review">
                        <v-col cols="12" md="4">
                            <div class="review-stat">
                                <div class="text-h4">{{ duplicateGroups.length }}</div>
                                <div class="text-body-2">{{ tt('Possible Duplicate Charges') }}</div>
                            </div>
                        </v-col>
                        <v-col cols="12" md="4">
                            <div class="review-stat">
                                <div class="text-h4">{{ unusualItems.length }}</div>
                                <div class="text-body-2">{{ tt('Unusual Spending') }}</div>
                            </div>
                        </v-col>
                        <v-col cols="12" md="4">
                            <div class="review-stat">
                                <div class="text-h4" v-if="estimatedMonthlyCosts.length < 1">—</div>
                                <div class="text-h4" :key="cost.currency" v-for="cost in estimatedMonthlyCosts">
                                    {{ formatAmountInCurrency(cost.amount, cost.currency) }}
                                </div>
                                <div class="text-body-2">{{ tt('Estimated Monthly Cost') }} · {{ tt('Subscriptions') }}: {{ subscriptions.length }}</div>
                            </div>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" md="6">
            <v-card height="100%">
                <template #title>
                    <div class="d-flex align-center">
                        <v-icon :icon="mdiContentDuplicate" start />
                        <span>{{ tt('Possible Duplicate Charges') }}</span>
                    </div>
                </template>
                <v-card-text>
                    <div v-if="!loading && duplicateGroups.length < 1" class="text-body-2 text-medium-emphasis">
                        {{ tt('No duplicate charges found') }}
                    </div>
                    <div class="review-item" :key="index" v-for="(group, index) in duplicateGroups">
                        <div class="d-flex align-center">
                            <span class="text-subtitle-2">{{ accountName(group.accountId) }}</span>
                            <v-spacer />
                            <span class="text-subtitle-2">{{ formatAmount(group.amount, group.accountId) }}</span>
                        </div>
                        <div class="text-caption text-medium-emphasis mt-1"
                             :key="transaction.id" v-for="transaction in group.transactions">
                            {{ formatTime(transaction.time) }}
                            <template v-if="transaction.comment"> · {{ transaction.comment }}</template>
                        </div>
                    </div>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" md="6">
            <v-card height="100%">
                <template #title>
                    <div class="d-flex align-center">
                        <v-icon :icon="mdiAlertOutline" start />
                        <span>{{ tt('Unusual Spending') }}</span>
                    </div>
                </template>
                <v-card-text>
                    <div v-if="!loading && unusualItems.length < 1" class="text-body-2 text-medium-emphasis">
                        {{ tt('No unusual spending found') }}
                    </div>
                    <div class="review-item" :key="item.transaction.id" v-for="item in unusualItems">
                        <div class="d-flex align-center">
                            <span class="text-subtitle-2">{{ categoryName(item.transaction.categoryId) }}</span>
                            <v-spacer />
                            <span class="text-subtitle-2">{{ formatAmount(item.transaction.amount, item.transaction.accountId) }}</span>
                        </div>
                        <div class="text-caption text-medium-emphasis mt-1">
                            {{ formatTime(item.transaction.time) }}
                            · {{ tt('Category Average') }} {{ formatAmount(item.categoryMean, item.transaction.accountId) }}
                            · {{ tt('Z-Score') }} {{ item.zScore }}
                        </div>
                    </div>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="d-flex align-center">
                        <v-icon :icon="mdiAutorenew" start />
                        <span>{{ tt('Subscriptions') }}</span>
                    </div>
                </template>
                <v-card-text>
                    <div v-if="!loading && subscriptions.length < 1" class="text-body-2 text-medium-emphasis">
                        {{ tt('No subscriptions detected') }}
                    </div>
                    <v-table v-if="subscriptions.length > 0" density="compact">
                        <thead>
                        <tr>
                            <th>{{ tt('Name') }}</th>
                            <th>{{ tt('Cadence') }}</th>
                            <th class="text-end">{{ tt('Occurrences') }}</th>
                            <th class="text-end">{{ tt('Average Amount') }}</th>
                            <th class="text-end">{{ tt('Price Change') }}</th>
                            <th>{{ tt('Last Charge') }}</th>
                        </tr>
                        </thead>
                        <tbody>
                        <tr :key="subscription.name + subscription.lastTime" v-for="subscription in subscriptions">
                            <td>{{ subscription.name }}</td>
                            <td>{{ cadenceName(subscription.cadence) }}</td>
                            <td class="text-end">{{ subscription.occurrences }}</td>
                            <td class="text-end">{{ formatAmount(subscription.averageAmount, subscription.accountId) }}</td>
                            <td class="text-end">
                                <v-chip v-if="subscription.priceChange !== 0"
                                        size="small" variant="tonal"
                                        :color="subscription.priceChange > 0 ? 'error' : 'success'">
                                    {{ subscription.priceChange > 0 ? '+' : '' }}{{ formatAmount(subscription.priceChange, subscription.accountId) }}
                                </v-chip>
                                <span v-else class="text-medium-emphasis">—</span>
                            </td>
                            <td>{{ formatTime(subscription.lastTime) }}</td>
                        </tr>
                        </tbody>
                    </v-table>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>
</template>

<script setup lang="ts">
import { mdiAlertOutline, mdiAutorenew, mdiContentDuplicate, mdiRefresh } from '@mdi/js';

import { useI18n } from '@/locales/helpers.ts';
import { useTransactionReviewPageBase } from '@/views/base/insights/TransactionReviewPageBase.ts';

const { tt } = useI18n();
const {
    periodOptions,
    loading,
    months,
    review,
    errorMessage,
    duplicateGroups,
    unusualItems,
    subscriptions,
    estimatedMonthlyCosts,
    reviewPeriod,
    accountName,
    categoryName,
    formatAmount,
    formatAmountInCurrency,
    formatTime,
    cadenceName,
    reload,
    init
} = useTransactionReviewPageBase();

init();
</script>

<style scoped>
.review-period-select {
    max-width: 140px;
}

.review-stat {
    padding: 8px 0;
}

.review-item {
    padding: 10px 0;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.review-item:last-child {
    border-bottom: none;
}
</style>
