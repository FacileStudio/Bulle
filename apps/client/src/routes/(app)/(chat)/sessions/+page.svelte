<script lang="ts">
	import { goto } from '$app/navigation';
	import { Alert, EmptyState } from '@facile/muse';
	import { api } from '$lib/api';
	import { sessions } from '$lib/chat/sessions.svelte';
	import { session } from '$lib/session.svelte';
	import Composer from '$lib/components/Composer.svelte';
	import OnboardingModal from '$lib/components/OnboardingModal.svelte';
	import { CHAT_ICONS } from '$lib/icons';

	let error = $state<string | null>(null);

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
	<EmptyState
		bare
		icon={CHAT_ICONS.chat}
		title="What should the agent do?"
		description="Ask a question or hand over a task. It can use its own computer to get it done."
	/>
	{#if error}
		<Alert tone="danger">{error}</Alert>
	{/if}
	<Composer onsend={start} placeholder="Describe a task" />
</div>

{#if session.status === 'signed-in' && (!session.user?.assistant || !session.user.assistant.name)}
	<OnboardingModal />
{/if}
