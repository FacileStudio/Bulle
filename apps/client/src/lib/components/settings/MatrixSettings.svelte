<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Badge, Button, Field, Input, Section, Spinner, Switch, Textarea, toast } from '@facile/muse';
	import { getMatrix, parseAllow, saveMatrix, statusTone, type MatrixSettings } from '$lib/matrix';
	import { spaces } from '$lib/spaces.svelte';

	let settings = $state<MatrixSettings | null>(null);
	let password = $state('');
	let allowText = $state('');
	let rooms = $state<Record<string, string>>({});
	let loadError = $state<string | null>(null);
	let saving = $state(false);

	function load(next: MatrixSettings) {
		settings = next;
		allowText = next.allow.join('\n');
		rooms = { ...next.rooms };
		password = '';
	}

	onMount(async () => {
		try {
			load(await getMatrix());
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Could not load the Matrix settings';
		}
	});

	async function save() {
		if (!settings) return;
		saving = true;
		try {
			load(
				await saveMatrix({
					...settings,
					password: password || undefined,
					allow: parseAllow(allowText),
					rooms
				})
			);
			toast.success('Matrix settings saved');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not save the Matrix settings');
		} finally {
			saving = false;
		}
	}
</script>

<Section
	card
	id="matrix"
	title="Matrix"
	description="Your assistant joins each space's Matrix room as a teammate. It answers only when someone mentions it, in a thread, and works in that space."
>
	{#if loadError}
		<Alert tone="danger">{loadError}</Alert>
	{:else if !settings}
		<div class="flex justify-center py-6"><Spinner /></div>
	{:else}
		<form
			class="flex flex-col gap-5"
			onsubmit={(e) => {
				e.preventDefault();
				void save();
			}}
		>
			<div class="flex flex-wrap items-center gap-3">
				<Switch label="Bot enabled" bind:checked={settings.enabled} />
				<Badge tone={statusTone(settings.status)}>{settings.status}</Badge>
			</div>

			<Field label="Homeserver" helper="For example https://matrix.org">
				{#snippet children({ id, describedBy })}
					<Input {id} aria-describedby={describedBy} placeholder="https://matrix.org" bind:value={settings!.homeserver} />
				{/snippet}
			</Field>

			<Field label="Bot user id" helper="The Matrix account the assistant speaks as, such as @bulle:matrix.org">
				{#snippet children({ id, describedBy })}
					<Input {id} aria-describedby={describedBy} placeholder="@bulle:matrix.org" bind:value={settings!.user_id} />
				{/snippet}
			</Field>

			<Field
				label="Bot password"
				helper={settings.password_set ? 'Stored encrypted. Leave empty to keep it.' : 'Stored encrypted on the server, never shown again.'}
			>
				{#snippet children({ id, describedBy })}
					<Input
						{id}
						aria-describedby={describedBy}
						type="password"
						autocomplete="off"
						placeholder={settings!.password_set ? '••••••••' : ''}
						bind:value={password}
					/>
				{/snippet}
			</Field>

			<Field
				label="Who may call it"
				helper="Matrix user ids, one per line. The assistant runs with your access, so list only people you trust."
			>
				{#snippet children({ id, describedBy })}
					<Textarea {id} aria-describedby={describedBy} rows={3} placeholder="@you:matrix.org" bind:value={allowText} />
				{/snippet}
			</Field>

			<div class="flex flex-col gap-3">
				<span class="text-fc-sm font-medium text-fc-fg">Space rooms</span>
				<span class="text-fc-sm text-fc-fg-muted">
					The room id of each space's team room, under room settings, advanced. Invite the bot to the room. Leave empty
					for no room.
				</span>
				{#each spaces.items as space (space.id)}
					<Field label={space.name}>
						{#snippet children({ id, describedBy })}
							<Input
								{id}
								aria-describedby={describedBy}
								placeholder="!roomid:matrix.org"
								value={rooms[space.id] ?? ''}
								oninput={(e) => (rooms[space.id] = e.currentTarget.value)}
							/>
						{/snippet}
					</Field>
				{/each}
			</div>

			<div>
				<Button type="submit" variant="primary" disabled={saving}>
					{#if saving}<Spinner size="sm" />{/if}
					Save
				</Button>
			</div>
		</form>
	{/if}
</Section>
