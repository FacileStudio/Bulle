<script lang="ts">
	import { onMount } from 'svelte';
	import { Button, Card, Icon, icons } from '@facile/muse';
	import InstallCommands from '$lib/components/InstallCommands.svelte';
	import PublicShell from '$lib/components/PublicShell.svelte';
	import {
		APK_SHA256_URL,
		APK_URL,
		DESKTOP_BUILDS,
		DESKTOP_CHECKSUMS_URL,
		RELEASES_URL,
		detectDesktopOs,
		type DesktopOs
	} from '$lib/downloads';
	import { CHAT_ICONS } from '$lib/icons';

	let os = $state<DesktopOs | null>(null);
	const primary = $derived(DESKTOP_BUILDS.find((build) => build.os === os));

	onMount(() => {
		os = detectDesktopOs(navigator.userAgent);
	});
</script>

<svelte:head>
	<title>Downloads · Bulle</title>
	<meta
		name="description"
		content="Get Bulle for macOS, Windows, Linux, Android, or your terminal."
	/>
</svelte:head>

<PublicShell>
	<section class="mx-auto max-w-5xl px-6 pt-16 pb-10 text-center sm:pt-20">
		<h1 class="font-fc-title text-3xl font-semibold tracking-tight text-fc-fg sm:text-4xl">
			Downloads
		</h1>
		<p class="mx-auto mt-4 max-w-xl text-fc-sm text-fc-fg-muted sm:text-fc-base">
			Bulle runs on your desktop, your phone, in your terminal and in the browser. Every app signs in with your
			Facile account and shares the same sessions.
		</p>
	</section>

	<section class="mx-auto grid max-w-5xl grid-cols-1 gap-4 px-6 pb-16 sm:pb-20 md:grid-cols-2">
		<Card class="flex min-w-0 flex-col gap-4">
			<div class="flex items-center gap-3">
				<span
					class="flex size-10 items-center justify-center rounded-fc-md border border-fc-border bg-fc-surface text-fc-fg"
				>
					<Icon icon={CHAT_ICONS.phone} size={20} />
				</span>
				<div class="flex flex-col">
					<h2 class="font-fc-title text-fc-lg font-semibold text-fc-fg">Android</h2>
					<span class="text-fc-xs text-fc-fg-muted">Android 7 or later · signed APK</span>
				</div>
			</div>
			<p class="text-fc-sm text-fc-fg-muted">
				When Android asks, allow installs from your browser or file manager. To update, download the
				new APK and install it over the old one. Your data stays, because every release is signed
				with the same key.
			</p>
			<div class="mt-auto flex flex-wrap items-center gap-3">
				<Button href={APK_URL} icon={icons.download} class="rounded-fc-md">Download APK</Button>
				<a
					href={APK_SHA256_URL}
					class="text-fc-xs text-fc-fg-muted underline underline-offset-2 transition-colors hover:text-fc-fg"
				>
					SHA-256 checksum
				</a>
			</div>
		</Card>

		<Card class="flex min-w-0 flex-col gap-4">
			<div class="flex items-center gap-3">
				<span
					class="flex size-10 items-center justify-center rounded-fc-md border border-fc-border bg-fc-surface text-fc-fg"
				>
					<Icon icon={icons.monitor} size={20} />
				</span>
				<div class="flex flex-col">
					<h2 class="font-fc-title text-fc-lg font-semibold text-fc-fg">Desktop app</h2>
					<span class="text-fc-xs text-fc-fg-muted">macOS, Windows and Linux · updates itself</span>
				</div>
			</div>
			<p class="text-fc-sm text-fc-fg-muted">
				The installers are not notarized yet. On macOS, right-click Bulle and choose Open the first
				time. On Windows, choose More info, then Run anyway.
			</p>
			<div class="mt-auto flex flex-col gap-3">
				{#if primary}
					<Button href={primary.url} icon={icons.download} class="rounded-fc-md">
						Download for {primary.label}
					</Button>
				{/if}
				<ul class="grid gap-2 sm:grid-cols-2">
					{#each DESKTOP_BUILDS.filter((build) => build !== primary) as build (build.url)}
						<li>
							<a
								href={build.url}
								class="flex min-h-11 items-center gap-2 rounded-fc-md px-3 py-2 text-fc-sm text-fc-fg transition-colors hover:bg-fc-surface"
							>
								<Icon icon={icons.download} size={16} />
								<span class="font-medium">{build.label}</span>
								<span class="text-fc-xs text-fc-fg-muted">{build.detail}</span>
							</a>
						</li>
					{/each}
				</ul>
				<a
					href={DESKTOP_CHECKSUMS_URL}
					class="text-fc-xs text-fc-fg-muted underline underline-offset-2 transition-colors hover:text-fc-fg"
				>
					SHA-256 checksums
				</a>
			</div>
		</Card>

		<Card class="flex min-w-0 flex-col gap-4 md:col-span-2">
			<div class="flex items-center gap-3">
				<span
					class="flex size-10 items-center justify-center rounded-fc-md border border-fc-border bg-fc-surface text-fc-fg"
				>
					<Icon icon={CHAT_ICONS.terminal} size={20} />
				</span>
				<div class="flex flex-col">
					<h2 class="font-fc-title text-fc-lg font-semibold text-fc-fg">Terminal</h2>
					<span class="text-fc-xs text-fc-fg-muted">macOS and Linux · then run bulle login</span>
				</div>
			</div>
			<InstallCommands />
		</Card>

		<p class="text-center text-fc-xs text-fc-fg-muted md:col-span-2">
			Release notes and older versions are on
			<a
				href={RELEASES_URL}
				target="_blank"
				rel="noopener"
				class="font-medium underline underline-offset-2 transition-colors hover:text-fc-fg">GitHub</a
			>.
		</p>
	</section>
</PublicShell>
