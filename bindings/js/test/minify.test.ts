import { describe, it, expect } from 'vitest';
import { minify } from '@tdewolff/minify';

describe('minify', () => {
    it('minifies HTML while keeping document tags', async () => {
        const input = "<html><span style=\"color:#ff0000;\">A  phrase</span>";
        const expected = "<html><span style=color:red>A phrase</span>";

        const output = await minify({
            data: input,
            type: 'text/html',
            htmlKeepDocumentTags: true,
        });

        expect(output).toBe(expected);
    });

    it('trims redundant whitespace in an HTML document', async () => {
        const input = "<!doctype html><html><body>  hi  </body></html>";
        const expected = "<!doctype html><html><body>hi</body></html>";

        const output = await minify({
            data: input,
            type: 'text/html',
            htmlKeepDocumentTags: true,
        });

        expect(output).toBe(expected);
    });

    it('minifies CSS declarations', async () => {
        const input = "body { margin: 0px; color: red; }";
        const expected = "body{margin:0;color:red}";

        const output = await minify({
            data: input,
            type: 'text/css',
        });

        expect(output).toBe(expected);
    });

    it('minifies JavaScript', async () => {
        const input = "function add(a, b) { return a + b; }";
        const expected = "function add(e,t){return e+t}";

        const output = await minify({
            data: input,
            type: 'application/javascript',
        });

        expect(output).toBe(expected);
    });

    it('handles concurrent minify calls', async () => {
        const jobs = [
            {
                options: {
                    data: "<html>  <body>ok</body>  </html>",
                    type: 'text/html',
                    htmlKeepDocumentTags: true,
                },
                expected: "<html><body>ok</body></html>",
            },
            {
                options: {
                    data: "body { padding: 0px; }",
                    type: 'text/css',
                },
                expected: "body{padding:0}",
            },
            {
                options: {
                    data: "(() => { return 1 + 2; })();",
                    type: 'application/javascript',
                },
                expected: "(()=>1+2)()",
            },
        ] as const;

        const outputs = await Promise.all(jobs.map((job) => minify(job.options)));

        jobs.forEach((job, index) => {
            expect(outputs[index]).toBe(job.expected);
        });
    });
});
