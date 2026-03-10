// Unit tests for common utility functions
import { describe, expect, test } from '@jest/globals';

import {
    isDefined,
    isFunction,
    isObject,
    isArray,
    isString,
    isNumber,
    isInteger,
    isBoolean,
    isYearMonth,
    isEquals,
    isYearMonthEquals,
    isArray1SubsetOfArray2,
    isObjectEmpty,
    replaceAll,
    removeAll
} from '@/lib/common.ts';

// ── Type Guards ───────────────────────────────────────────────────

describe('isDefined', () => {
    test('returns true for primitive values', () => {
        expect(isDefined(0)).toBe(true);
        expect(isDefined('')).toBe(true);
        expect(isDefined(false)).toBe(true);
        expect(isDefined(42)).toBe(true);
    });

    test('returns false for null and undefined', () => {
        expect(isDefined(null)).toBe(false);
        expect(isDefined(undefined)).toBe(false);
    });

    test('returns true for objects and arrays', () => {
        expect(isDefined({})).toBe(true);
        expect(isDefined([])).toBe(true);
    });
});

describe('isFunction', () => {
    test('returns true for functions', () => {
        expect(isFunction(() => {})).toBe(true);
        expect(isFunction(Math.max)).toBe(true);
    });

    test('returns false for non-functions', () => {
        expect(isFunction(42)).toBe(false);
        expect(isFunction('hello')).toBe(false);
        expect(isFunction(null)).toBe(false);
    });
});

describe('isObject', () => {
    test('returns true for plain objects', () => {
        expect(isObject({})).toBe(true);
        expect(isObject({ a: 1 })).toBe(true);
    });

    test('returns false for arrays, null, and primitives', () => {
        expect(isObject([])).toBe(false);
        expect(isObject(null)).toBe(false);
        expect(isObject(42)).toBe(false);
        expect(isObject('str')).toBe(false);
    });
});

describe('isArray', () => {
    test('returns true for arrays', () => {
        expect(isArray([])).toBe(true);
        expect(isArray([1, 2, 3])).toBe(true);
    });

    test('returns false for non-arrays', () => {
        expect(isArray({})).toBe(false);
        expect(isArray('hello')).toBe(false);
        expect(isArray(null)).toBe(false);
    });
});

describe('isString', () => {
    test('returns true for strings', () => {
        expect(isString('')).toBe(true);
        expect(isString('hello')).toBe(true);
    });

    test('returns false for non-strings', () => {
        expect(isString(42)).toBe(false);
        expect(isString(null)).toBe(false);
        expect(isString(undefined)).toBe(false);
    });
});

describe('isNumber', () => {
    test('returns true for numbers', () => {
        expect(isNumber(0)).toBe(true);
        expect(isNumber(3.14)).toBe(true);
        expect(isNumber(-100)).toBe(true);
        expect(isNumber(NaN)).toBe(true);
    });

    test('returns false for non-numbers', () => {
        expect(isNumber('42')).toBe(false);
        expect(isNumber(null)).toBe(false);
    });
});

describe('isInteger', () => {
    test('returns true for integers', () => {
        expect(isInteger(0)).toBe(true);
        expect(isInteger(42)).toBe(true);
        expect(isInteger(-7)).toBe(true);
    });

    test('returns false for floats and non-numbers', () => {
        expect(isInteger(3.14)).toBe(false);
        expect(isInteger('42')).toBe(false);
        expect(isInteger(NaN)).toBe(false);
    });
});

describe('isBoolean', () => {
    test('returns true for booleans', () => {
        expect(isBoolean(true)).toBe(true);
        expect(isBoolean(false)).toBe(true);
    });

    test('returns false for non-booleans', () => {
        expect(isBoolean(0)).toBe(false);
        expect(isBoolean(1)).toBe(false);
        expect(isBoolean('true')).toBe(false);
        expect(isBoolean(null)).toBe(false);
    });
});

// ── isYearMonth ───────────────────────────────────────────────────

describe('isYearMonth', () => {
    test('returns true for valid year-month strings', () => {
        expect(isYearMonth('2024-01')).toBe(true);
        expect(isYearMonth('2024-12')).toBe(true);
        expect(isYearMonth('1999-06')).toBe(true);
    });

    test('returns false for invalid formats', () => {
        expect(isYearMonth('2024')).toBe(false);
        expect(isYearMonth('2024-01-01')).toBe(false);
        expect(isYearMonth('')).toBe(false);
        expect(isYearMonth('abc-def')).toBe(false);
        expect(isYearMonth('0-0')).toBe(false);
    });

    test('returns false for non-string types', () => {
        expect(isYearMonth(202401)).toBe(false);
        expect(isYearMonth(null)).toBe(false);
        expect(isYearMonth(undefined)).toBe(false);
    });
});

