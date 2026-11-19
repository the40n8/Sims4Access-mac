// Saved before the project was renamed to SimSight: the
// pairing key, the games list and the preferences move to their new names
// once, so a phone that was paired stays paired. Runs before anything reads
// the storage (first import of the layout).
const OLD = 'ts4' + 'access.'; // split: mac/rebrand.py must not rename it
const NEW = 'simsight.';

try {
	const old = Object.keys(localStorage).filter((k) => k.startsWith(OLD));
	for (const key of old) {
		const fresh = NEW + key.slice(OLD.length);
		const value = localStorage.getItem(key);
		if (value !== null && localStorage.getItem(fresh) === null) localStorage.setItem(fresh, value);
		localStorage.removeItem(key);
	}
} catch {
	// no storage: nothing to move
}

export {};
