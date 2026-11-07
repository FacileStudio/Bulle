<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button, Card, DownloadCta, Icon, PricingPlans, SignInWithFacile, Spinner, TextElevate, icons } from '@facile/muse';
	import { Blob, gaze } from '@facile/blob';
	import NativeIntro from '$lib/components/native/NativeIntro.svelte';
	import InstallCommands from '$lib/components/InstallCommands.svelte';
	import PublicShell from '$lib/components/PublicShell.svelte';
	import { ssoLoginUrl } from '$lib/backend';
	import { APK_URL } from '$lib/downloads';
	import { CHAT_ICONS } from '$lib/icons';
	import { isDesktop, isNative } from '$lib/platform';
	import { loadSession, session } from '$lib/session.svelte';

	interface Feature {
		icon: string;
		title: string;
		description: string;
		blob: string;
	}

	const features: Feature[] = $derived([
		{
			icon: icons.server,
			title: t('landing.feature.vm.title'),
			description: t('landing.feature.vm.body'),
			blob: 'Boite'
		},
		{
			icon: icons.bolt,
			title: t('landing.feature.models.title'),
			description: t('landing.feature.models.body'),
			blob: 'Nova'
		},
		{
			icon: icons.monitor,
			title: t('landing.feature.workspace.title'),
			description: t('landing.feature.workspace.body'),
			blob: 'Pixel'
		}
	]);

	const eyesHeroCenter = gaze({ travel: 6, target: 'pointer' });
	const eyesHeroLeft = gaze({ travel: 5, target: 'pointer' });
	const eyesHeroRight = gaze({ travel: 5, target: 'pointer' });
	const eyesCta = gaze({ travel: 6, target: 'pointer' });

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
		if (isDesktop() || (native && introSeen())) {
			await goto('/login', { replaceState: true });
			return;
		}
		ready = true;
	});
</script>