describe('isYearMonthEquals', () => {
    test('returns true for equal year-month values', () => {
        expect(isYearMonthEquals('2024-01', '2024-01')).toBe(true);
        expect(isYearMonthEquals('2024-1', '2024-01')).toBe(true);
    });

    test('returns false for different year-month values', () => {
        expect(isYearMonthEquals('2024-01', '2024-02')).toBe(false);
        expect(isYearMonthEquals('2023-01', '2024-01')).toBe(false);
    });

    test('returns false for invalid inputs', () => {
        expect(isYearMonthEquals('invalid', '2024-01')).toBe(false);
        expect(isYearMonthEquals('2024', '2024-01')).toBe(false);
    });
});

// ── isEquals (deep equality) ──────────────────────────────────────

describe('isEquals', () => {
    test('returns true for identical primitives', () => {
        expect(isEquals(42, 42)).toBe(true);
        expect(isEquals('hello', 'hello')).toBe(true);
        expect(isEquals(true, true)).toBe(true);
        expect(isEquals(null, null)).toBe(true);
    });

    test('returns false for different primitives', () => {
        expect(isEquals(42, 43)).toBe(false);
        expect(isEquals('a', 'b')).toBe(false);
    });

    test('deep-compares arrays', () => {
        expect(isEquals([1, 2, 3], [1, 2, 3])).toBe(true);
        expect(isEquals([1, 2], [1, 2, 3])).toBe(false);
        expect(isEquals([1, 2, 3], [1, 2, 4])).toBe(false);
        expect(isEquals([], [])).toBe(true);
    });

    test('deep-compares nested arrays', () => {
        expect(isEquals([[1, 2], [3]], [[1, 2], [3]])).toBe(true);
        expect(isEquals([[1, 2], [3]], [[1, 2], [4]])).toBe(false);
    });

    test('deep-compares objects', () => {
        expect(isEquals({ a: 1, b: 2 }, { a: 1, b: 2 })).toBe(true);
        expect(isEquals({ a: 1 }, { a: 2 })).toBe(false);
        expect(isEquals({ a: 1 }, { a: 1, b: 2 })).toBe(false);
        expect(isEquals({}, {})).toBe(true);
    });

    test('deep-compares nested objects', () => {
        expect(isEquals({ a: { b: 1 } }, { a: { b: 1 } })).toBe(true);
        expect(isEquals({ a: { b: 1 } }, { a: { b: 2 } })).toBe(false);
    });
});

// ── isArray1SubsetOfArray2 ────────────────────────────────────────

describe('isArray1SubsetOfArray2', () => {
    test('returns true when array1 is a subset', () => {
        expect(isArray1SubsetOfArray2([1, 2], [1, 2, 3])).toBe(true);
        expect(isArray1SubsetOfArray2([], [1, 2])).toBe(true);
    });

    test('returns false when array1 is not a subset', () => {
        expect(isArray1SubsetOfArray2([1, 4], [1, 2, 3])).toBe(false);
        expect(isArray1SubsetOfArray2([1, 2, 3], [1, 2])).toBe(false);
    });
});

// ── isObjectEmpty ─────────────────────────────────────────────────

describe('isObjectEmpty', () => {
    test('returns true for empty objects and falsy values', () => {
        expect(isObjectEmpty({})).toBe(true);
    });

    test('returns false for non-empty objects', () => {
        expect(isObjectEmpty({ a: 1 })).toBe(false);
    });
});

// ── String Utilities ──────────────────────────────────────────────

describe('replaceAll', () => {
    test('replaces all occurrences', () => {
        expect(replaceAll('hello world hello', 'hello', 'hi')).toBe('hi world hi');
    });

    test('handles special regex characters in search string', () => {
        expect(replaceAll('price is $10.00', '$10.00', '$20.00')).toBe('price is $20.00');
    });

    test('returns unchanged string if no match', () => {
        expect(replaceAll('hello', 'xyz', 'abc')).toBe('hello');
    });
});

describe('removeAll', () => {
    test('removes all occurrences', () => {
        expect(removeAll('a,b,c,d', ',')).toBe('abcd');
    });

    test('removes nothing if no match', () => {
        expect(removeAll('hello', 'x')).toBe('hello');
    });
});
