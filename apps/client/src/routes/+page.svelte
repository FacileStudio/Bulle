<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Badge, Button, Card, Divider, Icon, Spinner, TextElevate, icons } from '@facile/muse';
	import NativeIntro from '$lib/components/native/NativeIntro.svelte';
	import InstallCommands from '$lib/components/InstallCommands.svelte';
	import PublicShell from '$lib/components/PublicShell.svelte';
	import { APK_URL } from '$lib/downloads';
	import { CHAT_ICONS } from '$lib/icons';
	import { isNative } from '$lib/platform';
	import { loadSession, session } from '$lib/session.svelte';

	interface Feature {
		icon: string;
		title: string;
		description: string;
	}

	const features: Feature[] = [
		{
			icon: icons.server,
			title: 'Dedicated Sandbox VM',
			description:
				'Its own computer: every user gets a persistent virtual machine to run commands, edit files, and build safely.'
		},
		{
			icon: icons.bolt,
			title: 'Multi-Model & OpenUI',
			description:
				'Anthropic, Google, and OpenRouter behind one interface, with interactive UI rendered directly inside replies.'
		},
		{
			icon: icons.monitor,
			title: 'Dual Workspace',
			description:
				'A rich Bubble Tea terminal TUI and a modern Svelte 5 web workspace. Work from anywhere seamlessly.'
		}
	];

	const INTRO_SEEN_KEY = 'bulle.intro-seen';
	const native = isNative();

	let ready = $state(false);

	function introSeen(): boolean {
		try {
			return localStorage.getItem(INTRO_SEEN_KEY) === '1';
		} catch {
			return false;
		}
	}

	function markIntroSeen() {
		try {
			localStorage.setItem(INTRO_SEEN_KEY, '1');
		} catch {
			return;
		}
	}

	async function finishIntro() {
		markIntroSeen();
		await goto('/login', { replaceState: true });
	}

	onMount(async () => {
		await loadSession();
		if (session.status === 'signed-in') {
			await goto('/sessions', { replaceState: true });
			return;
		}
		if (native && introSeen()) {
			await goto('/login', { replaceState: true });
			return;
		}
		ready = true;
	});
</script>

<svelte:head>
	<title>Bulle — Coding agent with its own computer</title>
	<meta
		name="description"
		content="Bulle is an autonomous coding agent with a persistent sandbox VM per user, multi-model intelligence, and terminal + web workspaces."
	/>
</svelte:head>

{#if !ready}
	<div class="flex min-h-dvh items-center justify-center bg-fc-page">
		<Spinner />
	</div>
{:else if native}
	<NativeIntro onfinish={finishIntro} />
{:else}
	<PublicShell>
		<section class="mx-auto max-w-5xl px-6 py-20 text-center sm:py-24">
			<h1
				class="font-fc-title text-4xl leading-[1.1] font-semibold tracking-tight text-fc-fg sm:text-5xl lg:text-6xl"
			>
				<span class="block"><TextElevate text="Your coding agent." delay={0.1} /></span>
				<span class="block"><TextElevate text="With its own computer." delay={0.25} /></span>
			</h1>
			<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted sm:text-fc-lg">
				Autonomous coding agent with a persistent sandbox VM per user. Multi-model support,
				interactive OpenUI components, and seamless terminal and web workspaces.
			</p>
			<div class="mt-10 flex flex-col items-stretch justify-center gap-3 sm:flex-row">
				<Button href="/login" size="lg" iconRight={icons.arrow} class="rounded-fc-md">
					Sign in with Facile
				</Button>
				<Button href="/downloads" variant="outline" size="lg" icon={icons.download} class="rounded-fc-md">
					Downloads
				</Button>
			</div>
		</section>

		<div class="mx-auto max-w-5xl px-6"><Divider class="my-0" /></div>

		<section class="mx-auto max-w-5xl px-6 py-16 sm:py-20">
			<div class="grid gap-6 md:grid-cols-3">
				{#each features as feature (feature.title)}
					<Card class="flex flex-col gap-4">
						<div
							class="flex size-10 items-center justify-center rounded-fc-md border border-fc-border bg-fc-surface text-fc-fg"
						>
							<Icon icon={feature.icon} size={20} />
						</div>
						<div class="flex flex-col gap-1.5">
							<h3 class="font-fc-title text-fc-lg font-semibold text-fc-fg">{feature.title}</h3>
							<p class="text-fc-sm text-fc-fg-muted">{feature.description}</p>
						</div>
					</Card>
				{/each}
			</div>
		</section>

		<div class="mx-auto max-w-5xl px-6"><Divider class="my-0" /></div>

		<section class="mx-auto max-w-5xl px-6 py-16 sm:py-20">
			<div class="text-center">
				<h2 class="font-fc-title text-2xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
					Install in seconds
				</h2>
				<p class="mx-auto mt-3 max-w-xl text-fc-sm text-fc-fg-muted">
					Run Bulle locally from your favorite terminal or manage it with Facile.
				</p>
			</div>
			<div class="mx-auto mt-8 max-w-2xl">
				<Card>
					<InstallCommands />
				</Card>
			</div>
		</section>

		<div class="mx-auto max-w-5xl px-6"><Divider class="my-0" /></div>

		<section class="mx-auto max-w-5xl px-6 py-16 sm:py-20">
			<div class="text-center">
				<h2 class="font-fc-title text-2xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
					Take Bulle with you
				</h2>
				<p class="mx-auto mt-3 max-w-xl text-fc-sm text-fc-fg-muted">
					The Android app keeps your sessions in your pocket. A desktop app is coming soon.
				</p>
			</div>
			<div class="mx-auto mt-8 grid max-w-2xl gap-4 sm:grid-cols-2">
				<Card class="flex flex-col gap-3">
					<div class="flex items-center gap-2 text-fc-fg">
						<Icon icon={CHAT_ICONS.phone} size={18} />
						<h3 class="font-fc-title text-fc-md font-semibold">Android</h3>
					</div>
					<p class="text-fc-sm text-fc-fg-muted">Signed APK, Android 7 or later.</p>
					<div class="mt-auto">
						<Button href={APK_URL} size="sm" icon={icons.download} class="rounded-fc-md">
							Download APK
						</Button>
					</div>
				</Card>
				<Card class="flex flex-col gap-3">
					<div class="flex flex-wrap items-center gap-2 text-fc-fg">
						<Icon icon={icons.monitor} size={18} />
						<h3 class="font-fc-title text-fc-md font-semibold">Desktop</h3>
						<Badge tone="neutral">Coming soon</Badge>
					</div>
					<p class="text-fc-sm text-fc-fg-muted">macOS, Windows and Linux.</p>
					<div class="mt-auto">
						<Button href="/downloads" size="sm" variant="outline" class="rounded-fc-md">
							All downloads
						</Button>
					</div>
				</Card>
			</div>
		</section>

		<div class="mx-auto max-w-5xl px-6"><Divider class="my-0" /></div>

		<section class="mx-auto max-w-5xl px-6 py-16 text-center sm:py-20">
			<h2 class="font-fc-title text-2xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
				Ready to code with Bulle?
			</h2>
			<p class="mx-auto mt-4 max-w-xl text-fc-sm text-fc-fg-muted sm:text-fc-base">
				Sign in with your Facile account to get started.
			</p>
			<div class="mt-8 flex justify-center">
				<Button href="/login" size="lg" icon={icons.key} class="rounded-fc-md">
					Sign in with Facile
				</Button>
			</div>
		</section>
	</PublicShell>
{/if}
