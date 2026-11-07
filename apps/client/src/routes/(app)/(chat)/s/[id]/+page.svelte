<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { tick, untrack } from 'svelte';
	import { page } from '$app/state';
	import { Alert, Button, EmptyState, PlanLimit, Skeleton, icons } from '@facile/muse';
	import { chat } from '$lib/chat/chat.svelte';
	import Composer from '$lib/components/Composer.svelte';
	import FileDropZone from '$lib/components/FileDropZone.svelte';
	import { actionMessage } from '$lib/openui/actions';
	import Message from '$lib/components/Message.svelte';
	import { CHAT_ICONS } from '$lib/icons';

	let scroller = $state<HTMLElement | null>(null);
	let pinned = true;

	const id = $derived(page.params.id ?? '');
	const view = $derived(chat.state);
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

	let composer = $state<Composer | null>(null);
</script>

<svelte:head>
	<title>{chat.title || t('chat.defaultTitle')} - Bulle</title>
</svelte:head>

<FileDropZone class="h-full" disabled={chat.loading || !!chat.loadError} ondrop={(files) => composer?.addFiles(files)}>
	<div class="flex h-full min-h-0 flex-col">
		<div bind:this={scroller} {onscroll} class="min-h-0 flex-1 overflow-y-auto overscroll-contain">
			<div
				class="mx-auto flex w-full max-w-fc-md flex-col gap-6 px-4 py-6"
				role="log"
				aria-live="polite"
				aria-label={t('chat.conversation')}
			>
				{#if chat.loading}
					<Skeleton class="h-10 w-2/3 self-end" />
					<Skeleton class="h-24 w-full" />
					<Skeleton class="h-16 w-4/5" />
				{:else if chat.loadError}
					<EmptyState
						bare
						icon={icons.error}
						title={t('chat.loadFailed')}
						description={chat.loadError}
					>
						<Button icon={icons.refresh} onclick={() => chat.open(id)}>{t('common.tryAgain')}</Button>
					</EmptyState>
				{:else if !view.items.length}
					<EmptyState
						bare
						icon={CHAT_ICONS.chat}
						title={t('chat.nothingTitle')}
						description={t('chat.nothingBody')}
					/>
				{:else}
					{#each view.items as item, index (item.id)}
						{#if !(focus && index === 0 && item.role === 'user')}
						<Message
							{item}
							{focus}
							ondecide={(callId, decision) => chat.decide(callId, decision)}
							ondecideall={(callIds, decision) => chat.decideAll(callIds, decision)}
							onaction={(action) => chat.send(actionMessage(action))}
							answered={index < view.items.length - 1}
							busy={view.running}
							onanswer={(text) => chat.send(text)}
						/>
						{/if}
					{/each}
				{/if}
			</div>
		</div>

		<div class="shrink-0 px-4 pb-3">
			<div class="mx-auto flex w-full max-w-fc-md flex-col gap-2">
				<PlanLimit error={chat.planLimit} />
				{#if view.error && !view.items.some((i) => i.error)}
					<Alert tone="danger" title={t('chat.errorTitle')}>
						<div class="flex items-center justify-between gap-3">
							<span>{view.error}</span>
							<Button size="sm" variant="ghost" onclick={() => chat.dismissError()}>{t('common.dismiss')}</Button>
						</div>
					</Alert>
				{/if}
				{#if chat.connection === 'reconnecting' && view.running}
					<p class="text-fc-xs text-fc-fg-muted" role="status">{t('chat.reconnecting')}</p>
				{/if}
				<Composer
				bind:this={composer}
					running={view.running}
					disabled={chat.loading || !!chat.loadError}
					onsend={(text, attachments) => chat.send(text, attachments)}
					onstop={() => chat.stop()}
				/>
			</div>
		</div>
	</div>
</FileDropZone>
