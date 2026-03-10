import type { HebrewCalendarLocaleData } from '@/core/calendar.ts';
import type { Year1BasedMonth, YearMonthDay, CalendarAlternateDate } from '@/core/datetime.ts';

import { getGregorianCalendarYearMonthDays } from '../datetime.ts';

// Hebrew month values (1-based)
export type HebrewMonthValue = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13;
export type HebrewDayValue = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24 | 25 | 26 | 27 | 28 | 29 | 30;

export interface HebrewYearMonthDayInfo {
    readonly gregorianYear: number;
    readonly gregorianMonth: number;
    readonly gregorianDay: number;
    readonly year: number;
    readonly month: HebrewMonthValue;
    readonly day: HebrewDayValue;
    readonly displayYear: string;
    readonly displayMonth: string;
    readonly displayDay: string;
    readonly isLeapYear: boolean;
}

// Hebrew calendar epoch: Monday, October 7, 3761 BCE (Julian) = JD 347995.5
const HEBREW_EPOCH = 347995.5;

/**
 * Check if a Hebrew year is a leap year.
 * A Hebrew year is a leap year if (7*year + 1) mod 19 < 7.
 */
export function isHebrewLeapYear(year: number): boolean {
    return ((7 * year + 1) % 19) < 7;
}

/**
 * Get the number of months in a Hebrew year.
 */
function getHebrewMonthsInYear(year: number): number {
    return isHebrewLeapYear(year) ? 13 : 12;
}

/**
 * Compute the Hebrew calendar delay of year 1 (Molad Tohu).
 * Returns the number of days from the epoch to the 1st of Tishrei (RH).
 */
function hebrewDelay1(year: number): number {
    const months = Math.floor((235 * year - 234) / 19);
    const parts = 12084 + 13753 * months;
    let day = months * 29 + Math.floor(parts / 25920);

    if ((3 * (day + 1)) % 7 < 3) {
        day += 1;
    }

    return day;
}

/**
 * Compute the Hebrew calendar delay of year 2.
 */
function hebrewDelay2(year: number): number {
    const last = hebrewDelay1(year - 1);
    const present = hebrewDelay1(year);
    const next = hebrewDelay1(year + 1);

    if (next - present === 356) {
        return 2;
    }

    if (present - last === 382) {
        return 1;
    }

    return 0;
}

/**
 * Get the number of days in a Hebrew year.
 */
function getHebrewYearDays(year: number): number {
    return hebrewNewYear(year + 1) - hebrewNewYear(year);
}

/**
 * Get the day number (from epoch) of 1 Tishrei of the given Hebrew year.
 */
function hebrewNewYear(year: number): number {
    return hebrewDelay1(year) + hebrewDelay2(year);
}

/**
 * Get the number of days in a specific month of a Hebrew year.
 */
export function getHebrewMonthDays(year: number, month: number): number {
    // Months: 1=Tishrei, 2=Cheshvan, 3=Kislev, 4=Tevet, 5=Shevat,
    //         6=Adar (or Adar I in leap year), 7=Adar II (leap year only),
    //         8=Nisan, 9=Iyar, 10=Sivan, 11=Tammuz, 12=Av, 13=Elul
    // In non-leap years: months 1-6, then 8-13 (skip 7)
    // In leap years: months 1-13

    if (month === 2) {
        // Cheshvan: 29 or 30 days
        const yearDays = getHebrewYearDays(year);
        return (yearDays % 10 === 5) ? 30 : 29;
    }

    if (month === 3) {
        // Kislev: 29 or 30 days
        const yearDays = getHebrewYearDays(year);
        return (yearDays % 10 === 3) ? 29 : 30;
    }

    if (month === 4 || month === 7 || month === 9 || month === 11 || month === 13) {
        return 29;
    }

    if (month === 6 && !isHebrewLeapYear(year)) {
        return 29; // Adar in non-leap year
    }

    // Tishrei(1)=30, Shevat(5)=30, Adar I in leap(6)=30, Nisan(8)=30, Sivan(10)=30, Av(12)=30
    return 30;
}

/**
 * Convert a Julian Day Number to a Hebrew date.
 */
function jdToHebrew(jd: number): { year: number; month: number; day: number } {
    const jdFloor = Math.floor(jd) + 0.5;
    const count = Math.floor(((jdFloor - HEBREW_EPOCH) * 98496.0) / 35975351.0);
    let year = count - 1;

    // Find the correct year
    for (let i = year; i <= year + 2; i++) {
        if (hebrewNewYear(i) + HEBREW_EPOCH > jdFloor) {
            year = i - 1;
            break;
        }
    }

    const months = getHebrewMonthsInYear(year);
    // Start from Tishrei (month 1)
    let first = 1;

    if (jdFloor < hebrewToJd(year, 1 , 1)) {
        // Before Tishrei 1 of this year - should not normally happen with correct year calc
        first = 1;
    }

    let month = first;
    let day = 1;

    // Find the correct month
    for (let m = first; m <= months; m++) {
        const nextMonthJd = hebrewToJd(year, m, getHebrewMonthDays(year, m)) + 1;
        if (jdFloor < nextMonthJd) {
            month = m;
            break;
        }
        month = m + 1;
    }

    // Find the correct day
    day = Math.floor(jdFloor - hebrewToJd(year, month, 1)) + 1;

    return { year, month, day };
}

