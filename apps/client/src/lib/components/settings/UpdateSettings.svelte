<script lang="ts">
	import { Alert, Badge, Button, Icon, Section, Spinner, icons } from '@facile/muse';
	import { updater } from '$lib/updater.svelte';

	const isBusy = $derived(
		updater.state === 'checking' || updater.state === 'downloading'
	);

	const statusLabel = $derived.by(() => {
		switch (updater.state) {
			case 'checking':
				return 'Checking for updates…';
			case 'available':
				return `Update v${updater.availableVersion} available`;
			case 'downloading':
				return `Downloading ${updater.downloadProgress}%`;
			case 'ready':
				return 'Update ready to install';
			case 'error':
				return 'Update failed';
			case 'upToDate':
				return 'Bulle is up to date';
			default:
				return 'Ready to check';
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
	title="Updates"
	description="Check for new Bulle desktop releases and install them."
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
					<span class="text-fc-xs text-fc-fg-muted">Current release build</span>
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
						Download & Install
					</Button>
				{:else if updater.state === 'ready'}
					<Button
						variant="primary"
						size="sm"
						icon={icons.refresh}
						onclick={() => updater.restartNow()}
					>
						Restart & Install
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
						{updater.state === 'checking' ? 'Checking…' : 'Check for updates'}
					</Button>
				{/if}
			</div>
		</div>

		{#if updater.state === 'downloading'}
			<div class="flex flex-col gap-1.5">
				<div class="flex items-center justify-between text-fc-xs text-fc-fg-muted">
					<span>Downloading update…</span>
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

		<div class="flex items-center justify-between border-t border-fc-border pt-4 text-fc-sm">
			<span class="text-fc-fg-muted">View full changelog and release history</span>
			<a
				href="https://github.com/FacileStudio/Bulle/releases"
				target="_blank"
				rel="noreferrer"
				class="inline-flex items-center gap-1.5 text-fc-accent hover:underline"
			>
				<Icon icon={icons.history} size={16} />
				<span>Release history</span>
			</a>
		</div>
	</div>
</Section>
