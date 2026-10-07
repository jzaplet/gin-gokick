import { describe, expect, it } from 'vitest';
import { deviceName } from '@/app/Auth/Device/deviceName';

describe('the device name', () => {
    it.each([
        [
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 '
            + 'Safari/537.36',
            'Chrome · Windows',
        ],
        [
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 '
            + 'Safari/537.36 Edg/129.0.0.0',
            'Edge · Windows',
        ],
        [
            'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 '
            + 'Safari/605.1.15',
            'Safari · macOS',
        ],
        [
            'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Mobile '
            + 'Safari/537.36',
            'Chrome · Android',
        ],
        [
            'Mozilla/5.0 (iPad; CPU OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/129.0 '
            + 'Mobile/15E148 Safari/604.1',
            'Chrome · iPadOS',
        ],
        [
            'Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0',
            'Firefox · Linux',
        ],
        [
            'Mozilla/5.0 (Linux; Android 14; SM-S921B) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/26.0 '
            + 'Chrome/122.0.0.0 Mobile Safari/537.36',
            'Samsung Internet · Android',
        ],
        [
            'curl/8.7.1',
            null,
        ],
        [
            '',
            null,
        ],
    ])('names %s', (userAgent, name) => {
        expect(deviceName(userAgent)).toBe(name);
    });
});
