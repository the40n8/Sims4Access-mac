<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { live } from '$lib/live.svelte.js';
	import { awaySentence, duration, levelName, moodName, needName, reasonText, simName } from '$lib/text.js';
	import { ActionState, NeedLevel, type Action, type Sim } from '$lib/gen/ts4access/v1/game_pb.js';
	import { announce } from './Announcer.svelte';

	const sims = $derived(live.state?.inGame ? live.state.sims : []);

	async function cancel(sim: Sim, action: Action) {
		const out = await run(() => client.cancelAction({ simId: sim.id, actionId: action.id }));
		announce(out.ok ? m.done() : m.failed({ reason: reasonText(out.reason) }));
	}
</script>

{#each sims as sim (sim.id)}
	<section class="card stack" aria-labelledby={`sim-${sim.id}`}>
		<h2 id={`sim-${sim.id}`}>
			{simName(sim)}
			{#if sim.active}<span class="muted">· {m.active_badge()}</span>{/if}
		</h2>
		{#if sim.career}
			<p>
				{m.career_line({ career: sim.career })}
				{#if sim.nextShift && !sim.awayAt}<span class="muted">· {m.next_shift({ when: sim.nextShift })}</span
					>{/if}
			</p>
		{/if}
		{#if sim.performance !== undefined}<p>{m.performance_line({ percent: sim.performance })}</p>{/if}
		{#if !sim.here}
			<p class="muted">{awaySentence(sim)}</p>
		{:else}
			<p>{m.mood_line({ mood: moodName(sim.mood) })}</p>

			<h3>{m.needs_title()}</h3>
			<ul class="needs">
				{#each [...sim.needs].sort((a, b) => a.percent - b.percent) as need (need.key)}
					<li class:low={need.level === NeedLevel.LOW} class:critical={need.level === NeedLevel.CRITICAL}>
						<span>
							{m.need_value({
								need: needName(need.key),
								percent: need.percent,
								level: levelName(need.level)
							})}
						</span>
						<span class="bar" aria-hidden="true"><span style:width={`${need.percent}%`}></span></span>
					</li>
				{/each}
			</ul>

			<h3>{m.moodlets_title()}</h3>
			{#if sim.moodlets.length === 0}
				<p class="muted">{m.moodlets_none()}</p>
			{:else}
				<ul class="stack">
					{#each sim.moodlets as moodlet, i (i)}
						<li>
							<strong>{moodlet.name}</strong>
							{#if moodlet.moodKey}· {moodName({ key: moodlet.moodKey, name: '' })}{/if}
							{#if moodlet.minutesLeft >= 0}
								<span class="muted">· {m.minutes_left({ time: duration(moodlet.minutesLeft) })}</span>
							{/if}
							{#if moodlet.description}<p class="muted">{moodlet.description}</p>{/if}
						</li>
					{/each}
				</ul>
			{/if}

			<h3>{m.actions_title()}</h3>
			{#if sim.actions.length === 0}
				<p class="muted">{m.actions_none()}</p>
			{:else}
				<ul class="stack">
					{#each sim.actions as action (action.id)}
						<li class="action">
							<span>
								{action.name}
								<span class="muted">
									· {action.state === ActionState.RUNNING ? m.action_running() : m.action_queued()}
								</span>
							</span>
							{#if action.cancelable && sim.active}
								<button type="button" onclick={() => cancel(sim, action)}>
									{m.cancel_action({ action: action.name })}
								</button>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
		{/if}

		{#if sim.goals && (sim.goals.aspiration || sim.goals.wants.length || sim.goals.workTasks.length || sim.goals.promotionSteps.length)}
			{@const g = sim.goals}
			<details>
				<summary>{m.goals_title()}</summary>
				<div class="stack">
					{#if g.aspiration}
						<p>{m.goal_aspiration({ name: g.aspiration })}</p>
						{#if g.milestone}<p>{m.goal_milestone({ name: g.milestone })}</p>{/if}
						{#if g.milestoneSteps.length}
							<ul class="stack">
								{#each g.milestoneSteps as step, i (i)}
									<li>{step.done ? m.step_done({ name: step.name }) : m.step_todo({ name: step.name })}</li>
								{/each}
							</ul>
						{/if}
					{/if}
					{#if g.wants.length}
						<h4>{m.wants_title()}</h4>
						<ul class="stack">
							{#each g.wants as w, i (i)}
								<li>{w.fear ? m.fear_line({ name: w.name }) : m.want_line({ name: w.name })}</li>
							{/each}
						</ul>
					{/if}
					{#if g.workTasks.length}
						<h4>{m.work_tasks_title()}</h4>
						<ul class="stack">
							{#each g.workTasks as t, i (i)}<li>{t.name}</li>{/each}
						</ul>
					{/if}
					{#if g.promotionSteps.length}
						<h4>{m.promotion_title()}</h4>
						<ul class="stack">
							{#each g.promotionSteps as step, i (i)}
								<li>{step.done ? m.step_done({ name: step.name }) : m.step_todo({ name: step.name })}</li>
							{/each}
						</ul>
					{/if}
				</div>
			</details>
		{/if}

		<details>
			<summary>{m.skills_title()}</summary>
			{#if sim.skills.length === 0}
				<p class="muted">{m.skills_none()}</p>
			{:else}
				<ul class="stack">
					{#each sim.skills as skill (skill.key)}
						<li>{m.skill_line({ name: skill.name, level: skill.level, max: skill.maxLevel })}</li>
					{/each}
				</ul>
			{/if}
		</details>

		<details>
			<summary>{m.relationships_title()}</summary>
			{#if sim.relationships.length === 0}
				<p class="muted">{m.relationships_none()}</p>
			{:else}
				<p class="muted">{m.relationships_help()}</p>
				<ul class="stack">
					{#each sim.relationships as rel (rel.simId)}
						<li>
							{rel.romance
								? m.rel_both({ name: rel.name, friendship: rel.friendship, romance: rel.romance })
								: m.rel_friend({ name: rel.name, friendship: rel.friendship })}
							{#if rel.labels.length}<span class="muted"> · {rel.labels.join(', ')}</span>{/if}
						</li>
					{/each}
				</ul>
			{/if}
		</details>
	</section>
{/each}

<style>
	.needs {
		display: grid;
		gap: 0.5rem;
	}
	.bar {
		display: block;
		height: 0.5rem;
		border-radius: 999px;
		background: var(--surface-2);
		overflow: hidden;
	}
	.bar > span {
		display: block;
		height: 100%;
		background: var(--accent);
	}
	.low .bar > span {
		background: var(--warn);
	}
	.critical .bar > span {
		background: var(--bad);
	}
	.critical {
		font-weight: 700;
	}
	summary {
		font-weight: 600;
		min-height: var(--tap);
		display: flex;
		align-items: center;
		cursor: pointer;
	}
	.action {
		display: grid;
		gap: 0.4rem;
	}
</style>
