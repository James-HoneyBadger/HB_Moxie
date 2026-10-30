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
                            <th class="text-end">{{ tt('Balance') }}</th>
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
                            <td class="text-end" :class="item.total < 0 ? 'text-error' : 'text-success'">
                                <div>{{ formatAmountInCurrency(Math.abs(item.total), item.currency) }}</div>
                                <div class="text-caption text-medium-emphasis">
                                    {{ item.total < 0 ? tt('You Owe') : tt('Owes You') }}
                                </div>
                            </td>
                            <td class="text-end">{{ item.entryCount }}</td>
                            <td>{{ formatTime(item.lastTime) }}</td>
                            <td class="text-end text-no-wrap">
                                <v-btn size="small" variant="text" :disabled="settling" @click.stop="openSettle(item)">
                                    {{ tt('Settle Up') }}
                                </v-btn>
                                <v-btn size="small" variant="text" :disabled="settling" @click.stop="openRename(item)">
                                    {{ tt('Rename') }}
                                </v-btn>
                                <v-icon class="ms-1" :icon="isPersonSelected(item) ? mdiChevronUp : mdiChevronDown" size="20" />
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
                        <tr :key="(entry.isSettlement ? entry.settlementId : entry.transactionId) + entry.time"
                            v-for="entry in personEntries">
                            <td>{{ formatTime(entry.time) }}</td>
                            <td>
                                <v-chip v-if="entry.isSettlement" size="small" variant="tonal">{{ tt('Settlement') }}</v-chip>
                                <template v-else>{{ categoryName(entry.categoryId) }}</template>
                            </td>
                            <td class="text-end">
                                <template v-if="!entry.isSettlement">{{ formatAmountInCurrency(entry.transactionAmount, entry.currency) }}</template>
                                <template v-else>—</template>
                            </td>
                            <td class="text-end" :class="entry.splitAmount < 0 ? 'text-error' : 'text-success'">
                                {{ formatAmountInCurrency(Math.abs(entry.splitAmount), entry.currency) }}
                                <div class="text-caption text-medium-emphasis">
                                    {{ entry.splitAmount < 0 ? tt('You Owe') : tt('Owes You') }}
                                </div>
                            </td>
                            <td class="text-truncate">{{ entry.comment }}</td>
                            <td class="text-end text-no-wrap">
                                <template v-if="entry.isSettlement">
                                    <v-btn size="small" variant="text" color="error" :disabled="settling" @click="deleteSettlementFor(entry)">
                                        {{ tt('Delete') }}
                                    </v-btn>
                                </template>
                                <template v-else>
                                    <v-btn size="small" variant="text" :disabled="clearing" @click="openEditSplit(entry)">
                                        {{ tt('Edit') }}
                                    </v-btn>
                                    <v-btn size="small" variant="text" color="error" :disabled="clearing" @click="clearSplit(entry)">
                                        {{ tt('Clear') }}
                                    </v-btn>
                                </template>
                            </td>
                        </tr>
                        </tbody>
                    </v-table>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>

    <v-dialog width="420" :persistent="settling" v-model="showSettleDialog">
        <v-card>
            <template #title>
                <div class="d-flex align-center">
                    <v-icon :icon="mdiCashMultiple" start />
                    <span>{{ tt('Settle Up') }}</span>
                </div>
            </template>
            <v-card-text>
                <div class="text-body-2 mb-3">{{ settlePersonName }} · {{ settleCurrency }}</div>

                <v-btn-toggle class="mb-3" density="compact" variant="outlined" v-model="settleDirection">
                    <v-btn :value="1" size="small">{{ tt('They Paid Me') }}</v-btn>
                    <v-btn :value="-1" size="small">{{ tt('I Paid Them') }}</v-btn>
                </v-btn-toggle>

                <amount-input :currency="settleCurrency" :show-currency="true"
                              :label="tt('Amount')" :placeholder="tt('Amount')"
                              :disabled="settling" v-model="settleAmount" />

                <v-text-field class="mt-3" density="compact" variant="outlined"
                              :label="tt('Description')" :disabled="settling"
                              v-model="settleComment" />
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn color="default" variant="text" :disabled="settling" @click="showSettleDialog = false">
                    {{ tt('Cancel') }}
                </v-btn>
                <v-btn color="primary" :loading="settling" :disabled="!canSettle" @click="submitSettlement">
                    {{ tt('Save') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <rename-dialog ref="renameDialog" />

    <split-expense-dialog ref="splitDialog" @saved="loadSummary" />
</template>

<script setup lang="ts">
import { computed, ref, useTemplateRef } from 'vue';

import {
    mdiAccountGroupOutline,
    mdiAccountPlusOutline,
    mdiCashMultiple,
    mdiChevronDown,
    mdiChevronUp,
    mdiFormatListBulleted,
    mdiRefresh
} from '@mdi/js';

import RenameDialog from '@/components/desktop/RenameDialog.vue';
import SplitExpenseDialog from './dialogs/SplitExpenseDialog.vue';

import type { SharedExpenseEntry, SharedExpenseSummaryItem } from '@/models/transaction_split.ts';

import { useI18n } from '@/locales/helpers.ts';
import { useSharedExpensesPageBase } from '@/views/base/insights/SharedExpensesPageBase.ts';

// RenameDialog exposes an open() that resolves with the new name.
type RenameDialogType = InstanceType<typeof RenameDialog>;

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
    settling,
    totalOwedByCurrency,
    hasSummary,
    categoryName,
    formatAmountInCurrency,
    formatTime,
    isPersonSelected,
    selectPerson,
    loadSummary,
    clearSplit,
    settleSharedExpense,
    deleteSettlement,
    renamePerson,
    init
} = useSharedExpensesPageBase();

const splitDialog = useTemplateRef<InstanceType<typeof SplitExpenseDialog>>('splitDialog');
const renameDialog = useTemplateRef<RenameDialogType>('renameDialog');

const showSettleDialog = ref<boolean>(false);
const settlePersonName = ref<string>('');
const settleCurrency = ref<string>('');
const settleDirection = ref<number>(1);
const settleAmount = ref<number>(0);
const settleComment = ref<string>('');

const canSettle = computed<boolean>(() => settleAmount.value > 0 && !!settlePersonName.value && !settling.value);

function openNewSplit(): void {
    splitDialog.value?.open({ defaultPersonName: selectedPerson.value });
}

function openEditSplit(entry: SharedExpenseEntry): void {
    splitDialog.value?.open({
        transactionId: entry.transactionId,
        transactionAmount: entry.transactionAmount,
        currency: entry.currency,
        defaultPersonName: selectedPerson.value,
        direction: entry.direction
    });
}

function openSettle(item: SharedExpenseSummaryItem): void {
    settlePersonName.value = item.personName;
    settleCurrency.value = item.currency;
    settleAmount.value = Math.abs(item.total);
    settleDirection.value = item.total < 0 ? -1 : 1;
    settleComment.value = '';
    showSettleDialog.value = true;
}

function submitSettlement(): void {
    if (!canSettle.value) {
        return;
    }

    settleSharedExpense({
        personName: settlePersonName.value,
        currency: settleCurrency.value,
        amount: Math.abs(settleAmount.value) * settleDirection.value,
        comment: settleComment.value
    }).then(saved => {
        if (saved) {
            showSettleDialog.value = false;
        }
    });
}

function openRename(item: SharedExpenseSummaryItem): void {
    renameDialog.value?.open(item.personName, tt('Person Name')).then((newName: string) => {
        if (newName && newName !== item.personName) {
            renamePerson(item.personName, newName);
        }
    }).catch(() => {
        // Dismissed without renaming.
    });
}

function deleteSettlementFor(entry: SharedExpenseEntry): void {
    deleteSettlement(entry.settlementId);
}

init();
</script>

<style scoped>
.shared-expense-stat {
    padding: 8px 0;
}
</style>
