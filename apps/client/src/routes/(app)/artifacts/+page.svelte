<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Alert, Badge, Button, EmptyState, Icon, Input, PageHeader, Select, Skeleton, Stack, icons } from '@facile/muse';
	import {
		artifactError,
		artifactHref,
		artifactSpaceName,
		artifactSpaces,
		filterArtifacts,
		groupByMonth,
		inSpace,
		listArtifacts,
		type ArtifactList
	} from '$lib/artifacts';
	import { relativeTime } from '$lib/chat/format';
	import MyceliumConnect from '$lib/components/MyceliumConnect.svelte';

	let list = $state<ArtifactList | null>(null);
	let error = $state<string | null>(null);
	let query = $state('');
	let space = $state('');
	let attempt = $state(0);

	const spaceOptions = $derived(artifactSpaces(list?.artifacts ?? []));
	const months = $derived(groupByMonth(filterArtifacts(inSpace(list?.artifacts ?? [], space), query)));

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
</script>

<svelte:head>
	<title>{t('artifacts.pageTitle')}</title>
</svelte:head>

<Stack gap="section">
	<PageHeader
		title={t('artifacts.title')}
		description={t('artifacts.subtitle')}
	/>

	{#if error}
		<Alert tone="danger" title={t('artifacts.loadError')}>
			<div class="flex flex-col items-start gap-3">
				<span>{error}</span>
				<Button variant="ghost" icon={icons.refresh} class="h-11" onclick={() => (attempt += 1)}>{t('common.tryAgain')}</Button>
			</div>
		</Alert>
	{:else if !list}
		<div class="flex flex-col gap-2" aria-busy="true" aria-label={t('artifacts.loading')}>
			<Skeleton class="h-11 w-full" />
			<Skeleton class="h-14 w-full" />
			<Skeleton class="h-14 w-full" />
		</div>
	{:else if !list.available}
		<MyceliumConnect />
	{:else if list.artifacts.length === 0}
		<EmptyState
			icon={icons.layers}
			title={t('artifacts.empty.title')}
			description={t('artifacts.empty.description')}
		/>
	{:else}
		<div class="flex flex-col gap-6">
			<div class="flex flex-col gap-2 sm:flex-row">
				<Input
					type="search"
					bind:value={query}
					class="border-0 bg-fc-component"
					placeholder={t('artifacts.filter.placeholder')}
					aria-label={t('artifacts.filter.label')}
				/>
				{#if spaceOptions.length > 1}
					<Select
						bind:value={space}
						class="border-0 bg-fc-component"
						wrapperClass="sm:w-56 sm:shrink-0"
						aria-label={t('artifacts.filter.space')}
						options={[
							{ value: '', label: t('artifacts.filter.allSpaces') },
							...spaceOptions.map((s) => ({ value: s.key, label: s.name }))
						]}
					/>
				{/if}
			</div>
			{#each months as month (month.key)}
				<section class="flex flex-col gap-2" aria-labelledby="month-{month.key}">
					<h2 id="month-{month.key}" class="px-1 text-fc-xs font-semibold tracking-wide text-fc-fg-muted uppercase">
						{month.label}
					</h2>
					<ul class="flex flex-col gap-1">
						{#each month.items as artifact (`${artifact.mycelium_space ?? ''}/${artifact.id}`)}
							<li>
								<a
									href={artifactHref(artifact)}
									class="flex min-h-14 items-center gap-3 rounded-fc-md bg-fc-component px-4 py-2 transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
								>
									<Icon icon={icons.layers} size={18} class="text-fc-accent" />
									<span class="flex min-w-0 flex-1 flex-col">
										<span class="truncate text-fc-sm font-medium text-fc-fg">{artifact.title}</span>
										<span class="flex min-w-0 items-center gap-2 text-fc-xs text-fc-fg-muted">
											<span class="shrink-0">{relativeTime(artifact.updated)}</span>
											<span class="truncate">{artifactSpaceName(artifact)}</span>
										</span>
									</span>
									{#if artifact.machine}
										<Badge tone={artifact.machine === 'bulle' ? 'accent' : 'neutral'} class="max-w-24 shrink-0">
											{artifact.machine}
										</Badge>
									{/if}
									<Icon icon={icons.arrow} size={16} class="text-fc-fg-muted" />
								</a>
							</li>
						{/each}
					</ul>
				</section>
			{:else}
				<p class="px-1 text-fc-sm text-fc-fg-muted">{t('artifacts.noMatch', { query })}</p>
			{/each}
		</div>
	{/if}
</Stack>
