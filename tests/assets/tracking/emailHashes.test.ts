import { describe, expect, it } from 'vitest';
import { googleEmailHash, metaEmailHash } from '@/shared/Tracking/UserData/emailHashes';

describe('email hashes', () => {
    it.each([
        [
            ' Jan.Novak@Gmail.com ',
            '005ed88a887dbd4c32e8d7ca3665981df82512b3a7fbf451328ef0a46835d803',
        ],
        [
            'Jan.Novak@GoogleMail.com',
            'c0656a02384b99c29e94152bce05b466ca5d5951733f6abfa7b35807ad8e5b9f',
        ],
        [
            'Eva.Nova@Example.test',
            'fa763aa862f859e132831a3cd916814407fd7653936e30835a705411c8f9d28a',
        ],
    ])('hash %s for Google without the dots of a Gmail name', async (email, hash) => {
        expect(await googleEmailHash(email)).toBe(hash);
    });

    it('hashes the trimmed lowercase email for Meta and keeps its dots', async () => {
        expect(await metaEmailHash(' Jan.Novak@Gmail.com ')).toBe(
            '7d4f91fdbfb1c424c7b2b762fb811d6493d9290dc167732be776e70c4ae72c2d',
        );
    });
});
