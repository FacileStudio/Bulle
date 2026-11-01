<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button, Card, Spinner, Stack, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import { HOME } from '$lib/nav';
	import { loadSession, session } from '$lib/session.svelte';

	const features = [
		'A coding agent that reasons, plans and works in the open',
		'Its own computer: a persistent virtual machine per user',
		'Anthropic, Google, OpenAI and OpenRouter behind one interface',
		'Read and edit files, run commands, fetch web pages, call tools',
		'Parallel subagents for wide searches and long logs',
		'Project context from CLAUDE.md and AGENTS.md, plus agent skills'
	];

	const installs = [
		{
			label: 'Install the terminal app',
			command: 'curl -fsSL https://raw.githubusercontent.com/FacileStudio/Bulle/main/install.sh | bash'
		},
		{ label: 'With facile', command: 'facile install bulle' },
		{ label: 'With Homebrew', command: 'brew install FacileStudio/tap/bulle' }
	];

	onMount(async () => {
		await loadSession();
		if (session.status === 'signed-in') await goto(HOME, { replaceState: true });
	});
</script>

<svelte:head>
	<title>Bulle - a coding agent with its own computer</title>
</svelte:head>

{#if session.status === 'loading'}
	<div class="flex min-h-dvh items-center justify-center bg-fc-page">
		<Spinner />
	</div>
{:else}
	<div class="flex min-h-dvh flex-col bg-fc-page px-4 py-8 sm:px-6">
		<main class="mx-auto flex w-full max-w-fc-md flex-1 flex-col justify-center gap-10">
			<Card>
				<Stack gap="section">
					<Stack gap="content">
						<div class="flex items-center gap-3 text-fc-fg">
							<Logo size={32} />
							<span class="font-fc-title text-fc-xl font-semibold">Bulle</span>
						</div>
						<h1 class="font-fc-title text-fc-2xl font-semibold text-fc-fg">
							A coding agent with its own computer
						</h1>
						<p class="text-fc-md text-fc-fg-muted">
							Watch your agent reason, plan and work, in your browser or in your terminal.
						</p>
						<div>
							<Button href="/login" icon={icons.key} size="lg">Sign in with Facile</Button>
						</div>
					</Stack>

					<Stack gap="content">
						<h2 class="font-fc-title text-fc-lg font-semibold text-fc-fg-muted">What is inside</h2>
						<ul class="flex flex-col gap-2 text-fc-sm text-fc-fg">
							{#each features as feature (feature)}
								<li>{feature}</li>
							{/each}
						</ul>
					</Stack>

					<Stack gap="content">
						<h2 class="font-fc-title text-fc-lg font-semibold text-fc-fg-muted">Terminal app</h2>
						{#each installs as install (install.command)}
							<Stack gap="bound">
								<p class="text-fc-xs text-fc-fg-muted">{install.label}</p>
								<pre
									class="overflow-x-auto rounded-fc-md bg-fc-surface p-3 font-fc-mono text-fc-sm text-fc-fg">{install.command}</pre>
							</Stack>
						{/each}
					</Stack>
				</Stack>
			</Card>
			<a
				href="https://github.com/FacileStudio/Bulle"
				class="text-fc-sm text-fc-fg hover:underline"
			>
				github.com/FacileStudio/Bulle
			</a>
		</main>
	</div>
{/if}
