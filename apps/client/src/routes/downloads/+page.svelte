<script lang="ts">
	import { Badge, Button, Card, Icon, icons } from '@facile/muse';
	import InstallCommands from '$lib/components/InstallCommands.svelte';
	import PublicShell from '$lib/components/PublicShell.svelte';
	import { APK_SHA256_URL, APK_URL, RELEASES_URL } from '$lib/downloads';
	import { CHAT_ICONS } from '$lib/icons';
</script>

<svelte:head>
	<title>Downloads · Bulle</title>
	<meta
		name="description"
		content="Get Bulle on Android, in your terminal, or in the browser. A desktop app is coming soon."
	/>
</svelte:head>

<PublicShell>
	<section class="mx-auto max-w-5xl px-6 pt-16 pb-10 text-center sm:pt-20">
		<h1 class="font-fc-title text-3xl font-semibold tracking-tight text-fc-fg sm:text-4xl">
			Downloads
		</h1>
		<p class="mx-auto mt-4 max-w-xl text-fc-sm text-fc-fg-muted sm:text-fc-base">
			Bulle runs on your phone, in your terminal and in the browser. Every app signs in with your
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
					<div class="flex items-center gap-2">
						<h2 class="font-fc-title text-fc-lg font-semibold text-fc-fg">Desktop app</h2>
						<Badge tone="neutral">Coming soon</Badge>
					</div>
					<span class="text-fc-xs text-fc-fg-muted">macOS, Windows and Linux</span>
				</div>
			</div>
			<p class="text-fc-sm text-fc-fg-muted">
				A native desktop app is on the way. Until it ships, use the web app in your browser or the
				terminal app below.
			</p>
			<div class="mt-auto">
				<Button href="/login" variant="outline" iconRight={icons.arrow} class="rounded-fc-md">
					Open the web app
				</Button>
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
