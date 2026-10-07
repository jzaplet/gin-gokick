import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiDownload } from '@/shared/Fetch';

const file = (status: number, headers: Record<string, string> = {}): Response =>
    new Response(new Blob(['file-bytes']), {
        status,
        headers,
    });

describe('apiDownload', () => {
    const createObjectUrl = vi.fn((): string => 'blob:mock-url');
    const revokeObjectUrl = vi.fn();
    const click = vi.fn();

    beforeEach(() => {
        createObjectUrl.mockClear();
        revokeObjectUrl.mockClear();
        click.mockClear();
        URL.createObjectURL = createObjectUrl;
        URL.revokeObjectURL = revokeObjectUrl;
        vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(click);
    });

    it('falls back to the given filename and downloads through a temporary link', async () => {
        const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(file(200));

        expect(await apiDownload('/api/export', 'fallback.pdf')).toEqual({
            success: true,
            status: 200,
            filename: 'fallback.pdf',
        });
        expect(fetchSpy.mock.calls[0]?.[1]?.credentials).toBe('same-origin');
        expect(createObjectUrl).toHaveBeenCalledOnce();
        expect(click).toHaveBeenCalledOnce();
        expect(revokeObjectUrl).toHaveBeenCalledWith('blob:mock-url');
    });

    it('downloads nothing on an error status', async () => {
        vi.spyOn(globalThis, 'fetch').mockResolvedValue(file(403));

        expect(await apiDownload('/api/export', 'fallback.csv')).toEqual({
            success: false,
            status: 403,
            filename: null,
        });
        expect(click).not.toHaveBeenCalled();
    });

    it('turns a network failure into status 0', async () => {
        vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));

        expect(await apiDownload('/api/export', 'fallback.csv')).toEqual({
            success: false,
            status: 0,
            filename: null,
        });
    });
});
