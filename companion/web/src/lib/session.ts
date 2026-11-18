// Pairing: the Mac's link carries the token in the URL fragment (#k=...).
// It is kept in this browser; the page then removes it from the address bar
// (linkInAddressBar + replaceState), so it does not end up in screenshots, history or shared links.
const KEY = 'ts4access.token';

function read(): string | null {
	try {
		return localStorage.getItem(KEY);
	} catch {
		return null;
	}
}

/** Takes the token from the link if present, otherwise the saved one. */
export function loadToken(): string | null {
	const match = /(?:^#|&)k=([A-Za-z0-9_-]{20,})/.exec(location.hash);
	const token = match ? match[1] : read();
	if (match) {
		try {
			localStorage.setItem(KEY, match[1]);
		} catch {
			// Private browsing: the token lives for this page only.
		}
	}
	if (token) rememberForInstall(token);
	return token;
}

/**
 * iOS gives an app added to the Home Screen its own storage, without the key
 * Safari kept. This cookie lets the bridge put the key in the install's start
 * address (server/manifest.go); it is only sent to this same site.
 */
function rememberForInstall(token: string): void {
	try {
		const secure = location.protocol === 'https:' ? '; Secure' : '';
		document.cookie = `ts4k=${token}; Path=/; Max-Age=31536000; SameSite=Strict${secure}`;
	} catch {
		// cookies off: install from a scanned link still works in Safari
	}
}

export function forgetToken(): void {
	try {
		document.cookie = 'ts4k=; Path=/; Max-Age=0';
		localStorage.removeItem(KEY);
	} catch {
		// nothing stored
	}
}

/** True while the pairing link's token is still in the address bar. */
export function linkInAddressBar(): boolean {
	return /(?:^#|&)k=/.test(location.hash);
}
