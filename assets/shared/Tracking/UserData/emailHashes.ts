const gmailDomains = [
    'gmail.com',
    'googlemail.com',
];

const sha256 = async (text: string): Promise<string> => {
    const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(text));

    return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
};

const normalized = (email: string): string => email.trim().toLowerCase();

const withoutGmailDots = (address: string): string => {
    const at = address.lastIndexOf('@');
    const name = address.slice(0, at);
    const domain = address.slice(at + 1);

    return gmailDomains.includes(domain) ? `${name.replaceAll('.', '')}@${domain}` : address;
};

export const googleEmailHash = (email: string): Promise<string> => sha256(withoutGmailDots(normalized(email)));

export const metaEmailHash = (email: string): Promise<string> => sha256(normalized(email));
