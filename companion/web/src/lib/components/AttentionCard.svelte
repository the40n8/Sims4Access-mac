<script lang="ts">
	import { tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { live } from '$lib/live.svelte.js';
	import { levelName, needName, reasonText, simName } from '$lib/text.js';
	import { NeedLevel, type Problem, type Sim, type Target } from '$lib/gen/ts4access/v1/game_pb.js';
	import { NEED_HELP, PROBLEM_HELP, helpers, runQuick, type Help, type Quick } from '$lib/quick.js';
	import { objectName } from '$lib/targets.js';
	import { announce } from './Announcer.svelte';
	import InteractionMenu from './InteractionMenu.svelte';

	// What needs doing now, most urgent first, each with the game's own
	// actions one tap away: low needs, fire, broken objects, puddles, spoiled food.
	const game = $derived(live.state);
	const me = $derived(game?.sims.find((s) => s.active));

	type Item = { id: string; text: string; help: Help; targetId?: string; sim?: Sim; urgent: boolean };

	const PROBLEM_TEXT: Record<string, (p: Problem) => string> = {
		FIRE: (p) => m.pb_fire({ name: p.name }),
		BROKEN: (p) => m.pb_broken({ name: p.name }),
		PUDDLE: () => m.pb_puddle(),
		SPOILED: (p) => m.pb_spoiled({ name: p.name })
	};

	const items = $derived.by(() => {
		const out: Item[] = [];
		for (const p of game?.problems ?? []) {
			const help = PROBLEM_HELP[p.kind];
			if (!help) continue;
			const where = p.distance > 0 ? `, ${m.meters_away({ meters: p.distance })}` : '';
			out.push({
				id: `p-${p.kind}-${p.targetId}`,
				text: (PROBLEM_TEXT[p.kind] ?? ((x: Problem) => x.name))(p) + where,
				help,
				targetId: p.targetId,
				urgent: p.kind === 'FIRE'
			});
		}
		for (const sim of game?.sims ?? []) {
			if (!sim.here) continue;
			for (const need of sim.needs) {
				if (need.level !== NeedLevel.LOW && need.level !== NeedLevel.CRITICAL) continue;
				const help = NEED_HELP[need.key];
				if (!help) continue;
				out.push({
					id: `n-${sim.id}-${need.key}`,
					text: m.att_need({
						sim: simName(sim),
						need: needName(need.key),
						percent: need.percent,
						level: levelName(need.level)
					}),
					help,
					sim,
					urgent: need.level === NeedLevel.CRITICAL
				});
			}
		}
		// Fire and critical needs first; then the rest as found.
		return out.sort((a, b) => Number(b.urgent) - Number(a.urgent)).slice(0, 6);
	});

	// Objects near the Sim, for "nearest" actions and "Other ways".
	let targets = $state<Target[]>([]);
	let loadedAt = 0;
	async function loadTargets(force = false) {
		if (!force && Date.now() - loadedAt < 60_000) return;
		loadedAt = Date.now();
		try {
			const res = await client.listTargets({});
			if (res.outcome?.ok) targets = res.targets;
		} catch {
			// keep the last list
		}
	}
	$effect(() => {
		if (items.some((i) => i.sim)) void loadTargets();
	});
	const phone = $derived(targets.find((t) => t.group === 'PHONE'));

	let busy = $state(false);
	let selected = $state<Target | null>(null);
	let others = $state<string | null>(null);
	let returnFocus = '';

	async function quick(item: Item, q: Quick, buttonId: string) {
		if (busy) return;
		busy = true;
		returnFocus = buttonId;
		let targetId = item.targetId;
		if (q.on === 'phone') {
			await loadTargets();
			targetId = phone?.id ?? 'phone';
		} else if (q.on === 'nearest') {
			await loadTargets();
			// The nearest object whose menu has this action.
			const candidates = helpers(targets, item.help.objects);
			for (const c of candidates) {
				const out = await runQuick(c.id, q.action);
				if (out.kind === 'queued') {
					busy = false;
					announce(m.queued({ name: `${out.name} (${c.name})` }));
					return;
				}
				if (out.kind === 'unavailable' && candidates.length === 1) {
					busy = false;
					announce(m.unavailable_because({ name: out.name, reason: out.reason }));
					return;
				}
			}
			busy = false;
			announce(m.qa_none());
			others = item.id;
			return;
		}
		if (!targetId) {
			busy = false;
			return;
		}
		const out = await runQuick(targetId, q.action);
		busy = false;
		if (out.kind === 'queued') announce(m.queued({ name: out.name }));
		else if (out.kind === 'unavailable')
			announce(m.unavailable_because({ name: out.name, reason: out.reason }));
		else if (out.kind === 'failed') announce(m.failed({ reason: reasonText(out.reason) }));
		else {
			// Not in the menu under that name: show the whole menu instead.
			announce(m.qa_open_menu());
			const t = targets.find((x) => x.id === targetId) ?? fallbackTarget(item, targetId);
			await openMenu(t, buttonId);
		}
	}

	function fallbackTarget(item: Item, id: string): Target {
		const problem = game?.problems.find((p) => p.targetId === id);
		return { id, name: problem?.name ?? item.text, group: id === 'phone' ? 'PHONE' : 'MISC' } as Target;
	}

	async function openMenu(t: Target, from: string) {
		returnFocus = from;
		selected = t;
	}
	async function closeMenu() {
		selected = null;
		await tick();
		document.getElementById(returnFocus)?.focus();
	}

	async function control(sim: Sim, buttonId: string) {
		if (busy) return;
		busy = true;
		const out = await run(() => client.selectSim({ simId: sim.id }));
		busy = false;
		announce(out.ok ? m.control_now({ name: simName(sim) }) : m.failed({ reason: reasonText(out.reason) }));
		await tick();
		document.getElementById(buttonId)?.focus();
	}

	async function showOthers(item: Item) {
		others = others === item.id ? null : item.id;
		if (others) await loadTargets(true);
	}
</script>

{#if selected}
	<InteractionMenu target={selected} onBack={closeMenu} />
{:else if items.length}
	<section class="card stack attention" aria-labelledby="attention-title">
		<h2 id="attention-title">{m.attention_title()}</h2>
		<ul class="stack">
			{#each items as item (item.id)}
				<li class="stack item" class:urgent={item.urgent}>
					<p>{item.text}</p>
					<div class="buttons">
						{#if item.sim && !item.sim.active}
							<button
								type="button"
								id={`${item.id}-control`}
								aria-disabled={busy}
								onclick={() => control(item.sim!, `${item.id}-control`)}
							>
								{m.control_sim({ name: simName(item.sim) })}
							</button>
						{:else}
							{#each item.help.quick as q, i (i)}
								<button
									type="button"
									id={`${item.id}-q${i}`}
									aria-disabled={busy}
									onclick={() => quick(item, q, `${item.id}-q${i}`)}>{q.label()}</button
								>
							{/each}
							{#if item.help.objects || item.help.phone}
								<button
									type="button"
									id={`${item.id}-others`}
									aria-expanded={others === item.id}
									aria-disabled={busy}
									onclick={() => showOthers(item)}>{m.qa_other_ways()}</button
								>
							{/if}
						{/if}
					</div>
					{#if others === item.id}
						<ul class="stack">
							{#each helpers(targets, item.help.objects) as t (t.id)}
								<li>
									<button
										type="button"
										class="row"
										id={`${item.id}-o-${t.id}`}
										onclick={() => openMenu(t, `${item.id}-o-${t.id}`)}
									>
										{objectName(t)}, {m.meters_away({ meters: t.distance })}
									</button>
								</li>
							{/each}
							{#if item.help.phone && phone}
								<li>
									<button
										type="button"
										class="row"
										id={`${item.id}-o-phone`}
										onclick={() => openMenu(phone!, `${item.id}-o-phone`)}
									>
										{m.target_phone({ name: me ? simName(me) : '' })}
									</button>
								</li>
							{/if}
						</ul>
					{/if}
				</li>
			{/each}
		</ul>
	</section>
{/if}

<style>
	.attention {
		border: 3px solid var(--warn);
	}
	.item {
		padding-bottom: 0.5rem;
		border-bottom: 1px solid var(--line);
	}
	.item:last-child {
		border-bottom: none;
	}
	.item p {
		margin: 0;
		font-weight: 600;
	}
	.urgent p {
		color: var(--bad);
	}
	.buttons {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
		gap: 0.5rem;
	}
	.row {
		width: 100%;
		text-align: start;
	}
</style>
