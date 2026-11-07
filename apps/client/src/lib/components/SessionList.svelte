<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		Button,
		ConfirmModal,
		EmptyState,
		IconButton,
		Icon,
		Skeleton,
		icons,
		toast
	} from '@facile/muse';
	import type { SessionSummary } from '$lib/api';
	import { relativeTime } from '$lib/chat/format';
	import { shouldSearch } from '$lib/chat/search';
	import { sessions } from '$lib/chat/sessions.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import SessionSearch from './SessionSearch.svelte';
	import SessionStatus from './SessionStatus.svelte';

	let { onnavigate }: { onnavigate?: () => void } = $props();

	let pending = $state<SessionSummary | null>(null);
	let confirmOpen = $state(false);
	let query = $state('');

	const activeId = $derived(page.params.id);

	function ask(session: SessionSummary) {
		pending = session;
		confirmOpen = true;
	}

	async function confirmDelete() {
		const target = pending;
		if (!target) return;
		try {
			await sessions.remove(target.id);
			if (target.id === activeId) await goto('/sessions');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('session.deleteFailed'));
		}
		confirmOpen = false;
	}

	async function fresh() {
		await goto('/sessions');
		onnavigate?.();
	}
</script>

<div class="flex h-full min-h-0 flex-col gap-3 p-3">
	<Button icon={icons.plus} variant="ghost" class="h-11 bg-fc-surface text-fc-fg" onclick={fresh}>{t('chat.new')}</Button>

	<nav class="-mx-1 min-h-0 flex-1 overflow-y-auto px-1" aria-label={t('session.list')}>
		<div class="px-1 pb-2">
			<SessionSearch bind:query {onnavigate} />
		</div>
		{#if shouldSearch(query)}{:else if !sessions.loaded}
			<div class="flex flex-col gap-2">
				<Skeleton class="h-11 w-full" />
				<Skeleton class="h-11 w-full" />
				<Skeleton class="h-11 w-full" />
			</div>
		{:else if sessions.error && !sessions.items.length}
			<EmptyState
				bare
				icon={icons.error}
				title={t('session.loadFailed')}
				description={sessions.error}
			>
				<Button variant="ghost" class="h-11" icon={icons.refresh} onclick={() => sessions.refresh()}>
					{t('common.retry')}
				</Button>
			</EmptyState>
		{:else if !sessions.items.length}
			<EmptyState
				bare
				icon={CHAT_ICONS.chat}
				title={t('session.emptyTitle')}
				description={t('session.emptyBody')}
			/>
		{:else}
			<ul class="flex flex-col gap-1">
				{#each sessions.items as session (session.id)}
					{@const active = session.id === activeId}
					{@const busy = session.status === 'running' || session.status === 'waiting'}
					<li class="group relative">
						<a
							href={`/s/${session.id}`}
							aria-current={active ? 'page' : undefined}
							onclick={() => onnavigate?.()}
							class="flex min-h-11 flex-col justify-center rounded-fc-md py-1 pl-3 hover:bg-fc-surface {busy
								? 'pr-20 md:pr-12'
								: 'pr-12'} focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring {active
								? 'bg-fc-surface'
								: ''}"
						>
							<span class="truncate text-fc-sm {active ? 'font-medium' : ''}">
								{session.title || t('chat.new')}
							</span>
							<span class="text-fc-xs text-fc-fg-muted">{relativeTime(session.updated_at)}</span>
						</a>
						<SessionStatus
							status={session.status}
							class="pointer-events-none absolute top-1/2 right-12 -translate-y-1/2 md:right-4 md:group-focus-within:opacity-0 md:group-hover:opacity-0"
						/>
						<IconButton
							variant="danger"
							aria-label={t('session.delete', { title: session.title || t('session.one') })}
							class="absolute top-0 right-0 md:opacity-0 md:group-focus-within:opacity-100 md:group-hover:opacity-100"
							onclick={() => ask(session)}
						>
							<Icon icon={icons.remove} size={16} />
						</IconButton>
					</li>
				{/each}
			</ul>
		{/if}
	</nav>

</div>

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title={t('session.deleteTitle')}
	description={t('session.deleteBody')}
	confirmLabel={t('common.delete')}
	onConfirm={confirmDelete}
	onCancel={() => (confirmOpen = false)}
/>
