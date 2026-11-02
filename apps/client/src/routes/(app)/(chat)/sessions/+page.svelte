<script lang="ts">
	import { goto } from '$app/navigation';
	import { Alert } from '@facile/muse';
	import { Blobatar } from '@blobatar/svelte';
	import { api } from '$lib/api';
	import { sessions } from '$lib/chat/sessions.svelte';
	import { session } from '$lib/session.svelte';
	import Composer from '$lib/components/Composer.svelte';
	import OnboardingModal from '$lib/components/OnboardingModal.svelte';

	let error = $state<string | null>(null);
	const assistantName = $derived(session.user?.assistant?.name?.trim() || 'bulle');

	async function start(text: string): Promise<boolean> {
		error = null;
		try {
			const id = await sessions.create();
			await api().sendMessage(id, text);
			await goto(`/s/${id}`);
			return true;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not start the chat';
			return false;
		}
	}
</script>

<svelte:head>
	<title>New chat - Bulle</title>
</svelte:head>

<div class="mx-auto flex h-full w-full max-w-fc-md flex-col justify-center gap-6 px-4 pb-4">
	<div class="flex flex-col items-center gap-4 py-8 text-center select-none">
		<div
			class="relative flex items-center justify-center cursor-pointer transition-transform duration-200 hover:scale-105"
		>
			<Blobatar
				name={assistantName}
				size={128}
				animate="hover"
				class="drop-shadow-sm transition-transform duration-200"
			/>
		</div>
		<div class="flex flex-col gap-1 max-w-fc-sm">
			<h2 class="text-fc-md font-medium text-fc-fg">What should the agent do?</h2>
			<p class="text-fc-sm text-fc-fg-muted">
				Ask a question or hand over a task. It can use its own computer to get it done.
			</p>
		</div>
	</div>
	{#if error}
		<Alert tone="danger">{error}</Alert>
	{/if}
	<Composer onsend={start} placeholder="Describe a task" />
</div>

{#if session.status === 'signed-in' && (!session.user?.assistant || !session.user.assistant.name)}
	<OnboardingModal />
{/if}
