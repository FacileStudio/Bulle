<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { onDestroy, untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Alert, Badge, Button, Card, EmptyState, Icon, Input, PageHeader, Select, Skeleton, Spinner, Stack, icons } from '@facile/muse';
	import {
		artifactError,
		artifactHref,
		artifactSpaceName,
		artifactSpaces,
		inSpace,
		isEditableTarget,
		listArtifacts,
		normaliseQuery,
		searchArtifacts,
		snippetParts,
		type Artifact,
		type ArtifactHit,
		type ArtifactList
	} from '$lib/artifacts';
	import { relativeTime } from '$lib/chat/format';
	import { SEARCH_DELAY_MS, debounce } from '$lib/chat/search';
	import MyceliumConnect from '$lib/components/MyceliumConnect.svelte';
	import { ARTIFACT_ICON } from '$lib/suite-icons';

	const SEARCH_ID = 'artifact-search';

	type Found = { query: string; hits: ArtifactHit[]; next: string };

	let list = $state<ArtifactList | null>(null);
	let error = $state<string | null>(null);
	let query = $state(untrack(() => page.url.searchParams.get('q') ?? ''));
	let space = $state('');
	let attempt = $state(0);
	let found = $state<Found | null>(null);
	let searching = $state(false);
	let loadingMore = $state(false);
	let searchError = $state<string | null>(null);
	let latest = 0;

	const q = $derived(normaliseQuery(query));
	const spaceOptions = $derived(artifactSpaces(list?.artifacts ?? []));
	const newest = $derived(
		[...(list?.artifacts ?? [])].sort((a, b) => Date.parse(b.updated) - Date.parse(a.updated))
	);
	const current = $derived(found && found.query === q ? found : null);
	const shown = $derived<(Artifact | ArtifactHit)[]>(q ? inSpace(current?.hits ?? [], space) : inSpace(newest, space));

	$effect(() => {
		void attempt;
		let stale = false;
		error = null;
		list = null;
		listArtifacts()
			.then((result) => {
				if (!stale) list = result;
			})
			.catch((err: unknown) => {
				if (stale) return;
				const failure = artifactError(err);
				if (failure.connect) list = { artifacts: [], available: false };
				else error = failure.text;
			});
		return () => {
			stale = true;
		};
	});

	function keepInUrl(text: string) {
		const url = new URL(page.url);
		if ((url.searchParams.get('q') ?? '') === text) return;
		if (text) url.searchParams.set('q', text);
		else url.searchParams.delete('q');
		void goto(url, { replaceState: true, keepFocus: true, noScroll: true });
	}

	async function find(text: string, cursor = '') {
		const mine = ++latest;
		searchError = null;
		try {
			const result = await searchArtifacts(text, cursor);
			if (mine !== latest) return;
			const before = cursor && found?.query === text ? found.hits : [];
			found = { query: text, hits: [...before, ...result.artifacts], next: result.next_cursor };
		} catch (err) {
			if (mine === latest) searchError = artifactError(err).text;
		} finally {
			if (mine === latest) {
				searching = false;
				loadingMore = false;
			}
		}
	}

	const run = debounce((text: string) => {
		keepInUrl(text);
		void find(text);
	}, SEARCH_DELAY_MS);

	$effect(() => {
		const text = q;
		untrack(() => {
			searchError = null;
			if (!text) {
				run.cancel();
				latest += 1;
				found = null;
				searching = false;
				loadingMore = false;
				keepInUrl('');
				return;
			}
			searching = true;
			run.call(text);
		});
	});

	onDestroy(() => run.cancel());

	function more() {
		if (!current?.next || loadingMore) return;
		loadingMore = true;
		void find(current.query, current.next);
	}

	function onSearchKey(event: KeyboardEvent) {
		if (event.key !== 'Escape' || !query) return;
		event.preventDefault();
		event.stopPropagation();
		query = '';
	}

	function onWindowKey(event: KeyboardEvent) {
		if (event.key !== '/' || event.metaKey || event.ctrlKey || event.altKey || isEditableTarget(event.target)) return;
		const field = document.getElementById(SEARCH_ID);
		if (!field) return;
		event.preventDefault();
		field.focus();
	}

	function snippetOf(artifact: Artifact | ArtifactHit): string {
		return 'snippet' in artifact ? artifact.snippet : '';
	}
</script>

<svelte:head>
	<title>{t('artifacts.pageTitle')}</title>
</svelte:head>

<svelte:window onkeydown={onWindowKey} />

