<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { page } from '$app/state';
	import { Alert, Button, EmptyState, Skeleton, icons } from '@facile/muse';
	import { chat } from '$lib/chat/chat.svelte';
	import { formatCost, formatTokens } from '$lib/chat/format';
	import Composer from '$lib/components/Composer.svelte';
	import { actionMessage } from '$lib/openui/actions';
	import Message from '$lib/components/Message.svelte';
	import { CHAT_ICONS } from '$lib/icons';

	let scroller = $state<HTMLElement | null>(null);
	let pinned = true;

	const id = $derived(page.params.id ?? '');
	const view = $derived(chat.state);
	const hasFooter = $derived(view.tokens > 0 || view.costUsd > 0);
	const focus = $derived(chat.channel === 'cron');

	$effect(() => {
		const target = id;
		if (!target) return;
		untrack(() => void chat.open(target));
		return () => chat.close();
	});

	$effect(() => {
		void view;
		if (!pinned) return;
		void tick().then(() => scroller?.scrollTo({ top: scroller.scrollHeight }));
	});

	function onscroll() {
		if (!scroller) return;
		pinned = scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < 80;
	}
</script>

<svelte:head>
	<title>{chat.title || 'Chat'} - Bulle</title>
</svelte:head>

<div class="flex h-full min-h-0 flex-col">
	<div bind:this={scroller} {onscroll} class="min-h-0 flex-1 overflow-y-auto overscroll-contain">
		<div
			class="mx-auto flex w-full max-w-fc-md flex-col gap-6 px-4 py-6"
			role="log"
			aria-live="polite"
			aria-label="Conversation"
		>
			{#if chat.loading}
				<Skeleton class="h-10 w-2/3 self-end" />
				<Skeleton class="h-24 w-full" />
				<Skeleton class="h-16 w-4/5" />
			{:else if chat.loadError}
				<EmptyState
					bare
					icon={icons.error}
					title="Could not load this session"
					description={chat.loadError}
				>
					<Button icon={icons.refresh} onclick={() => chat.open(id)}>Try again</Button>
				</EmptyState>
			{:else if !view.items.length}
				<EmptyState
					bare
					icon={CHAT_ICONS.chat}
					title="Nothing here yet"
					description="Send a message to start the conversation."
				/>
			{:else}
				{#each view.items as item, index (item.id)}
					{#if !(focus && index === 0 && item.role === 'user')}
					<Message
						{item}
						{focus}
						ondecide={(callId, approve) => chat.decide(callId, approve)}
						onaction={(action) => chat.send(actionMessage(action))}
					/>
					{/if}
				{/each}
			{/if}
		</div>
	</div>

	<div class="shrink-0 px-4 pb-3">
		<div class="mx-auto flex w-full max-w-fc-md flex-col gap-2">
			{#if view.error && !view.items.some((i) => i.error)}
				<Alert tone="danger" title="Something went wrong">
					<div class="flex items-center justify-between gap-3">
						<span>{view.error}</span>
						<Button size="sm" variant="ghost" onclick={() => chat.dismissError()}>Dismiss</Button>
					</div>
				</Alert>
			{/if}
			{#if chat.connection === 'reconnecting' && view.running}
				<p class="text-fc-xs text-fc-fg-muted" role="status">Reconnecting to the live stream</p>
			{/if}
			<Composer
				running={view.running}
				disabled={chat.loading || !!chat.loadError}
				onsend={(text) => chat.send(text)}
				onstop={() => chat.stop()}
			/>
			<p class="min-h-4 text-center text-fc-xs text-fc-fg-muted">
				{#if hasFooter}
					{formatTokens(view.tokens)} tokens &middot; {formatCost(view.costUsd)} this visit
				{/if}
			</p>
		</div>
	</div>
</div>
