<script lang="ts">
	import { Alert, Button, Icon, IconButton, Input, Modal, Spinner } from '@facile/muse';
	import { ARCHETYPES, RANDOM_NAMES, TRAITS } from '$lib/assistant';
	import { updateAssistant } from '$lib/backend';
	import { CHAT_ICONS } from '$lib/icons';
	import { session, setAssistant } from '$lib/session.svelte';

	let { open = $bindable(true) }: { open?: boolean } = $props();

	let name = $state(session.user?.assistant?.name || '');
	let archetype = $state(session.user?.assistant?.archetype || 'pragmatic_senior');
	let traits = $state<string[]>(
		session.user?.assistant?.traits ? [...session.user.assistant.traits] : ['concise']
	);
	let saving = $state(false);
	let error = $state<string | null>(null);

	function randomizeName() {
		const pool = RANDOM_NAMES.filter((n) => n.toLowerCase() !== name.trim().toLowerCase());
		name = pool[Math.floor(Math.random() * pool.length)] || RANDOM_NAMES[0];
	}

	function toggleTrait(id: string) {
		if (hasTrait(id)) {
			const item = TRAITS.find((t) => t.id === id);
			traits = traits.filter((t) => t !== id && t !== item?.label);
		} else {
			traits = [...traits, id];
		}
	}

	function hasTrait(id: string): boolean {
		const item = TRAITS.find((t) => t.id === id);
		return traits.includes(id) || (item ? traits.includes(item.label) : false);
	}

	async function submit() {
		saving = true;
		error = null;
		const trimmedName = name.trim() || 'Bulle';
		const avatarUrl = `https://blobatar.dev/avatar/${encodeURIComponent(trimmedName)}?gen=2`;
		try {
			const profile = await updateAssistant({
				name: trimmedName,
				avatar_url: avatarUrl,
				archetype,
				traits
			});
			setAssistant(profile);
			open = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not save assistant profile';
		} finally {
			saving = false;
		}
	}
</script>

<Modal bind:open title="Meet your assistant" size="lg" dismissible={false}>
	<div class="flex flex-col gap-5">
		<p class="text-fc-sm text-fc-fg-muted">
			Choose how your assistant works, communicates, and looks.
		</p>

		{#if error}
			<Alert tone="danger">{error}</Alert>
		{/if}

		<div class="flex items-center gap-4">
			<img
				src={`https://blobatar.dev/avatar/${encodeURIComponent(name.trim() || 'bulle')}?gen=2`}
				alt="Avatar"
				class="size-16 rounded-full border-2 border-fc-border bg-fc-surface"
			/>
			<div class="flex flex-1 flex-col gap-1 min-w-0">
				<label for="assistant-name" class="text-fc-sm font-medium text-fc-fg">
					Assistant Name
				</label>
				<div class="flex items-center gap-2">
					<Input
						id="assistant-name"
						bind:value={name}
						placeholder="e.g. Ada, Atlas, Echo, Jarvis"
					/>
					<IconButton
						aria-label="Randomize name"
						onclick={randomizeName}
					>
						<Icon icon={CHAT_ICONS.dice} size={18} />
					</IconButton>
				</div>
			</div>
		</div>

		<div class="flex flex-col gap-2">
			<span class="text-fc-sm font-medium text-fc-fg">Archetype</span>
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2" role="radiogroup" aria-label="Archetype">
				{#each ARCHETYPES as item (item.id)}
					{@const selected = archetype === item.id || archetype === item.label}
					<button
						type="button"
						role="radio"
						aria-checked={selected}
						onclick={() => (archetype = item.id)}
						class="flex flex-col rounded-fc-lg border p-3 text-left transition-all cursor-pointer focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring {selected ? 'border-fc-accent bg-fc-accent/10 ring-2 ring-fc-accent text-fc-fg' : 'border-fc-border bg-fc-component text-fc-fg hover:border-fc-border-hover hover:bg-fc-surface'}"
					>
						<span class="font-medium text-fc-sm">{item.label}</span>
						<span class="mt-1 text-fc-xs text-fc-fg-muted">{item.description}</span>
					</button>
				{/each}
			</div>
		</div>

		<div class="flex flex-col gap-2">
			<span class="text-fc-sm font-medium text-fc-fg">Traits</span>
			<div class="flex flex-wrap gap-2">
				{#each TRAITS as item (item.id)}
					{@const active = hasTrait(item.id)}
					<button
						type="button"
						aria-pressed={active}
						onclick={() => toggleTrait(item.id)}
						class="inline-flex items-center rounded-fc-pill px-3 py-1.5 text-fc-xs font-medium border transition-colors cursor-pointer focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring {active ? 'border-fc-accent bg-fc-accent text-fc-accent-fg' : 'border-fc-border bg-fc-surface text-fc-fg hover:bg-fc-component hover:border-fc-border-hover'}"
					>
						{item.label}
					</button>
				{/each}
			</div>
		</div>

		<Button
			variant="primary"
			disabled={saving}
			onclick={submit}
			class="mt-2 w-full"
		>
			{#if saving}
				<Spinner size="sm" />
			{/if}
			Save and Start Coding
		</Button>
	</div>
</Modal>
