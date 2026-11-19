# Accessibility Audit — SimSight phone companion

**Verdict:** Can a blind VoiceOver user follow and play the game from the phone app independently? **Yes, with the fixes below applied; real-device VoiceOver testing still pending.**

**Summary:** 0 blockers · 3 serious (fixed) · 2 moderate (1 fixed, 1 open) · 1 minor (fixed). Automated: axe-core on 13 screens and states at iPhone size. Screen-reader judgement from the accessibility tree and source, 2026-11-18. Focus: blindness / screen-reader usability, WCAG 2.2 AA. No human VoiceOver session yet: see the checklist at the end and `docs/mac/TESTING.md`.

## Test environment & methodology

| | |
|---|---|
| **Platforms / AT** | Chromium (Playwright), iPhone-size viewport 390×844, touch; accessibility tree snapshots. Real iPhone + VoiceOver: not yet. |
| **Viewports** | 390×844 (mobile) |
| **Language(s) tested** | English UI; language switcher (fr/en/es) present, sets `<html lang>` (`src/lib/i18n/locale.svelte.ts`). |
| **Scope** | Live, History, Play (list, open category, search, interaction menu), game question card (recipe picker), House (list, room), Sims (with Goals/Skills/Relationships open), Settings, Help. Demo mode (fake household). |
| **Automated tools** | axe-core (rules from the a11y-audit skill's WCAG 2.2 AA table), multi-screen runner with the app's CSP bypassed for the test only. |
| **Human test** | Pending (moderated VoiceOver session with the player). |

## Serious

### 1. Six identical "Other ways" buttons in Needs attention — fixed

- **What she hears (before):** browsing buttons with the rotor: "Other ways, button" ×6, no hint which need each belongs to.
- **Why it fails:** WCAG 2.4.6 Headings and Labels / 4.1.2 Name, Role, Value: names not meaningful out of context.
- **Where:** `companion/web/src/lib/components/AttentionCard.svelte` (item markup).
- **Fix (applied):** each line has an id; every button points to it with `aria-describedby`. After: "Other ways, button. Bella: Hunger 0%, critical."

### 2. "Control Mortimer" repeated once per low need — fixed

- **What she hears (before):** "Mortimer: Hunger 1%, critical. Control Mortimer. Mortimer: Bladder 0%, critical. Control Mortimer…"
- **Why it fails:** WCAG 2.4.6 / 1.3.1: redundant structure, longer swipe path.
- **Where:** `AttentionCard.svelte` (needs loop).
- **Fix (applied):** one line per Sim you do not control, listing all its low needs, with one Control button.

### 3. Long action queue pushes the rest of Play away; duplicate Cancel names — fixed

- **What she hears (before):** a dozen "Cancel Tell Joke / Cancel Take Shower" before reaching You and the search; two identical "Cancel Take Shower".
- **Why it fails:** WCAG 2.4.6 (ambiguous labels); usability of 2.4.1 Bypass Blocks within the screen.
- **Where:** `companion/web/src/lib/components/PlayView.svelte` (queue section).
- **Fix (applied):** running action + next 3, then "Show all N actions" (`aria-expanded`); items numbered ("2. Take Shower · waiting") and each Cancel button described by its item.

## Moderate

### 4. Same page title on every screen — fixed

- **What she hears (before):** "The Sims, live" when returning to the app, whatever the screen.
- **Why it fails:** WCAG 2.4.2 Page Titled (SPA screens).
- **Where:** `companion/web/src/routes/+page.svelte` (`<svelte:head>`).
- **Fix (applied):** `<title>{screen} · The Sims, live</title>`.

### 5. English game names read with the French/Spanish voice — open

- **What she hears:** with the app in French, "Quick Meal", "Grab a Plate", object names spoken with French pronunciation.
- **Why it fails:** WCAG 3.1.2 Language of Parts (game text is in the game's language, not the page's).
- **Where:** game-provided strings in `PlayView.svelte`, `InteractionMenu.svelte`, `HouseView.svelte`, `DialogCard.svelte`, `AttentionCard.svelte`.
- **Fix (proposed):** send the game's language in the bridge's Hello, and wrap game-provided names in `<span lang={gameLang}>` when it differs from the app language.

## Minor

### 6. "On" of the Mac voice switch below contrast — fixed

- **Why it fails:** WCAG 1.4.3 Contrast (Minimum): 4.25:1 (accent on light surface), needs 4.5:1. Low-vision concern only; the switch state is spoken correctly ("Mac voice, switch, on").
- **Where:** `companion/web/src/lib/components/VoiceToggle.svelte`.
- **Fix (applied):** text colour instead of accent, bold.

Preventive (applied): `scroll-padding-top` in `companion/web/src/app.css` so the sticky top bar never covers the focused element (WCAG 2.4.11).

## What's already done well

- One `<main>` per screen, a labelled `<nav>` first in reading order, an `<h1>` per screen that receives focus on tab change, `<h2>`/`<h3>` sections: navigable with the Headings rotor.
- Native controls everywhere (`<button>`, radios, checkboxes, `<details>`), no div-buttons; states exposed (`aria-pressed`, `aria-expanded`, `role="switch"` with `aria-checked`).
- Busy buttons use `aria-disabled` instead of `disabled`, so VoiceOver keeps its place.
- A polite live region announces results, game questions and events; the game question card is announced and appears on every screen.
- Unavailable choices say why ("Lobster Thermidor, unavailable: Requires Cooking level 9").
- Search is a real `role="search"` form with a visible label; the keyboard's Search key closes the keyboard and moves to the results.
- Icons are drawn SVG with `aria-hidden`, never emoji in accessible names.

## Screen-reader test session — evidence

Pending: run the moderated session with `docs/mac/TESTING.md` (and the checklist below) on an iPhone with VoiceOver, then record results here.

## Appendix A — automated results (axe-core, 13 screens)

Before fixes: 1 violation (`color-contrast`, criterion 1.4.3, Live). After fixes: 0 violations on all 13 screens. `incomplete` (needs a human): `color-contrast` on 1–3 nodes in Play (open category, search) and Settings, text over translucent surfaces; verify visually.

## Appendix B — WCAG 2.2 AA coverage (55 criteria)

Verdicts: **pass** (evidence), **fixed** (failed, corrected, re-scanned), **open**, **n/a** (no such content), **manual** (needs the human session).

| Criterion | Bucket | Verdict | Evidence |
|---|---|---|---|
| [1.1.1 Non-text Content](https://www.w3.org/WAI/WCAG22/Understanding/non-text-content) | mixed | pass | Only decorative SVG icons, `aria-hidden`; manifest icons not in the page. |
| [1.2.1 Audio-only and Video-only](https://www.w3.org/WAI/WCAG22/Understanding/audio-only-and-video-only-prerecorded) | ai | n/a | No media. |
| [1.2.2 Captions (Prerecorded)](https://www.w3.org/WAI/WCAG22/Understanding/captions-prerecorded) | mixed | n/a | No media. |
| [1.2.3 Audio Description or Media Alternative](https://www.w3.org/WAI/WCAG22/Understanding/audio-description-or-media-alternative-prerecorded) | ai | n/a | No media. |
| [1.2.4 Captions (Live)](https://www.w3.org/WAI/WCAG22/Understanding/captions-live) | ai | n/a | No media. |
| [1.2.5 Audio Description (Prerecorded)](https://www.w3.org/WAI/WCAG22/Understanding/audio-description-prerecorded) | ai | n/a | No media. |
| [1.3.1 Info and Relationships](https://www.w3.org/WAI/WCAG22/Understanding/info-and-relationships) | mixed | fixed | Headings, lists, fieldsets with legends; repeated per-need lines merged (finding 2). |
| [1.3.2 Meaningful Sequence](https://www.w3.org/WAI/WCAG22/Understanding/meaningful-sequence) | ai | pass | DOM order = visual order (top bar, tabs, main); no CSS reordering. |
| [1.3.3 Sensory Characteristics](https://www.w3.org/WAI/WCAG22/Understanding/sensory-characteristics) | ai | pass | Help text names screens/buttons, not shapes or positions; directions are compass words from the game. |
| [1.3.4 Orientation](https://www.w3.org/WAI/WCAG22/Understanding/orientation) | mixed | pass | No orientation lock (CSS or manifest). |
| [1.3.5 Identify Input Purpose](https://www.w3.org/WAI/WCAG22/Understanding/identify-input-purpose) | auto | pass | axe clean; no personal-data fields. |
| [1.4.1 Use of Color](https://www.w3.org/WAI/WCAG22/Understanding/use-of-color) | mixed | pass | Pressed/checked/urgent states also in text or ARIA state, not colour alone. |
| [1.4.2 Audio Control](https://www.w3.org/WAI/WCAG22/Understanding/audio-control) | auto | pass | No page audio. |
| [1.4.3 Contrast (Minimum)](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum) | auto | fixed | Finding 6; re-scan clean. |
| [1.4.4 Resize Text](https://www.w3.org/WAI/WCAG22/Understanding/resize-text) | manual | manual | Checklist item 1. |
| [1.4.5 Images of Text](https://www.w3.org/WAI/WCAG22/Understanding/images-of-text) | ai | pass | No images of text. |
| [1.4.10 Reflow](https://www.w3.org/WAI/WCAG22/Understanding/reflow) | manual | manual | Checklist item 2 (scroll width = viewport on all tabs at 390 px measured). |
| [1.4.11 Non-text Contrast](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast) | ai | pass | Buttons have 2 px borders on contrasting surfaces; checkboxes/radios use the accent colour at 1.5 rem. |
| [1.4.12 Text Spacing](https://www.w3.org/WAI/WCAG22/Understanding/text-spacing) | manual | manual | Checklist item 3. |
| [1.4.13 Content on Hover or Focus](https://www.w3.org/WAI/WCAG22/Understanding/content-on-hover-or-focus) | manual | manual | Checklist item 4 (no hover content found in source). |
| [2.1.1 Keyboard](https://www.w3.org/WAI/WCAG22/Understanding/keyboard) | manual | manual | Checklist item 5; all controls native. |
| [2.1.2 No Keyboard Trap](https://www.w3.org/WAI/WCAG22/Understanding/no-keyboard-trap) | manual | manual | Checklist item 6. |
| [2.1.4 Character Key Shortcuts](https://www.w3.org/WAI/WCAG22/Understanding/character-key-shortcuts) | ai | pass | No single-key shortcuts in the web app. |
| [2.2.1 Timing Adjustable](https://www.w3.org/WAI/WCAG22/Understanding/timing-adjustable) | ai | pass | No time limits; game questions wait for an answer. |
| [2.2.2 Pause, Stop, Hide](https://www.w3.org/WAI/WCAG22/Understanding/pause-stop-hide) | ai | pass | Live feed updates are user content; announcements are opt-in per kind (Settings). |
| [2.3.1 Three Flashes](https://www.w3.org/WAI/WCAG22/Understanding/three-flashes-or-below-threshold) | ai | pass | No flashing content. |
| [2.4.1 Bypass Blocks](https://www.w3.org/WAI/WCAG22/Understanding/bypass-blocks) | auto | pass | Landmarks; focus moves to the screen's h1 on tab change; queue capped (finding 3). |
| [2.4.2 Page Titled](https://www.w3.org/WAI/WCAG22/Understanding/page-titled) | mixed | fixed | Finding 4. |
| [2.4.3 Focus Order](https://www.w3.org/WAI/WCAG22/Understanding/focus-order) | manual | manual | Checklist item 7; focus returns to the opened item after Back (source). |
| [2.4.4 Link Purpose (In Context)](https://www.w3.org/WAI/WCAG22/Understanding/link-purpose-in-context) | ai | pass | Only link: "Open ChatGPT" (Help), self-describing. |
| [2.4.5 Multiple Ways](https://www.w3.org/WAI/WCAG22/Understanding/multiple-ways) | ai | pass | Tabs + Help contents + search in Play. |
| [2.4.6 Headings and Labels](https://www.w3.org/WAI/WCAG22/Understanding/headings-and-labels) | ai | fixed | Findings 1 and 3. |
| [2.4.7 Focus Visible](https://www.w3.org/WAI/WCAG22/Understanding/focus-visible) | manual | manual | Checklist item 8 (`:focus-visible` outline in app.css). |
| [2.4.11 Focus Not Obscured (Minimum)](https://www.w3.org/WAI/WCAG22/Understanding/focus-not-obscured-minimum) | manual | manual | Checklist item 9; preventive `scroll-padding-top` applied. |
| [2.5.1 Pointer Gestures](https://www.w3.org/WAI/WCAG22/Understanding/pointer-gestures) | ai | pass | Single taps only. |
| [2.5.2 Pointer Cancellation](https://www.w3.org/WAI/WCAG22/Understanding/pointer-cancellation) | ai | pass | Native buttons act on click (up event). |
| [2.5.3 Label in Name](https://www.w3.org/WAI/WCAG22/Understanding/label-in-name) | auto | pass | axe clean; no `aria-label` overriding visible text (switch name "Mac voice" ⊂ visible "Mac voice On"). |
| [2.5.4 Motion Actuation](https://www.w3.org/WAI/WCAG22/Understanding/motion-actuation) | ai | pass | No motion input. |
| [2.5.7 Dragging Movements](https://www.w3.org/WAI/WCAG22/Understanding/dragging-movements) | ai | pass | No dragging. |
| [2.5.8 Target Size (Minimum)](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum) | auto | pass | axe clean; `--tap` min-height on buttons and inputs. |
| [3.1.1 Language of Page](https://www.w3.org/WAI/WCAG22/Understanding/language-of-page) | auto | pass | `<html lang>` set and updated by the language switcher. |
| [3.1.2 Language of Parts](https://www.w3.org/WAI/WCAG22/Understanding/language-of-parts) | mixed | open | Finding 5. |
| [3.2.1 On Focus](https://www.w3.org/WAI/WCAG22/Understanding/on-focus) | manual | manual | Checklist item 10. |
| [3.2.2 On Input](https://www.w3.org/WAI/WCAG22/Understanding/on-input) | manual | manual | Checklist item 11 (Settings apply on change, announced "Done"). |
| [3.2.3 Consistent Navigation](https://www.w3.org/WAI/WCAG22/Understanding/consistent-navigation) | manual | manual | Checklist item 12 (same top bar on every screen). |
| [3.2.4 Consistent Identification](https://www.w3.org/WAI/WCAG22/Understanding/consistent-identification) | manual | manual | Checklist item 13. |
| [3.2.6 Consistent Help](https://www.w3.org/WAI/WCAG22/Understanding/consistent-help) | manual | manual | Checklist item 14 (Help tab always in the same place). |
| [3.3.1 Error Identification](https://www.w3.org/WAI/WCAG22/Understanding/error-identification) | manual | manual | Checklist item 15 (failures announced with a reason). |
| [3.3.2 Labels or Instructions](https://www.w3.org/WAI/WCAG22/Understanding/labels-or-instructions) | mixed | pass | Visible labels on search, text answer, selects; pickers have a legend ("Choose one, then OK."). |
| [3.3.3 Error Suggestion](https://www.w3.org/WAI/WCAG22/Understanding/error-suggestion) | manual | manual | Checklist item 16. |
| [3.3.4 Error Prevention](https://www.w3.org/WAI/WCAG22/Understanding/error-prevention-legal-financial-data) | manual | manual | Checklist item 17 (paying bills is a game action, reversible in game terms). |
| [3.3.7 Redundant Entry](https://www.w3.org/WAI/WCAG22/Understanding/redundant-entry) | manual | manual | Checklist item 18. |
| [3.3.8 Accessible Authentication (Minimum)](https://www.w3.org/WAI/WCAG22/Understanding/accessible-authentication-minimum) | ai | pass | Pairing by QR link, no password or cognitive test. |
| [4.1.2 Name, Role, Value](https://www.w3.org/WAI/WCAG22/Understanding/name-role-value) | mixed | fixed | Native controls with states; finding 1 fixed. |
| [4.1.3 Status Messages](https://www.w3.org/WAI/WCAG22/Understanding/status-messages) | ai | pass | Polite live region (`Announcer.svelte`) for results, questions and events. |

Count: 55 criteria = pass 26 + fixed 5 + open 1 + n/a 5 + manual 18; each criterion once.

## Appendix C — manual checklist (about 30 minutes, iPhone + VoiceOver)

- [ ] 1. Text Size at the largest setting: nothing cut off (1.4.4) — 3 min
- [ ] 2. Every tab scrolls only vertically (1.4.10) — 2 min
- [ ] 3. Larger text spacing (if a tool is available) keeps text readable (1.4.12) — 2 min
- [ ] 4. No content appears on hover or focus only (1.4.13) — 1 min
- [ ] 5. With a Bluetooth keyboard: every control reachable and operable (2.1.1) — 4 min
- [ ] 6. No keyboard trap in menus, questions, Help (2.1.2) — 2 min
- [ ] 7. VoiceOver swipe order follows the screen; Back returns to the item opened (2.4.3) — 3 min
- [ ] 8. Keyboard focus visible (2.4.7) — 1 min
- [ ] 9. Focused item never hidden under the top bar while scrolling (2.4.11) — 2 min
- [ ] 10. Focusing a control never changes screen by itself (3.2.1) — 1 min
- [ ] 11. Changing a setting does not move you elsewhere (3.2.2) — 1 min
- [ ] 12. Top bar the same on every screen (3.2.3) — 1 min
- [ ] 13. Same function, same name everywhere (3.2.4) — 1 min
- [ ] 14. Help reachable from the same place everywhere (3.2.6) — 1 min
- [ ] 15. A failed action is announced with its reason (3.3.1) — 2 min
- [ ] 16. Failure messages suggest what to do (3.3.3) — 1 min
- [ ] 17. Paying bills / cancelling actions are deliberate (3.3.4) — 1 min
- [ ] 18. Nothing asks to retype information (3.3.7) — 1 min
