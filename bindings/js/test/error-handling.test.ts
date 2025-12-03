import { describe, it, expect } from 'vitest';
import { minify } from '@tdewolff/minify';

describe('minify error handling', () => {
    it('rejects unknown media types without killing the process', async () => {
        await expect(
            minify({
                data: '<div>oops</div>',
                // @ts-expect-error testing invalid type handling
                type: 'invalid/type',
            }),
        ).rejects.toThrow(/invalid type/i);

        const recovered = await minify({
            data: '<html><body>ok</body></html>',
            type: 'text/html',
            htmlKeepDocumentTags: true,
        });
        expect(recovered).toBe('<html><body>ok</body></html>');
    });

    it('surfaces native parse errors without crashing', async () => {
        await expect(
            minify({
                data: 'function () {',
                type: 'application/javascript',
            }),
        ).rejects.toThrow(/expected identifier/i);

        const recovered = await minify({
            data: 'const n = 1;',
            type: 'application/javascript',
        });
        expect(recovered).toBe('const n=1');
    });
});
