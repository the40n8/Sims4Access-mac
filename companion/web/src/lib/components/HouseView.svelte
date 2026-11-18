<script lang="ts">
	import { tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client } from '$lib/client.js';
	import { live } from '$lib/live.svelte.js';
	import { reasonText, simName } from '$lib/text.js';
	import type { HouseMap, Room, RoomItem, RoomLink, Target } from '$lib/gen/ts4access/v1/game_pb.js';
	import { isMinor, objectName, rank } from '$lib/targets.js';
	import { dir, floorName, kindName, part, roomName, side } from '$lib/rooms.js';
	import { announce } from './Announcer.svelte';
	import InteractionMenu from './InteractionMenu.svelte';

	// A mental map of the house: floors, rooms, what joins them, what is inside.
	const game = $derived(live.state);
	const me = $derived(game?.sims.find((s) => s.active));

	let house = $state<HouseMap | null>(null);
	let failed = $state(false);
	let room = $state<Room | null>(null);
	let target = $state<Target | null>(null);
	let roomHeading: HTMLHeadingElement | undefined = $state();

	async function load() {
		house = null;
		failed = false;
		try {
			const res = await client.getHouseMap({});
			if (!res.outcome?.ok) {
				failed = true;
				announce(m.failed({ reason: reasonText(res.outcome?.reason ?? '') }));
				return;
			}
			house = res.map ?? null;
			// Keep the open room up to date (Sims move, objects change).
			if (room) room = house?.rooms.find((r) => r.id === room!.id) ?? null;
		} catch {
			failed = true;
		}
	}

	let loaded = false;
	$effect(() => {
		if (game?.inGame && !loaded) {
			loaded = true;
			void load();
		}
	});

	const byId = (id: string) => house?.rooms.find((r) => r.id === id);
	function linkText(l: RoomLink): string {
		const args = { room: linkTo(l.roomId), side: side(l.side) };
		if (l.stairs === 'up') return m.link_stairs_up(args);
		if (l.stairs === 'down') return m.link_stairs_down(args);
		return l.door ? m.link_door(args) : m.link_open(args);
	}
	const linkTo = (id: string) => {
		const r = id ? byId(id) : undefined;
		return r ? roomName(r) : m.outside();
	};

	// The garden and the pool are listed like rooms but not counted as rooms.
	const OUTDOOR = ['OUTSIDE', 'POOL'];

	function roomLine(r: Room): string {
		const bits =
			r.kind === 'OUTSIDE'
				? [roomName(r)]
				: [roomName(r), part(r.part), m.room_size({ width: r.width, depth: r.depth })];
		if (r.here) bits.push(m.room_sims({ names: r.sims.join(', ') }));
		else {
			if (r.bearing) bits.push(m.room_away({ meters: r.distance, dir: dir(r.bearing) }));
			if (r.sims.length) bits.push(m.room_sims({ names: r.sims.join(', ') }));
		}
		return bits.join(', ');
	}

	const floors = $derived(
		(house?.floors ?? []).map((f) => ({ floor: f, rooms: (house?.rooms ?? []).filter((r) => r.floor === f) }))
	);
	// Stairs seen from a floor: going up from it, or coming down to it.
	function stairsOn(floor: number): string[] {
		const out: string[] = [];
		for (const st of house?.stairs ?? []) {
			const bottom = byId(st.bottomRoomId);
			const top = byId(st.topRoomId);
			if (st.floor === floor)
				out.push(
					m.stairs_up_line({
						where: part(st.part),
						room: bottom ? roomName(bottom) : m.outside(),
						to: floorName(floor + 1)
					})
				);
			else if (st.floor + 1 === floor)
				out.push(m.stairs_down_line({ room: top ? roomName(top) : m.outside(), to: floorName(st.floor) }));
		}
		return out;
	}

	const current = $derived(house?.rooms.find((r) => r.here));
	const summary = $derived.by(() => {
		if (!house) return '';
		const rooms = house.rooms.filter((r) => !OUTDOOR.includes(r.kind)).length;
		const count =
			house.floors.length > 1
				? m.house_floors({ rooms, floors: house.floors.length })
				: m.house_one_floor({ rooms });
		if (!me) return count;
		const where = current
			? m.house_sim_in({ name: simName(me), room: roomName(current) })
			: m.house_sim_outside({ name: simName(me) });
		return `${count} ${where}`;
	});

	function sorted(items: RoomItem[]): RoomItem[] {
		return [...items].sort((a, b) => rank(a.target!) - rank(b.target!));
	}
	const main = $derived(sorted((room?.items ?? []).filter((i) => i.target && !isMinor(i.target))));
	const minor = $derived(sorted((room?.items ?? []).filter((i) => i.target && isMinor(i.target))));

	let lastId = '';
	async function openRoom(r: Room) {
		room = r;
		lastId = `room-${r.id}`;
		await tick();
		roomHeading?.focus();
	}
	async function backToHouse() {
		room = null;
		await tick();
		document.getElementById(lastId)?.focus();
	}
	function openItem(t: Target) {
		target = t;
	}
	async function backToRoom() {
		const id = `item-${target?.id}`;
		target = null;
		await tick();
		document.getElementById(id)?.focus();
	}
