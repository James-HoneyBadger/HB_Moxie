export enum SplitDirection {
    TheyOweMe = 1,
    IOweThem = 2
}

export interface TransactionSplitInfo {
    readonly personName: string;
    readonly currency: string;
    readonly amount: number;
}

export interface TransactionSplits {
    readonly transactionId: string;
    readonly transactionAmount: number;
    readonly currency: string;
    readonly direction: number;
    readonly splits: TransactionSplitInfo[];
}

export interface SharedExpenseSummaryItem {
    readonly personName: string;
    readonly currency: string;
    /** Positive when the person owes the user, negative when the user owes them. */
    readonly total: number;
    readonly owedToMe: number;
    readonly iOwe: number;
    readonly settled: number;
    readonly entryCount: number;
    readonly lastTime: number;
}

export interface SharedExpenseEntry {
    readonly transactionId: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly categoryId: string;
    readonly accountId: string;
    readonly transactionAmount: number;
    /** Positive when the person owes the user, negative when the user owes them. */
    readonly splitAmount: number;
    readonly currency: string;
    readonly comment: string;
    readonly direction: number;
    readonly isSettlement: boolean;
    readonly settlementId: string;
}

export interface TransactionSplitRequestItem {
    personName: string;
    amount: number;
}

export interface TransactionSplitUpdateRequest {
    transactionId: string;
    direction?: number;
    splits: TransactionSplitRequestItem[];
}

export interface TransactionSplitIndexItem {
    readonly transactionId: string;
    readonly total: number;
    readonly currency: string;
    readonly personCount: number;
}

export interface TransactionSplitRenamePersonRequest {
    oldName: string;
    newName: string;
}

export interface SharedExpenseSettlementRequest {
    personName: string;
    currency: string;
    amount: number;
    time?: number;
    comment: string;
}

export interface SharedExpenseSettlementInfo {
    readonly settlementId: string;
    readonly personName: string;
    readonly currency: string;
    readonly amount: number;
    readonly time: number;
    readonly comment: string;
}
