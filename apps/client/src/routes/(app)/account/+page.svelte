<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Alert,
		Button,
		PageHeader,
		ProfileCard,
		Section,
		Spinner,
		Stack,
		icons,
		toast
	} from '@facile/muse';
	import { avatarUrlFor } from '$lib/assistant';
	import ComputerSettings from '$lib/components/settings/ComputerSettings.svelte';
	import ProviderKeys from '$lib/components/settings/ProviderKeys.svelte';
	import SuiteConnections from '$lib/components/SuiteConnections.svelte';
	import { backend, updateAssistant, type PushConfig } from '$lib/backend';
	import { isNative } from '$lib/platform';
	import ArchetypeCards from '$lib/components/onboarding/ArchetypeCards.svelte';
	import NameField from '$lib/components/onboarding/NameField.svelte';
	import TraitChips from '$lib/components/onboarding/TraitChips.svelte';
	import { initialDraft, profileFromDraft } from '$lib/onboarding/profile';
	import { session, setAssistant, signOut } from '$lib/session.svelte';

	const start = initialDraft(session.user?.assistant);
	let name = $state(start.name);
	let archetype = $state(start.archetype);
	let traits = $state<string[]>(start.traits);
	let saving = $state(false);
	let saveSuccess = $state(false);
	let saveError = $state<string | null>(null);

	$effect(() => {
		if (session.user?.assistant && !name && session.user.assistant.name) {
			const loaded = initialDraft(session.user.assistant);
			name = loaded.name;
			archetype = loaded.archetype;
			traits = loaded.traits;
		}
	});

	async function saveAssistant() {
		saving = true;
		saveError = null;
		saveSuccess = false;
		try {
			const updated = await updateAssistant(profileFromDraft({ step: 0, name, archetype, traits }));
			setAssistant(updated);
			saveSuccess = true;
		} catch (err) {
			saveError = err instanceof Error ? err.message : 'Could not save assistant profile';
		} finally {
			saving = false;
		}
	}

	let push = $state<PushConfig | null>(null);
	let testing = $state(false);
	const native = isNative();
	const ntfyUrl = $derived(
		push?.enabled ? `ntfy://${new URL(push.server).host}/${push.topic}` : ''
	);

	onMount(async () => {
		try {
			push = await backend.getPush();
		} catch {
			push = null;
		}
	});

	function openNtfy() {
		window.location.href = ntfyUrl;
	}

	async function copySubscribeUrl() {
		if (!push) return;
		try {
			await navigator.clipboard.writeText(push.subscribe_url);
			toast.success('Link copied');
		} catch {
			toast.danger('Could not copy the link');
		}
	}

	async function sendTest() {
		testing = true;
		try {
			await backend.testPush();
			toast.success('Test notification sent');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not send the test');
		} finally {
			testing = false;
		}
	}

	async function leave() {
		await signOut();
		await goto('/login', { replaceState: true });
	}
</script>

<svelte:head>
	<title>Settings - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader title="Settings" description="Your account, assistant, providers and computer." />
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
						src={avatarUrlFor(name)}
						alt=""
						aria-hidden="true"
						class="size-16 rounded-full object-cover"
					/>
					<div class="min-w-0 flex-1">
						<NameField id="account-assistant-name" bind:value={name} />
					</div>
				</div>

				<div class="flex flex-col gap-2">
					<span class="text-fc-sm font-medium text-fc-fg">Personality</span>
					<ArchetypeCards bind:archetype {traits} />
				</div>

				<div class="flex flex-col gap-2">
					<span class="text-fc-sm font-medium text-fc-fg">Traits</span>
					<TraitChips bind:traits />
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

		{#if push?.enabled}
			<Section
				card
				title="Notifications"
				description="Get a push notification when a job finishes."
			>
				<div class="flex flex-col gap-4">
					<p class="text-fc-sm text-fc-fg-muted">
						Install the ntfy app on your phone. Subscribe to the topic below on this server.
					</p>
					<dl class="grid gap-1 text-fc-sm">
						<dt class="text-fc-fg-muted">Server</dt>
						<dd class="font-mono break-all text-fc-fg">{push.server}</dd>
						<dt class="mt-2 text-fc-fg-muted">Topic</dt>
						<dd class="font-mono break-all text-fc-fg">{push.topic}</dd>
					</dl>
					<div class="flex flex-wrap gap-2">
						{#if native}
							<Button variant="primary" onclick={openNtfy}>Open in ntfy</Button>
						{:else}
							<Button variant="outline" onclick={copySubscribeUrl}>Copy subscribe link</Button>
						{/if}
						<Button variant="outline" disabled={testing} onclick={sendTest}>
							{#if testing}
								<Spinner size="sm" />
							{/if}
							Send test
						</Button>
					</div>
					{#if !native}
						<p class="font-mono text-fc-xs break-all text-fc-fg-subtle">{push.subscribe_url}</p>
					{/if}
				</div>
			</Section>
		{/if}

		<ProviderKeys />

		<ComputerSettings />

		<SuiteConnections />
	{/if}
</Stack>
