<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button, Card, Divider, Icon, Spinner, TextElevate, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import { loadSession, session } from '$lib/session.svelte';

	interface Feature {
		icon: string;
		title: string;
		description: string;
	}

	interface InstallMethod {
		label: string;
		command: string;
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

	const installMethods: InstallMethod[] = [
		{
			label: 'Install script',
			command: 'curl -fsSL https://raw.githubusercontent.com/FacileStudio/Bulle/main/install.sh | bash'
		},
		{
			label: 'With Facile CLI',
			command: 'facile install bulle'
		},
		{
			label: 'With Homebrew',
			command: 'brew install FacileStudio/tap/bulle'
		}
	];

	let ready = $state(false);
	let copiedIndex = $state<number | null>(null);

	async function copyCommand(command: string, index: number): Promise<void> {
		try {
			await navigator.clipboard.writeText(command);
			copiedIndex = index;
			setTimeout(() => {
				if (copiedIndex === index) {
					copiedIndex = null;
				}
			}, 2000);
		} catch {}
	}

	onMount(async () => {
		await loadSession();
		if (session.status === 'signed-in') {
			await goto('/sessions', { replaceState: true });
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
{:else}
	<div class="flex min-h-dvh flex-col bg-fc-page text-fc-fg">
		<header class="border-b border-fc-border">
			<div class="mx-auto flex max-w-5xl items-center justify-between gap-4 px-6 py-4">
				<a href="/" class="flex items-center gap-2.5">
					<Logo size={28} />
					<span class="font-fc-title text-fc-xl font-semibold tracking-tight text-fc-fg">Bulle</span>
				</a>
				<div class="flex items-center gap-2">
					<Button href="/login" variant="ghost" size="sm" class="rounded-fc-md">Log in</Button>
					<Button href="/login" size="sm" class="rounded-fc-md">Sign in with Facile</Button>
				</div>
			</div>
		</header>

		<main class="flex-1">
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
					<Button href="/login" variant="outline" size="lg" class="rounded-fc-md">Log in</Button>
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
					<Card class="flex flex-col gap-4">
						{#each installMethods as method, index (method.command)}
							<div class="flex flex-col gap-1.5">
								<span class="text-fc-xs font-medium text-fc-fg-muted">{method.label}</span>
								<div
									class="flex items-center justify-between gap-3 rounded-fc-md border border-fc-border bg-fc-surface px-3.5 py-2.5 font-fc-mono text-fc-sm text-fc-fg"
								>
									<div class="flex items-center gap-2 overflow-x-auto">
										<span class="select-none text-fc-fg-muted">$</span>
										<span class="whitespace-nowrap">{method.command}</span>
									</div>
									<button
										type="button"
										class="inline-flex shrink-0 items-center justify-center rounded-fc-sm p-1 text-fc-fg-muted transition-colors hover:bg-fc-page hover:text-fc-fg"
										aria-label="Copy command"
										onclick={() => copyCommand(method.command, index)}
									>
										{#if copiedIndex === index}
											<Icon icon={icons.check} size={16} />
										{:else}
											<Icon icon={icons.copy} size={16} />
										{/if}
									</button>
								</div>
							</div>
						{/each}
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
		</main>

		<footer class="border-t border-fc-border">
			<div
				class="mx-auto flex max-w-5xl flex-col items-center justify-between gap-2 px-6 py-6 text-fc-xs text-fc-fg-muted sm:flex-row"
			>
				<p>
					© {new Date().getFullYear()} Bulle by{' '}
					<a
						href="https://facile.studio"
						target="_blank"
						rel="noopener"
						class="font-medium underline underline-offset-2 transition-colors hover:text-fc-fg"
					>
						Facile
					</a>.
				</p>
				<p>Part of the Facile Suite.</p>
			</div>
		</footer>
	</div>
{/if}
