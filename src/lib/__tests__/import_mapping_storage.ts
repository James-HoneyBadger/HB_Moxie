import { describe, expect, it } from '@jest/globals';

import { ImportTransactionDataMapping } from '@/core/import_transaction.ts';
import { createImportMappingStore, getImportMappingSignature } from '@/lib/import_mapping_storage.ts';

class FakeStorage {
    private readonly items = new Map<string, string>();

    public getItem(key: string): string | null {
        return this.items.has(key) ? this.items.get(key)! : null;
    }

    public setItem(key: string, value: string): void {
        this.items.set(key, value);
    }

    public removeItem(key: string): void {
        this.items.delete(key);
    }
}

function createMapping(timeFormat: string): ImportTransactionDataMapping {
    const mapping = ImportTransactionDataMapping.createEmpty();

    (mapping as unknown as { timeFormat: string }).timeFormat = timeFormat;

    return mapping;
}

describe('getImportMappingSignature', () => {
    it('normalizes the first row so the same layout matches again', () => {
        const first = getImportMappingSignature([[' Date ', 'Amount', 'Description ']]);
        const second = getImportMappingSignature([['date', 'amount', 'description']]);

        expect(first).toBeDefined();
        expect(first).toEqual(second);
    });

    it('gives different layouts different signatures', () => {
        expect(getImportMappingSignature([['Date', 'Amount']]))
            .not.toEqual(getImportMappingSignature([['Date', 'Value']]));
    });

    it('returns undefined when there is nothing to identify', () => {
        expect(getImportMappingSignature(undefined)).toBeUndefined();
        expect(getImportMappingSignature([])).toBeUndefined();
        expect(getImportMappingSignature([[]])).toBeUndefined();
        expect(getImportMappingSignature([['Date', '  ']])).toBeUndefined();
    });
});

describe('createImportMappingStore', () => {
    it('round-trips a mapping for a signature', () => {
        const store = createImportMappingStore(new FakeStorage());
        const signature = 'date|amount';

        expect(store.load(signature)).toBeUndefined();

        store.save(signature, createMapping('dd/MM/yyyy'));

        expect(store.load(signature)?.timeFormat).toEqual('dd/MM/yyyy');
    });

    it('keeps mappings for different layouts apart', () => {
        const store = createImportMappingStore(new FakeStorage());

        store.save('a', createMapping('one'));
        store.save('b', createMapping('two'));

        expect(store.load('a')?.timeFormat).toEqual('one');
        expect(store.load('b')?.timeFormat).toEqual('two');
        expect(store.listSignatures().sort()).toEqual(['a', 'b']);
    });

    it('forgets a mapping without touching the others', () => {
        const store = createImportMappingStore(new FakeStorage());

        store.save('a', createMapping('one'));
        store.save('b', createMapping('two'));
        store.forget('a');

        expect(store.load('a')).toBeUndefined();
        expect(store.load('b')?.timeFormat).toEqual('two');
    });

    it('caps how many mappings are kept', () => {
        const store = createImportMappingStore(new FakeStorage());

        for (let index = 0; index < 25; index++) {
            store.save(`signature-${index}`, createMapping(`format-${index}`));
        }

        expect(store.listSignatures()).toHaveLength(20);
        expect(store.load('signature-24')?.timeFormat).toEqual('format-24');
        expect(store.load('signature-0')).toBeUndefined();
    });

    it('ignores corrupt stored data instead of throwing', () => {
        const storage = new FakeStorage();
        storage.setItem('honey_ledger_import_mappings', 'not json');

        const store = createImportMappingStore(storage);

        expect(store.load('anything')).toBeUndefined();
        expect(() => store.save('anything', createMapping('x'))).not.toThrow();
    });
});
