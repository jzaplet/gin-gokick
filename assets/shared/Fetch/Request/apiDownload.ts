import type { DownloadResult } from '@/shared/Fetch/types/DownloadResult';

const quotedFilename = /filename="?(.+?)"?$/;

const parseFilename = (response: Response, fallback: string): string => {
    const match = response.headers.get('Content-Disposition')?.match(quotedFilename);

    return match?.[1] ?? fallback;
};

const triggerDownload = (blob: Blob, filename: string): void => {
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');

    link.href = url;
    link.download = filename;
    link.click();
    URL.revokeObjectURL(url);
};

export const apiDownload = async (url: string, fallbackFilename: string): Promise<DownloadResult> => {
    const response = await fetch(url, { credentials: 'same-origin' }).catch(() => null);

    if (response === null) {
        return {
            success: false,
            status: 0,
            filename: null,
        };
    }

    if (response.ok === false) {
        return {
            success: false,
            status: response.status,
            filename: null,
        };
    }

    const filename = parseFilename(response, fallbackFilename);

    triggerDownload(await response.blob(), filename);

    return {
        success: true,
        status: response.status,
        filename,
    };
};