<Stack gap="section">
	<PageHeader title={t('artifacts.title')} description={t('artifacts.subtitle')} />

	{#if error}
		<Alert tone="danger" title={t('artifacts.loadError')}>
			<div class="flex flex-col items-start gap-3">
				<span>{error}</span>
				<Button variant="ghost" icon={icons.refresh} class="h-11" onclick={() => (attempt += 1)}>{t('common.tryAgain')}</Button>
			</div>
		</Alert>
	{:else if !list}
		<div class="flex flex-col gap-4" aria-busy="true" aria-label={t('artifacts.loading')}>
			<Skeleton class="h-11 w-full" />
			<div class="grid grid-cols-[repeat(auto-fill,minmax(min(18rem,100%),1fr))] gap-4">
				<Skeleton class="h-32 w-full" />
				<Skeleton class="h-32 w-full" />
				<Skeleton class="h-32 w-full" />
			</div>
		</div>
	{:else if !list.available}
		<MyceliumConnect />
	{:else if list.artifacts.length === 0}
		<EmptyState icon={ARTIFACT_ICON} title={t('artifacts.empty.title')} description={t('artifacts.empty.description')} />
	{:else}
		<div class="flex flex-col gap-6">
			<div class="flex flex-col gap-2 sm:flex-row">
				<div class="relative min-w-0 flex-1">
					<Input
						id={SEARCH_ID}
						type="search"
						bind:value={query}
						onkeydown={onSearchKey}
						class="bg-fc-component pr-12"
						placeholder={t('artifacts.search.placeholder')}
						aria-label={t('artifacts.search.label')}
						aria-keyshortcuts="/"
						aria-controls="artifact-results"
						autocomplete="off"
						maxlength={200}
					/>
					{#if searching}
						<span class="pointer-events-none absolute inset-y-0 right-4 flex items-center">
							<Spinner size="sm" label={t('artifacts.searching')} />
						</span>
					{:else if !query}
						<kbd class="pointer-events-none absolute inset-y-0 right-4 hidden items-center font-fc-mono text-fc-xs text-fc-fg-muted sm:flex">/</kbd>
					{/if}
				</div>
				{#if spaceOptions.length > 1}
					<Select
						bind:value={space}
						class="bg-fc-component"
						wrapperClass="sm:w-56 sm:shrink-0"
						aria-label={t('artifacts.filter.space')}
						options={[
							{ value: '', label: t('artifacts.filter.allSpaces') },
							...spaceOptions.map((s) => ({ value: s.key, label: s.name }))
						]}
					/>
				{/if}
			</div>

			{#if searchError}
				<Alert tone="danger" title={t('artifacts.searchError')}>
					<div class="flex flex-col items-start gap-3">
						<span>{searchError}</span>
						<Button variant="ghost" icon={icons.refresh} class="h-11" onclick={() => void find(q)}>{t('common.tryAgain')}</Button>
					</div>
				</Alert>
			{/if}

			<div id="artifact-results" aria-live="polite" aria-busy={q !== '' && !current} class="flex flex-col gap-4">
				{#if q && !current}
					{#if !searchError}
						<div class="flex min-h-32 items-center justify-center">
							<Spinner label={t('artifacts.searching')} />
						</div>
					{/if}
				{:else if shown.length === 0}
					<EmptyState icon={icons.search} title={t('artifacts.noMatch', { query: q })} description={t('artifacts.noMatchHint')} />
				{:else}
					<ul class="grid grid-cols-[repeat(auto-fill,minmax(min(18rem,100%),1fr))] gap-4">
						{#each shown as artifact (`${artifact.mycelium_space ?? ''}/${artifact.id}`)}
							{@const parts = q ? snippetParts(snippetOf(artifact), q) : []}
							<li class="min-w-0">
								<Card href={artifactHref(artifact)} class="flex h-full flex-col gap-3 p-4">
									<span class="flex min-w-0 items-start gap-3">
										<span class="flex size-9 shrink-0 items-center justify-center rounded-fc-md bg-fc-accent/15 text-fc-accent">
											<Icon icon={ARTIFACT_ICON} size={18} />
										</span>
										<span class="line-clamp-2 min-w-0 flex-1 pt-1.5 text-fc-sm font-medium break-words text-fc-fg">{artifact.title}</span>
									</span>
									{#if parts.length}
										<p class="line-clamp-3 text-fc-xs leading-relaxed break-words text-fc-fg-muted">
											{#each parts as part, i (i)}
												{#if part.hit}<mark class="bg-transparent font-medium text-fc-fg">{part.text}</mark>{:else}{part.text}{/if}
											{/each}
										</p>
									{/if}
									<span class="mt-auto flex min-w-0 items-center gap-2 text-fc-xs text-fc-fg-muted">
										<time class="shrink-0" datetime={artifact.updated}>{relativeTime(artifact.updated)}</time>
										<span class="min-w-0 flex-1 truncate">{artifactSpaceName(artifact)}</span>
										{#if artifact.machine}
											<Badge tone={artifact.machine === 'bulle' ? 'accent' : 'neutral'} class="max-w-24 shrink-0">
												{artifact.machine}
											</Badge>
										{/if}
									</span>
								</Card>
							</li>
						{/each}
					</ul>
				{/if}
				{#if q && current?.next}
					<div class="flex justify-center">
						<Button variant="ghost" class="h-11" disabled={loadingMore} aria-busy={loadingMore} onclick={more}>
							{loadingMore ? t('artifacts.searching') : t('artifacts.loadMore')}
						</Button>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</Stack>