/**
 * Convert a Hebrew date to a Julian Day Number.
 */
function hebrewToJd(year: number, month: number, day: number): number {
    let jd = HEBREW_EPOCH + hebrewNewYear(year) + day - 1;

    if (month < 7) {
        // Months from Tishrei to Adar (or Adar I/II)
        for (let m = 1; m < month; m++) {
            jd += getHebrewMonthDays(year, m);
        }
    } else {
        // Months from Nisan onward - count from Tishrei through the first half
        for (let m = 1; m <= 6; m++) {
            jd += getHebrewMonthDays(year, m);
        }

        if (isHebrewLeapYear(year)) {
            jd += getHebrewMonthDays(year, 7); // Adar II
        }

        // Months from Nisan (8) to the target month
        for (let m = 8; m < month; m++) {
            jd += getHebrewMonthDays(year, m);
        }
    }

    return jd;
}

/**
 * Convert a Gregorian date to a Julian Day Number.
 */
function gregorianToJd(year: number, month: number, day: number): number {
    const a = Math.floor((14 - month) / 12);
    const y = year + 4800 - a;
    const m = month + 12 * a - 3;
    return day + Math.floor((153 * m + 2) / 5) + 365 * y + Math.floor(y / 4) - Math.floor(y / 100) + Math.floor(y / 400) - 32045;
}

/**
 * Convert a Gregorian date to a Hebrew date.
 */
export function gregorianToHebrew(gYear: number, gMonth: number, gDay: number): { year: number; month: number; day: number } {
    const jd = gregorianToJd(gYear, gMonth, gDay);
    return jdToHebrew(jd);
}

/**
 * Get display month name from Hebrew calendar locale data.
 * Month mapping: 1=Tishrei, 2=Cheshvan, 3=Kislev, 4=Tevet, 5=Shevat,
 * 6=Adar/Adar I, 7=Adar II, 8=Nisan, 9=Iyar, 10=Sivan, 11=Tammuz, 12=Av, 13=Elul
 */
function getHebrewMonthDisplayName(month: number, isLeapYear: boolean, localeData: HebrewCalendarLocaleData): string {
    // In a non-leap year, there's no month 7 (Adar II)
    // The locale data monthNames has 13 entries: indices 0-12 for months 1-13
    if (!isLeapYear && month >= 7) {
        // In non-leap year, months 8-13 map to what would logically follow Adar
        // But we still use the same month number system
        return localeData.monthNames[month - 1] ?? '';
    }

    return localeData.monthNames[month - 1] ?? '';
}



/**
 * Get Hebrew date info for a specific Gregorian date.
 */
export function getHebrewYearMonthDayInfo(yearMonthDay: YearMonthDay, localeData: HebrewCalendarLocaleData): HebrewYearMonthDayInfo | undefined {
    const hebrew = gregorianToHebrew(yearMonthDay.year, yearMonthDay.month, yearMonthDay.day);
    const leap = isHebrewLeapYear(hebrew.year);

    return {
        gregorianYear: yearMonthDay.year,
        gregorianMonth: yearMonthDay.month,
        gregorianDay: yearMonthDay.day,
        year: hebrew.year,
        month: hebrew.month as HebrewMonthValue,
        day: hebrew.day as HebrewDayValue,
        displayYear: hebrew.year.toString(),
        displayMonth: getHebrewMonthDisplayName(hebrew.month, leap, localeData),
        displayDay: hebrew.day.toString(),
        isLeapYear: leap
    };
}

/**
 * Get Hebrew date info for all days in a Gregorian year-month.
 */
export function getHebrewYearMonthAllDayInfos(yearMonth: Year1BasedMonth, localeData: HebrewCalendarLocaleData): HebrewYearMonthDayInfo[] {
    const monthDays = getGregorianCalendarYearMonthDays(yearMonth);
    const result: HebrewYearMonthDayInfo[] = [];

    for (let day = 1; day <= monthDays; day++) {
        const info = getHebrewYearMonthDayInfo(
            { year: yearMonth.year, month: yearMonth.month1base, day },
            localeData
        );

        if (info) {
            result.push(info);
        }
    }

    return result;
}

/**
 * Get the alternate display date for a Hebrew calendar date (used in calendar overlay).
 */
export function getHebrewCalendarAlternateDisplayDate(info: HebrewYearMonthDayInfo): CalendarAlternateDate {
    // Show month name on day 1, otherwise show day number
    let displayDate: string;

    if (info.day === 1) {
        displayDate = info.displayMonth;
    } else {
        displayDate = info.displayDay;
    }

    return {
        year: info.gregorianYear,
        month: info.gregorianMonth,
        day: info.gregorianDay,
        displayDate: displayDate
    };
}
