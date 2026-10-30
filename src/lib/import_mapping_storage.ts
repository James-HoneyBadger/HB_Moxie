import { ImportTransactionDataMapping } from '@/core/import_transaction.ts';

import logger from '@/lib/logger.ts';

const STORAGE_KEY = 'honey_ledger_import_mappings';
const MAX_SAVED_MAPPINGS = 20;

// The browser storage the mappings are kept in. Narrowing it to the three methods
// used here keeps the logic testable outside a browser.
export interface ImportMappingStorage {
    getItem(key: string): string | null;
    setItem(key: string, value: string): void;
    removeItem(key: string): void;
}

interface SavedMappingEntry {
    savedAt: number;
    json: string;
}

type SavedMappingEntries = Record<string, SavedMappingEntry>;

export interface ImportMappingStore {
    load(signature: string): ImportTransactionDataMapping | undefined;
    save(signature: string, mapping: ImportTransactionDataMapping): void;
    forget(signature: string): void;
    listSignatures(): string[];
}

/**
 * Identifies a file layout by its first row, so the same export from the same bank
 * finds its mapping again. Files without a header row simply do not match, which
 * only means no saved mapping is applied.
 */
export function getImportMappingSignature(rows?: string[][]): string | undefined {
    const header = rows?.[0];

    if (!header || header.length < 1) {
        return undefined;
    }

    const normalized = header
        .map(cell => (cell ?? '').trim().toLowerCase())
        .filter(cell => cell.length > 0);

    if (normalized.length < 2) {
        return undefined;
    }

    return normalized.slice(0, 40).join('|');
}

export function createImportMappingStore(storage: ImportMappingStorage): ImportMappingStore {
    function readEntries(): SavedMappingEntries {
        try {
            const raw = storage.getItem(STORAGE_KEY);

            if (!raw) {
                return {};
            }

            const parsed = JSON.parse(raw);

            return parsed && typeof parsed === 'object' ? parsed as SavedMappingEntries : {};
        } catch (error) {
            logger.error('failed to read saved import mappings', error);
            return {};
        }
    }

    function writeEntries(entries: SavedMappingEntries): void {
        try {
            storage.setItem(STORAGE_KEY, JSON.stringify(entries));
        } catch (error) {
            logger.error('failed to write saved import mappings', error);
        }
    }

    return {
        load(signature: string): ImportTransactionDataMapping | undefined {
            const entry = readEntries()[signature];

            if (!entry) {
                return undefined;
            }

            return ImportTransactionDataMapping.parseFromJson(entry.json) ?? undefined;
        },
        save(signature: string, mapping: ImportTransactionDataMapping): void {
            const entries = readEntries();

            // A counter beats Date.now() here: several mappings saved in the same
            // millisecond must still have a defined age order.
            const newestFirst = Object.values(entries).map(entry => entry.savedAt);
            const savedAt = Math.max(Date.now(), ...newestFirst.map(value => value + 1));

            entries[signature] = { savedAt, json: mapping.toJson() };

            // Keep the store small: drop the oldest mappings once the cap is hit.
            const ordered = Object.entries(entries).sort((a, b) => b[1].savedAt - a[1].savedAt);

            for (const [key] of ordered.slice(MAX_SAVED_MAPPINGS)) {
                delete entries[key];
            }

            writeEntries(entries);
        },
        forget(signature: string): void {
            const entries = readEntries();

            if (!(signature in entries)) {
                return;
            }

            delete entries[signature];
            writeEntries(entries);
        },
        listSignatures(): string[] {
            return Object.keys(readEntries());
        }
    };
}

function getBrowserStorage(): ImportMappingStorage | undefined {
    if (typeof window === 'undefined' || !window.localStorage) {
        return undefined;
    }

    return window.localStorage;
}

let defaultStore: ImportMappingStore | undefined;

/**
 * Returns the mapping store backed by the browser, or an inert one when storage is
 * unavailable so importing still works in private windows.
 */
export function getImportMappingStore(): ImportMappingStore {
    if (!defaultStore) {
        const storage = getBrowserStorage();

        defaultStore = storage
            ? createImportMappingStore(storage)
            : createImportMappingStore({
                getItem: () => null,
                setItem: () => undefined,
                removeItem: () => undefined
            });
    }

    return defaultStore;
}
