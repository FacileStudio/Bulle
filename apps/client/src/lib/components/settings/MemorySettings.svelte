<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Alert, Button, EmptyState, Icon, Skeleton, icons } from '@facile/muse';
	import { listMemoryPages, memoryError, pageLabel, PROFILE_PAGE, type MemoryPage } from '$lib/memory';
	import { formatDate } from '$lib/i18n/format';
	import { spaces } from '$lib/spaces.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import MyceliumConnect from '$lib/components/MyceliumConnect.svelte';
	import MemoryPageModal from './MemoryPageModal.svelte';

	let pages = $state<MemoryPage[] | null>(null);
	let available = $state(true);
	let failure = $state<string | null>(null);
	let attempt = $state(0);
	let opened = $state<string | null>(null);
	let modalOpen = $state(false);

	$effect(() => {
		void attempt;
		const space = spaces.currentId;
		let stale = false;
		pages = null;
		failure = null;
		listMemoryPages(space)
			.then((list) => {
				if (stale) return;
				available = list.available;
				pages = list.pages;
			})
			.catch((err: unknown) => {
				if (stale) return;
				const reason = memoryError(err, t('memory.error.load'));
				available = !reason.connect;
				failure = reason.connect ? null : reason.text;
				pages = [];
			});
		return () => {
			stale = true;
		};
	});

	function open(path: string) {
		opened = path;
		modalOpen = true;
	}

	function removed(path: string) {
		pages = (pages ?? []).filter((p) => p.path !== path);
	}

	function updatedAt(value: string): string {
		return value ? formatDate(value, { dateStyle: 'medium' }) : '';
	}
</script>

<section class="flex flex-col gap-6" aria-labelledby="memory-title">
	<div class="flex flex-col gap-1">
		<h2 id="memory-title" class="text-fc-lg font-semibold text-fc-fg">{t('memory.title')}</h2>
		<p class="text-fc-sm text-fc-fg-muted">
			{spaces.current ? t('memory.subtitleSpace', { space: spaces.current.name }) : t('memory.subtitle')}
		</p>
	</div>

	{#if failure}
		<Alert tone="danger">
			<div class="flex flex-col items-start gap-3">
				<span>{failure}</span>
				<Button variant="ghost" icon={icons.refresh} class="h-11" onclick={() => (attempt += 1)}>
					{t('common.tryAgain')}
				</Button>
			</div>
		</Alert>
	{:else if !available}
		<MyceliumConnect />
	{:else if pages === null}
		<div class="flex flex-col gap-2" aria-busy="true" aria-label={t('memory.loading')}>
			<Skeleton class="h-14 w-full" />
			<Skeleton class="h-14 w-full" />
			<Skeleton class="h-14 w-full" />
		</div>
	{:else if pages.length === 0}
		<EmptyState icon={icons.layers} title={t('memory.empty.title')} description={t('memory.empty.description')} />
	{:else}
		<ul class="flex flex-col gap-2">
			{#each pages as page (page.path)}
				<li>
					<button
						type="button"
						class="flex min-h-14 w-full items-center gap-3 rounded-fc-md bg-fc-surface px-4 py-3 text-left transition-colors hover:bg-fc-component focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
						aria-label={t('memory.openAria', { page: pageLabel(page.path) })}
						onclick={() => open(page.path)}
					>
						<Icon icon={page.path === PROFILE_PAGE ? icons.userCircle : CHAT_ICONS.file} size={18} class="shrink-0 text-fc-fg-muted" />
						<span class="flex min-w-0 flex-1 flex-col">
							<span class="truncate text-fc-sm font-medium text-fc-fg">{pageLabel(page.path)}</span>
							<span class="truncate text-fc-xs text-fc-fg-muted">
								{page.path}{#if updatedAt(page.updated)} · {updatedAt(page.updated)}{/if}
							</span>
						</span>
						<Icon icon={CHAT_ICONS.chevronRight} size={16} class="shrink-0 text-fc-fg-muted" />
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</section>

{#if opened}
	<MemoryPageModal bind:open={modalOpen} path={opened} spaceId={spaces.currentId} ondelete={removed} />
{/if}
