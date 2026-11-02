<script lang="ts">
	import { goto } from '$app/navigation';
	import {
		Alert,
		Button,
		Icon,
		IconButton,
		Input,
		PageHeader,
		ProfileCard,
		Section,
		Spinner,
		Stack,
		icons
	} from '@facile/muse';
	import { ARCHETYPES, RANDOM_NAMES, TRAITS } from '$lib/assistant';
	import SuiteConnections from '$lib/components/SuiteConnections.svelte';
	import { updateAssistant } from '$lib/backend';
	import { CHAT_ICONS } from '$lib/icons';
	import { session, setAssistant, signOut } from '$lib/session.svelte';

	let name = $state(session.user?.assistant?.name || '');
	let archetype = $state(session.user?.assistant?.archetype || 'pragmatic_senior');
	let traits = $state<string[]>(
		session.user?.assistant?.traits ? [...session.user.assistant.traits] : ['concise']
	);
	let saving = $state(false);
	let saveSuccess = $state(false);
	let saveError = $state<string | null>(null);

	$effect(() => {
		if (session.user?.assistant && !name && session.user.assistant.name) {
			name = session.user.assistant.name;
			archetype = session.user.assistant.archetype || 'pragmatic_senior';
			traits = session.user.assistant.traits ? [...session.user.assistant.traits] : ['concise'];
		}
	});

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

	async function saveAssistant() {
		saving = true;
		saveError = null;
		saveSuccess = false;
		const trimmedName = name.trim() || 'Bulle';
		const avatarUrl = `https://blobatar.dev/avatar/${encodeURIComponent(trimmedName)}?gen=2`;
		try {
			const updated = await updateAssistant({
				name: trimmedName,
				avatar_url: avatarUrl,
				archetype,
				traits
			});
			setAssistant(updated);
			saveSuccess = true;
		} catch (err) {
			saveError = err instanceof Error ? err.message : 'Could not save assistant profile';
		} finally {
			saving = false;
		}
	}

	async function leave() {
		await signOut();
		await goto('/login', { replaceState: true });
	}
</script>

<svelte:head>
	<title>Account - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader title="Account" description="Your identity comes from Facile." />
	{#if session.user}
		<ProfileCard name={session.user.name.trim() || session.user.email} email={session.user.email}>
			{#snippet actions()}
				<Button variant="outline" icon={icons.logout} onclick={leave}>Sign out</Button>
			{/snippet}
		</ProfileCard>

		<Section card title="Assistant Profile" description="Customize your assistant persona, avatar, and traits.">
			<div class="flex flex-col gap-5">
				{#if saveSuccess}
					<Alert tone="success">Assistant profile updated.</Alert>
				{/if}
				{#if saveError}
					<Alert tone="danger">{saveError}</Alert>
				{/if}

				<div class="flex items-center gap-4">
					<img
						src={`https://blobatar.dev/avatar/${encodeURIComponent(name.trim() || 'bulle')}?gen=2`}
						alt=""
						aria-hidden="true"
						class="size-16 rounded-full object-cover"
					/>
					<div class="flex flex-1 flex-col gap-1 min-w-0">
						<label for="account-assistant-name" class="text-fc-sm font-medium text-fc-fg">
							Assistant Name
						</label>
						<div class="flex items-center gap-2">
							<Input
								id="account-assistant-name"
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

				<div>
					<Button
						variant="primary"
						disabled={saving}
						onclick={saveAssistant}
					>
						{#if saving}
							<Spinner size="sm" />
						{/if}
						Save Changes
					</Button>
				</div>
			</div>
		</Section>

		<SuiteConnections />
	{/if}
</Stack>
