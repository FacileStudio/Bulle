<script lang="ts">
	import { Alert, Badge, Button, Icon, Section, Spinner, icons } from '@facile/muse';
	import { updater } from '$lib/updater.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const isBusy = $derived(
		updater.state === 'checking' || updater.state === 'downloading'
	);

	const statusLabel = $derived.by(() => {
		switch (updater.state) {
			case 'checking':
				return t('updater.checking');
			case 'available':
				return t('updater.available', { version: updater.availableVersion ?? '' });
			case 'downloading':
				return t('updater.downloadingPercent', { percent: updater.downloadProgress });
			case 'ready':
				return t('updater.ready');
			case 'error':
				return t('updater.failed');
			case 'upToDate':
				return t('updater.upToDate');
			default:
				return t('updater.idle');
		}
	});

	const statusTone = $derived.by(() => {
		switch (updater.state) {
			case 'available':
				return 'accent' as const;
			case 'downloading':
			case 'checking':
				return 'info' as const;
			case 'ready':
			case 'upToDate':
				return 'success' as const;
			case 'error':
				return 'danger' as const;
			default:
				return 'neutral' as const;
		}
	});
</script>

<Section
	card
	id="updates"
	title={t('updater.title')}
	description={t('updater.description')}
>
	<div class="flex flex-col gap-5">
		{#if updater.state === 'error' && updater.errorMessage}
			<Alert tone="danger">{updater.errorMessage}</Alert>
		{/if}

		<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
			<div class="flex items-center gap-3">
				<div
					class="flex size-10 shrink-0 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg"
				>
					<Icon icon={icons.refresh} size={20} />
				</div>
				<div class="flex flex-col gap-0.5">
					<div class="flex items-center gap-2">
						<span class="font-mono text-fc-sm font-semibold text-fc-fg">
							v{updater.currentVersion}
						</span>
						<Badge tone={statusTone}>{statusLabel}</Badge>
					</div>
					<span class="text-fc-xs text-fc-fg-muted">{t('updater.currentBuild')}</span>
				</div>
			</div>

			<div class="flex items-center gap-2">
				{#if updater.state === 'available'}
					<Button
						variant="primary"
						size="sm"
						icon={icons.download}
						onclick={() => updater.installNow()}
					>
						{t('updater.install')}
					</Button>
				{:else if updater.state === 'ready'}
					<Button
						variant="primary"
						size="sm"
						icon={icons.refresh}
						onclick={() => updater.restartNow()}
					>
						{t('updater.restart')}
					</Button>
				{:else}
					<Button
						variant="outline"
						size="sm"
						icon={icons.refresh}
						disabled={isBusy}
						onclick={() => updater.checkNow(false)}
					>
						{#if updater.state === 'checking'}
							<Spinner size="sm" />
						{/if}
						{updater.state === 'checking' ? t('updater.checkingButton') : t('updater.check')}
					</Button>
				{/if}
			</div>
		</div>

		{#if updater.state === 'downloading'}
			<div class="flex flex-col gap-1.5">
				<div class="flex items-center justify-between text-fc-xs text-fc-fg-muted">
					<span>{t('updater.downloading')}</span>
					<span class="font-mono">{updater.downloadProgress}%</span>
				</div>
				<div class="h-2 w-full overflow-hidden rounded-fc-pill bg-fc-surface">
					<div
						class="h-full bg-fc-accent transition-all duration-200"
						style="width: {updater.downloadProgress}%"
					></div>
				</div>
			</div>
		{/if}

		<div class="flex items-center justify-between pt-4 text-fc-sm">
			<span class="text-fc-fg-muted">{t('updater.changelog')}</span>
			<a
				href="https://github.com/FacileStudio/Bulle/releases"
				target="_blank"
				rel="noreferrer"
				class="inline-flex items-center gap-1.5 text-fc-accent hover:underline"
			>
				<Icon icon={icons.history} size={16} />
				<span>{t('updater.releaseHistory')}</span>
			</a>
		</div>
	</div>
</Section>
