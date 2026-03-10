import logger from '@/lib/logger.ts';

/**
 * Standard error handler for store API calls.
 * Replaces ~96 duplicated catch blocks across all store files.
 *
 * Usage in a store action:
 *   services.someApi(req).then(response => { ... }).catch(handleStoreError('action name', 'User-facing error message', reject));
 */
export function handleStoreError(
    operation: string,
    fallbackMessage: string,
    reject: (reason?: unknown) => void
): (error: unknown) => void {
    return (error: unknown) => {
        logger.error(`failed to ${operation}`, error);

        const err = error as Record<string, unknown>;

        if (err && err['processed']) {
            reject(error);
        } else if (err && err['response'] && (err['response'] as Record<string, unknown>)['data'] && ((err['response'] as Record<string, unknown>)['data'] as Record<string, unknown>)['errorMessage']) {
            reject({ error: (err['response'] as Record<string, unknown>)['data'] });
        } else {
            reject({ message: fallbackMessage });
        }
    };
}
