// Unit tests for numeral utility functions
import { describe, expect, test } from '@jest/globals';

import type { NumberFormatOptions } from '@/core/numeral.ts';
import { NumeralSystem, DecimalSeparator, DigitGroupingSymbol, DigitGroupingType } from '@/core/numeral.ts';

import {
    sumAmounts,
    parseAmount,
    formatAmount,
    appendDigitGroupingSymbolAndDecimalSeparator
} from '@/lib/numeral.ts';

// Default options matching the most common (Western Arabic / dot / comma) configuration
const defaultOptions: NumberFormatOptions = {
    numeralSystem: NumeralSystem.Default,
    decimalSeparator: DecimalSeparator.Default.symbol,
    digitGroupingSymbol: DigitGroupingSymbol.Default.symbol,
    digitGrouping: DigitGroupingType.Default
};

// ── sumAmounts ─────────────────────────────────────────────────────

describe('sumAmounts', () => {
    test('returns 0 for empty array', () => {
        expect(sumAmounts([])).toBe(0);
    });

    test('sums a single element', () => {
        expect(sumAmounts([4200])).toBe(4200);
    });

    test('sums multiple positive values', () => {
        expect(sumAmounts([100, 200, 300])).toBe(600);
    });

    test('sums mixed positive and negative values', () => {
        expect(sumAmounts([100, -50, 200, -150])).toBe(100);
    });

    test('handles large numbers', () => {
        expect(sumAmounts([1000000, 2000000, 3000000])).toBe(6000000);
    });
});

// ── parseAmount ────────────────────────────────────────────────────
// parseAmount converts a formatted string (human-readable) to an integer
// representing the amount in cents (×100).

describe('parseAmount', () => {
    test('parses a simple integer string to cents', () => {
        // "42" → 42 dollars → 4200 cents
        expect(parseAmount('42', defaultOptions)).toBe(4200);
    });

    test('parses a string with decimal separator', () => {
        // "42.50" → 4250 cents
        expect(parseAmount('42.50', defaultOptions)).toBe(4250);
    });

    test('parses a string with one decimal digit', () => {
        // "42.5" → 42.5 dollars → 4250 cents
        expect(parseAmount('42.5', defaultOptions)).toBe(4250);
    });

    test('truncates beyond 2 decimal digits', () => {
        // "42.999" → reads "99" → 4299
        expect(parseAmount('42.999', defaultOptions)).toBe(4299);
    });

    test('parses negative amounts', () => {
        expect(parseAmount('-10.50', defaultOptions)).toBe(-1050);
    });

    test('returns 0 for empty or null-ish input', () => {
        expect(parseAmount('', defaultOptions)).toBe(0);
        expect(parseAmount(null as unknown as string, defaultOptions)).toBe(0);
    });

    test('strips digit grouping symbols', () => {
        // "1,234" → 1234 dollars → 123400 cents
        expect(parseAmount('1,234', defaultOptions)).toBe(123400);
    });

    test('handles leading decimal separator', () => {
        // ".50" → 0.50 → 50 cents
        expect(parseAmount('.50', defaultOptions)).toBe(50);
    });

    test('handles integer with trailing decimal separator', () => {
        // "42." → 42 dollars → 4200
        expect(parseAmount('42.', defaultOptions)).toBe(4200);
    });
});

// ── formatAmount ───────────────────────────────────────────────────
// formatAmount converts an integer (cents) to a formatted display string.

describe('formatAmount', () => {
    test('formats zero', () => {
        const result = formatAmount(0, defaultOptions);
        expect(result).toBe('0.00');
    });

    test('formats a positive amount in cents', () => {
        // 4250 cents → "42.50"
        const result = formatAmount(4250, defaultOptions);
        expect(result).toBe('42.50');
    });

    test('formats a negative amount', () => {
        const result = formatAmount(-4250, defaultOptions);
        expect(result).toBe('-42.50');
    });

    test('formats amount with digit grouping', () => {
        // 1234567 cents → "12,345.67"
        const result = formatAmount(1234567, defaultOptions);
        expect(result).toBe('12,345.67');
    });

    test('formats amounts less than 1 dollar', () => {
        // 50 cents → "0.50"
        const result = formatAmount(50, defaultOptions);
        expect(result).toBe('0.50');
    });

    test('formats single-digit cents', () => {
        // 5 cents → "0.05"
        const result = formatAmount(5, defaultOptions);
        expect(result).toBe('0.05');
    });

    test('throws for non-safe integer', () => {
        expect(() => formatAmount(1.5, defaultOptions)).toThrow();
    });
});

// ── appendDigitGroupingSymbolAndDecimalSeparator ───────────────────

describe('appendDigitGroupingSymbolAndDecimalSeparator', () => {
    test('returns empty string for empty input', () => {
        expect(appendDigitGroupingSymbolAndDecimalSeparator('', defaultOptions)).toBe('');
    });

    test('formats a simple number string with grouping and decimal', () => {
        // "1234567.89" → "1,234,567.89"
        const result = appendDigitGroupingSymbolAndDecimalSeparator('1234567.89', defaultOptions);
        expect(result).toBe('1,234,567.89');
    });

    test('formats a negative number', () => {
        const result = appendDigitGroupingSymbolAndDecimalSeparator('-1234567.89', defaultOptions);
        expect(result).toBe('-1,234,567.89');
    });

    test('formats a number with no decimal part', () => {
        const result = appendDigitGroupingSymbolAndDecimalSeparator('1000', defaultOptions);
        expect(result).toBe('1,000');
    });
});
