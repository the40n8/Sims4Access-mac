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
	if (match) {
		try {
			localStorage.setItem(KEY, match[1]);
		} catch {
			// Private browsing: the token lives for this page only.
		}
		return match[1];
	}
	return read();
}

export function forgetToken(): void {
	try {
		localStorage.removeItem(KEY);
	} catch {
		// nothing stored
	}
}

/** True while the pairing link's token is still in the address bar. */
export function linkInAddressBar(): boolean {
	return /(?:^#|&)k=/.test(location.hash);
}
