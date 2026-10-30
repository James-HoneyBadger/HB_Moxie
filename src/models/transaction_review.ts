export interface TransactionReviewTransaction {
    readonly id: string;
    readonly time: number;
    readonly amount: number;
    readonly comment: string;
    readonly categoryId: string;
    readonly accountId: string;
}

export interface TransactionReviewDuplicateGroup {
    readonly accountId: string;
    readonly type: number;
    readonly amount: number;
    readonly transactions: TransactionReviewTransaction[];
}

export interface TransactionReviewUnusual {
    readonly transaction: TransactionReviewTransaction;
    readonly categoryMean: number;
    readonly categoryStdDev: number;
    readonly zScore: number;
    readonly sampleCount: number;
}

export interface TransactionReviewSubscription {
    readonly name: string;
    readonly cadence: string;
    readonly occurrences: number;
    readonly averageAmount: number;
    readonly firstAmount: number;
    readonly lastAmount: number;
    readonly priceChange: number;
    readonly firstTime: number;
    readonly lastTime: number;
    readonly accountId: string;
    readonly categoryId: string;
}

export interface TransactionReviewResponse {
    readonly startTime: number;
    readonly endTime: number;
    readonly totalTransactions: number;
    readonly duplicates: TransactionReviewDuplicateGroup[];
    readonly unusual: TransactionReviewUnusual[];
    readonly subscriptions: TransactionReviewSubscription[];
}
