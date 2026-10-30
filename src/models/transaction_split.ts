export interface TransactionSplitInfo {
    readonly personName: string;
    readonly currency: string;
    readonly amount: number;
}

export interface TransactionSplits {
    readonly transactionId: string;
    readonly transactionAmount: number;
    readonly currency: string;
    readonly splits: TransactionSplitInfo[];
}

export interface SharedExpenseSummaryItem {
    readonly personName: string;
    readonly currency: string;
    readonly total: number;
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
    readonly splitAmount: number;
    readonly currency: string;
    readonly comment: string;
}

export interface TransactionSplitRequestItem {
    personName: string;
    amount: number;
}

export interface TransactionSplitUpdateRequest {
    transactionId: string;
    splits: TransactionSplitRequestItem[];
}

export interface TransactionSplitIndexItem {
    readonly transactionId: string;
    readonly total: number;
    readonly currency: string;
    readonly personCount: number;
}