<svelte:head>
	<title>{t('landing.title')}</title>
	<meta
		name="description"
		content={t('landing.description')}
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
		<div class="relative overflow-hidden">
			<div
				aria-hidden="true"
				class="pointer-events-none absolute -top-32 left-1/2 -translate-x-1/2 h-80 w-[560px] rounded-full bg-fc-accent/10 blur-[110px]"
			></div>
			<div
				aria-hidden="true"
				class="pointer-events-none absolute top-1/3 -left-32 size-72 rounded-full bg-fc-accent/5 blur-[90px]"
			></div>
			<div
				aria-hidden="true"
				class="pointer-events-none absolute top-2/3 -right-32 size-72 rounded-full bg-fc-accent/5 blur-[90px]"
			></div>

			<section class="relative mx-auto max-w-5xl px-6 py-20 text-center sm:py-24">
				<div class="relative mx-auto mb-8 flex items-center justify-center -space-x-3 sm:-space-x-4">
					<div class="relative z-10 -rotate-6 transition-transform duration-300 hover:z-30 hover:scale-110 cursor-pointer">
						<Blob
							{@attach eyesHeroLeft}
							name="Spark"
							size={56}
							animate="always"
							class="drop-shadow-md"
						/>
					</div>
					<div class="relative z-20 transition-transform duration-300 hover:z-30 hover:scale-110 cursor-pointer">
						<Blob
							{@attach eyesHeroCenter}
							name="Bulle"
							size={96}
							animate="always"
							class="drop-shadow-lg"
						/>
					</div>
					<div class="relative z-10 rotate-6 transition-transform duration-300 hover:z-30 hover:scale-110 cursor-pointer">
						<Blob
							{@attach eyesHeroRight}
							name="Pixel"
							size={56}
							animate="always"
							class="drop-shadow-md"
						/>
					</div>
				</div>

				<h1
					class="font-fc-title text-4xl leading-[1.1] font-semibold tracking-tight text-fc-fg sm:text-5xl lg:text-6xl"
				>
					<span class="block"><TextElevate text={t('landing.hero.line1')} delay={0.1} /></span>
					<span class="block"><TextElevate text={t('landing.hero.line2')} delay={0.25} /></span>
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted sm:text-fc-lg">
					{t('landing.hero.body')}
				</p>
				<div class="mt-10 flex flex-col items-stretch justify-center gap-3 sm:flex-row">
					<SignInWithFacile href={ssoLoginUrl} />
					<DownloadCta
						href="/downloads"
						platforms={['macos', 'windows', 'linux', 'android']}
						class="justify-center self-auto"
					/>
				</div>
			</section>

			<section class="relative mx-auto max-w-5xl px-6 py-16 sm:py-20">
				<div class="grid gap-6 md:grid-cols-3">
					{#each features as feature (feature.title)}
						<Card class="flex flex-col gap-4">
							<div class="flex items-center justify-between">
								<div
									class="flex size-10 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg"
								>
									<Icon icon={feature.icon} size={20} />
								</div>
								<div class="transition-transform duration-200 hover:scale-110">
									<Blob name={feature.blob} size={36} animate="hover" />
								</div>
							</div>
							<div class="flex flex-col gap-1.5">
								<h3 class="font-fc-title text-fc-lg font-semibold text-fc-fg">{feature.title}</h3>
								<p class="text-fc-sm text-fc-fg-muted">{feature.description}</p>
							</div>
						</Card>
					{/each}
				</div>
			</section>

			<PricingPlans
				highlight={['ai.access']}
				title={t('landing.pricing')}
				class="relative mx-auto max-w-5xl px-6 py-16 sm:py-20"
			/>

			<section class="relative mx-auto max-w-5xl px-6 py-16 sm:py-20">
				<div class="text-center">
					<h2 class="font-fc-title text-2xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
						{t('landing.install.title')}
					</h2>
					<p class="mx-auto mt-3 max-w-xl text-fc-sm text-fc-fg-muted">
						{t('landing.install.body')}
					</p>
				</div>
				<div class="mx-auto mt-8 max-w-2xl">
					<Card>
						<InstallCommands />
					</Card>
				</div>
			</section>

			<section class="relative mx-auto max-w-5xl px-6 py-16 sm:py-20">
				<div class="text-center">
					<h2 class="font-fc-title text-2xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
						{t('landing.mobile.title')}
					</h2>
					<p class="mx-auto mt-3 max-w-xl text-fc-sm text-fc-fg-muted">
						{t('landing.mobile.body')}
					</p>
				</div>
				<div class="mx-auto mt-8 grid max-w-2xl gap-4 sm:grid-cols-2">
					<Card class="flex flex-col gap-3">
						<div class="flex items-center gap-2 text-fc-fg">
							<Icon icon={CHAT_ICONS.phone} size={18} />
							<h3 class="font-fc-title text-fc-md font-semibold">Android</h3>
						</div>
						<p class="text-fc-sm text-fc-fg-muted">{t('landing.android.body')}</p>
						<div class="mt-auto">
							<Button href={APK_URL} size="sm" icon={icons.download} class="rounded-fc-md">
								{t('landing.android.cta')}
							</Button>
						</div>
					</Card>
					<Card class="flex flex-col gap-3">
						<div class="flex flex-wrap items-center gap-2 text-fc-fg">
							<Icon icon={icons.monitor} size={18} />
							<h3 class="font-fc-title text-fc-md font-semibold">{t('landing.desktop.title')}</h3>
						</div>
						<p class="text-fc-sm text-fc-fg-muted">{t('landing.desktop.body')}</p>
						<div class="mt-auto">
							<Button href="/downloads" size="sm" icon={icons.download} class="rounded-fc-md">
								{t('landing.desktop.cta')}
							</Button>
						</div>
					</Card>
				</div>
			</section>

			<section class="relative mx-auto max-w-5xl px-6 py-16 text-center sm:py-20">
				<div class="mx-auto mb-5 flex justify-center">
					<div class="transition-transform duration-300 hover:scale-110 cursor-pointer">
						<Blob
							{@attach eyesCta}
							name="Bulle"
							size={76}
							animate="always"
							class="drop-shadow-md"
						/>
					</div>
				</div>
				<h2 class="font-fc-title text-2xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
					{t('landing.cta.title')}
				</h2>
				<p class="mx-auto mt-4 max-w-xl text-fc-sm text-fc-fg-muted sm:text-fc-base">
					{t('landing.cta.body')}
				</p>
				<div class="mt-8 flex justify-center">
					<SignInWithFacile href={ssoLoginUrl} />
				</div>
			</section>
		</div>
	</PublicShell>
{/if}
