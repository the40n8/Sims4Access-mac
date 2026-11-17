// The interface language: which one is active and how to change it (remembered in this browser).
import { getLocale, type Locale, locales, setLocale } from './paraglide/runtime.js';

export type { Locale } from './paraglide/runtime.js';

/** Each language in its own language, for the picker. */
export const LANGUAGES: readonly { readonly locale: Locale; readonly name: string }[] = [
	{ locale: 'fr', name: 'Français' },
	{ locale: 'en', name: 'English' },
	{ locale: 'es', name: 'Español' }
];

export function isLocale(value: string): value is Locale {
	return locales.some((l) => l === value);
}

/** The active language; markup keyed on it re-renders in the new language. */
export const i18n = $state({ locale: getLocale() });

/** Switches language without reloading: localStorage, <html lang>, re-render. */
export function changeLocale(locale: Locale): void {
	if (locale === i18n.locale) return;
	void setLocale(locale, { reload: false });
	i18n.locale = locale;
	document.documentElement.lang = locale;
}

export function applyDocumentLocale(): void {
	document.documentElement.lang = i18n.locale;
}