</script>

{#if game?.inGame}
	{#if target}
		<InteractionMenu {target} onBack={backToRoom} />
	{:else if room}
		<section class="card stack" aria-labelledby="room-title">
			<button type="button" onclick={backToHouse}>{m.back_to_house()}</button>
			<h2 id="room-title" tabindex="-1" bind:this={roomHeading}>{roomName(room)}</h2>
			<p>
				{#if room.kind === 'OUTSIDE'}
					{floorName(room.floor)}.
				{:else}
					{floorName(room.floor)}, {part(room.part)}, {m.room_size({ width: room.width, depth: room.depth })}.
				{/if}
				{#if !room.here && room.bearing}{m.room_away({ meters: room.distance, dir: dir(room.bearing) })}.{/if}
			</p>
			{#if room.areas.length > 1}
				<p>
					{room.areas.map((a) => m.room_area({ kind: kindName(a.kind), side: side(a.side) })).join('. ')}.
				</p>
			{/if}
			{#if room.sims.length}<p>{m.room_sims({ names: room.sims.join(', ') })}</p>{/if}

			<h3>{m.room_links()}</h3>
			{#if room.links.length === 0}
				<p class="muted">{m.link_none()}</p>
			{:else}
				<ul class="links">
					{#each room.links as l (l.roomId + l.side)}
						<li>
							{#if l.roomId && byId(l.roomId)}
								<button type="button" class="link" onclick={() => openRoom(byId(l.roomId)!)}>
									{linkText(l)}
								</button>
							{:else}
								{linkText(l)}
							{/if}
						</li>
					{/each}
				</ul>
			{/if}

			<h3>{m.in_room()}</h3>
			{#if main.length === 0 && minor.length === 0}
				<p class="muted">{m.room_nothing()}</p>
			{/if}
			<ul class="stack">
				{#each main as i (i.target!.id)}
					<li>
						<button type="button" class="row" id={`item-${i.target!.id}`} onclick={() => openItem(i.target!)}>
							{objectName(i.target!)}, {side(i.side)}
						</button>
					</li>
				{/each}
			</ul>
			{#if minor.length}
				<details>
					<summary>{m.minor_items({ count: minor.length })}</summary>
					<ul class="stack">
						{#each minor as i (i.target!.id)}
							<li>
								<button
									type="button"
									class="row"
									id={`item-${i.target!.id}`}
									onclick={() => openItem(i.target!)}
								>
									{objectName(i.target!)}, {side(i.side)}
								</button>
							</li>
						{/each}
					</ul>
				</details>
			{/if}
		</section>
	{:else}
		<section class="card stack">
			{#if house === null && !failed}
				<p class="muted" role="status">{m.loading()}</p>
			{:else if failed || !house || house.rooms.length === 0}
				<p class="muted">{m.house_empty()}</p>
			{:else}
				<p>{summary}</p>
				{#if me}<p class="muted">{m.house_help({ name: simName(me) })}</p>{/if}
			{/if}
			<button type="button" onclick={load}>{m.refresh()}</button>
		</section>
		{#each floors as f (f.floor)}
			<section class="card stack" aria-labelledby={`floor-${f.floor}`}>
				<h2 id={`floor-${f.floor}`}>{floorName(f.floor)}</h2>
				<ul class="stack">
					{#each f.rooms as r (r.id)}
						<li>
							<button type="button" class="row" id={`room-${r.id}`} onclick={() => openRoom(r)}
								>{roomLine(r)}</button
							>
						</li>
					{/each}
				</ul>
				{#each stairsOn(f.floor) as st, i (i)}
					<p>{st}</p>
				{/each}
			</section>
		{/each}
	{/if}
{/if}

<style>
	.row,
	.link {
		width: 100%;
		text-align: start;
	}
	.links {
		list-style: none;
		padding: 0;
		margin: 0;
		display: grid;
		gap: 0.5rem;
	}
	h2:focus {
		outline: none;
	}
	summary {
		min-height: var(--tap);
		display: flex;
		align-items: center;
		cursor: pointer;
	}
</style>
