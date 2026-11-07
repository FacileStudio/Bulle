<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { onMount } from 'svelte';
	import { Button, Icon, icons } from '@facile/muse';
	import InstallCommands from '$lib/components/InstallCommands.svelte';
	import {
		DESKTOP_CHECKSUMS_URL,
		PLATFORMS,
		RELEASES_URL,
		featuredPlatform,
		type Platform
	} from '$lib/downloads';
	import { CHAT_ICONS, PLATFORM_ICONS } from '$lib/icons';
	import { version } from '../../../package.json';

	const ICONS: Record<Platform, string> = {
		macos: PLATFORM_ICONS.apple,
		windows: PLATFORM_ICONS.windows,
		linux: PLATFORM_ICONS.linux,
		android: PLATFORM_ICONS.android
	};

	let featuredId = $state<Platform>('macos');
	let detected = $state(false);
	const featured = $derived(PLATFORMS.find((p) => p.id === featuredId) ?? PLATFORMS[0]);

	onMount(() => {
		featuredId = featuredPlatform(navigator.userAgent);
		detected = true;
	});
</script>

<div class="flex flex-col gap-10">
	<section
		class="flex flex-col items-center gap-6 rounded-fc-lg bg-fc-component px-6 py-10 text-center sm:px-10"
	>
		<span
			class="flex size-16 items-center justify-center rounded-fc-lg bg-fc-page text-fc-fg shadow-sm"
		>
			<Icon icon={ICONS[featured.id]} size={32} />
		</span>
		<div class="flex flex-col gap-2">
			<h2 class="font-fc-title text-fc-2xl font-semibold tracking-tight text-fc-fg">
				{t('downloads.for', { platform: featured.label })}
			</h2>
			<p class="text-fc-sm text-fc-fg-muted">
				{t('downloads.version', { version })} · {featured.summary}{featured.id === 'android'
					? ''
					: ` · ${t('downloads.updatesItself')}`}
			</p>
		</div>
		<div class="flex flex-wrap items-center justify-center gap-3">
			{#each featured.files as file, i (file.url)}
				<Button
					href={file.url}
					size="lg"
					variant={i === 0 ? 'primary' : 'outline'}
					icon={icons.download}
					class="rounded-fc-md"
				>
					{i === 0 ? t('downloads.downloadFor', { platform: featured.label }) : file.label}
				</Button>
			{/each}
		</div>
		{#if detected}
			<p class="text-fc-xs text-fc-fg-subtle">{t('downloads.picked')}</p>
		{/if}
	</section>

	<section class="flex flex-col gap-4">
		<h2 class="font-fc-title text-fc-lg font-semibold text-fc-fg">{t('downloads.all')}</h2>
		<ul class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
			{#each PLATFORMS as platform (platform.id)}
				<li
					class="flex flex-col gap-4 rounded-fc-lg bg-fc-component p-5"
				>
					<div class="flex items-center gap-3">
						<span class="flex size-10 items-center justify-center rounded-fc-md bg-fc-page text-fc-fg">
							<Icon icon={ICONS[platform.id]} size={20} />
						</span>
						<div class="flex min-w-0 flex-col">
							<h3 class="text-fc-sm font-semibold text-fc-fg">{platform.label}</h3>
							<span class="text-fc-xs text-fc-fg-muted">{platform.summary}</span>
						</div>
					</div>
					<div class="mt-auto flex flex-wrap gap-2">
						{#each platform.files as file (file.url)}
							<a
								href={file.url}
								class="inline-flex min-h-9 items-center gap-1.5 rounded-fc-md bg-fc-page px-3 text-fc-xs font-medium text-fc-fg transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
							>
								<Icon icon={icons.download} size={14} />
								{file.label}
							</a>
						{/each}
					</div>
				</li>
			{/each}
		</ul>
	</section>

	<section class="flex flex-col gap-4 rounded-fc-lg bg-fc-component p-5 sm:p-6">
		<div class="flex items-center gap-3">
			<span class="flex size-10 items-center justify-center rounded-fc-md bg-fc-page text-fc-fg">
				<Icon icon={CHAT_ICONS.terminal} size={20} />
			</span>
			<div class="flex flex-col">
				<h2 class="text-fc-sm font-semibold text-fc-fg">{t('downloads.terminal')}</h2>
				<span class="text-fc-xs text-fc-fg-muted">{t('downloads.terminalHint')}</span>
			</div>
		</div>
		<InstallCommands />
	</section>

	<section class="grid gap-4 text-fc-xs leading-relaxed text-fc-fg-muted sm:grid-cols-2">
		<p>
			{t('downloads.notarized')}
		</p>
		<p class="sm:text-right">
			<a
				href={DESKTOP_CHECKSUMS_URL}
				class="underline underline-offset-2 transition-colors hover:text-fc-fg">{t('downloads.checksums')}</a
			>
			·
			<a
				href={RELEASES_URL}
				target="_blank"
				rel="noopener"
				class="underline underline-offset-2 transition-colors hover:text-fc-fg">{t('downloads.releaseNotes')}</a
			>
		</p>
	</section>
</div>
