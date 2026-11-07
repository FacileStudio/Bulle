<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Icon, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import OnboardingAvatar from '$lib/components/onboarding/OnboardingAvatar.svelte';
	import SignInButton from './SignInButton.svelte';

	type Point = { icon: string; title: string; body: string };

	const POINTS: Point[] = $derived([
		{ icon: icons.mail, title: t('login.point.suite.title'), body: t('login.point.suite.body') },
		{ icon: icons.server, title: t('login.point.computer.title'), body: t('login.point.computer.body') },
		{ icon: icons.shield, title: t('login.point.control.title'), body: t('login.point.control.body') }
	]);

	let { error = $bindable('') }: { error?: string } = $props();
</script>

<div class="fixed inset-0 flex bg-fc-page text-fc-fg">
	<aside
		class="hidden w-[44%] max-w-xl flex-col bg-fc-accent px-12 pt-20 pb-10 text-fc-accent-fg md:flex"
		data-tauri-drag-region
	>
		<div class="flex items-center gap-3">
			<Logo size={24} />
			<span class="font-fc-title text-fc-xl font-semibold tracking-tight">Bulle</span>
		</div>

		<div class="my-auto flex flex-col gap-10">
			<h2 class="font-fc-title text-fc-3xl leading-tight font-semibold tracking-tight">
				{t('login.desktop.title')}
			</h2>
			<ul class="flex flex-col gap-6">
				{#each POINTS as point (point.title)}
					<li class="flex gap-4">
						<span
							class="flex size-10 shrink-0 items-center justify-center rounded-fc-md bg-fc-accent-fg/10 text-fc-accent-fg"
						>
							<Icon icon={point.icon} size={20} />
						</span>
						<div class="flex flex-col gap-1">
							<h3 class="text-fc-sm font-semibold">{point.title}</h3>
							<p class="text-fc-sm leading-relaxed text-fc-accent-fg/60">{point.body}</p>
						</div>
					</li>
				{/each}
			</ul>
		</div>

		<p class="text-fc-xs text-fc-accent-fg/40">{t('login.copyright', { year: new Date().getFullYear() })}</p>
	</aside>

	<main
		class="flex flex-1 flex-col items-center justify-center gap-8 overflow-y-auto px-8 py-12"
		data-tauri-drag-region
	>
		<OnboardingAvatar name="Pixel" size={136} />
		<div class="flex max-w-sm flex-col gap-3 text-center">
			<h1 class="font-fc-title text-fc-3xl leading-tight font-semibold tracking-tight">
				{t('login.native.title')}
			</h1>
			<p class="text-fc-md leading-relaxed text-fc-fg-muted">
				{t('login.desktop.body')}
			</p>
		</div>
		<SignInButton bind:error />
	</main>
</div>
