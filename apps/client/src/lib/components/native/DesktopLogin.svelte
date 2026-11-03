<script lang="ts">
	import { Icon, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import OnboardingAvatar from '$lib/components/onboarding/OnboardingAvatar.svelte';
	import SignInButton from './SignInButton.svelte';

	type Point = { icon: string; title: string; body: string };

	const POINTS: Point[] = [
		{
			icon: icons.mail,
			title: 'Works across your suite',
			body: 'It reads your mail, calendar and tasks in the Facile apps, and answers in seconds.'
		},
		{
			icon: icons.server,
			title: 'It has its own computer',
			body: 'Code and scripts run on a private machine in the cloud, never on this one.'
		},
		{
			icon: icons.shield,
			title: 'You stay in control',
			body: 'Risky actions wait for your approval before they run.'
		}
	];

	let { error = $bindable('') }: { error?: string } = $props();
</script>

<div class="fixed inset-0 flex bg-fc-page text-fc-fg">
	<aside
		class="hidden w-[44%] max-w-xl flex-col bg-black px-12 pt-20 pb-10 text-white md:flex"
		data-tauri-drag-region
	>
		<div class="flex items-center gap-3">
			<Logo size={24} />
			<span class="font-fc-title text-fc-xl font-semibold tracking-tight">Bulle</span>
		</div>

		<div class="my-auto flex flex-col gap-10">
			<h2 class="font-fc-title text-fc-3xl leading-tight font-semibold tracking-tight">
				Your assistant.<br />On your desktop.
			</h2>
			<ul class="flex flex-col gap-6">
				{#each POINTS as point (point.title)}
					<li class="flex gap-4">
						<span
							class="flex size-10 shrink-0 items-center justify-center rounded-fc-md bg-white/10 text-white"
						>
							<Icon icon={point.icon} size={20} />
						</span>
						<div class="flex flex-col gap-1">
							<h3 class="text-fc-sm font-semibold">{point.title}</h3>
							<p class="text-fc-sm leading-relaxed text-white/55">{point.body}</p>
						</div>
					</li>
				{/each}
			</ul>
		</div>

		<p class="text-fc-xs text-white/30">© {new Date().getFullYear()} Bulle by Facile.</p>
	</aside>

	<main
		class="flex flex-1 flex-col items-center justify-center gap-8 overflow-y-auto px-8 py-12"
		data-tauri-drag-region
	>
		<OnboardingAvatar name="Pixel" size={136} />
		<div class="flex max-w-sm flex-col gap-3 text-center">
			<h1 class="font-fc-title text-fc-3xl leading-tight font-semibold tracking-tight">
				Sign in to Bulle
			</h1>
			<p class="text-fc-md leading-relaxed text-fc-fg-muted">
				Use your Facile account. Your sessions, jobs and approvals follow you from the web and your
				phone.
			</p>
		</div>
		<SignInButton bind:error />
	</main>
</div>
