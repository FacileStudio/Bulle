<script lang="ts">
	import { locale, t } from '$lib/i18n/index.svelte';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		ForDevelopers,
		LANDING_HEADING,
		LandingCta,
		PricingPlans,
		Spinner,
		TextElevate,
		WorksWith,
		icons,
		signUpUrl,
		type DeveloperRow,
		type WorksWithApp
	} from '@facile/muse';
	import { Blob, gaze } from '@facile/blob';
	import NativeIntro from '$lib/components/native/NativeIntro.svelte';
	import Downloads from '$lib/components/Downloads.svelte';
	import PublicShell from '$lib/components/PublicShell.svelte';
	import HeroDemo from '$lib/components/landing/HeroDemo.svelte';
	import AppsTile from '$lib/components/landing/AppsTile.svelte';
	import ApproveTile from '$lib/components/landing/ApproveTile.svelte';
	import EverywhereTile from '$lib/components/landing/EverywhereTile.svelte';
	import { SUITE_APP_ICONS } from '$lib/suite-icons';
	import { ssoLoginUrl } from '$lib/backend';
	import { isDesktop, isNative } from '$lib/platform';
	import { loadSession, session } from '$lib/session.svelte';

	const eyesHeroCenter = gaze({ travel: 6, target: 'pointer' });
	const eyesHeroLeft = gaze({ travel: 5, target: 'pointer' });
	const eyesHeroRight = gaze({ travel: 5, target: 'pointer' });
	const eyesCta = gaze({ travel: 6, target: 'pointer' });

	const WIDTH = 'mx-auto w-full max-w-5xl px-4 sm:px-6';

	const apps = $derived<WorksWithApp[]>([
		{ name: 'Courrier', icon: SUITE_APP_ICONS.courrier, line: t('landing.with.courrier'), href: 'https://courrier.facile.studio' },
		{ name: 'Agenda', icon: SUITE_APP_ICONS.agenda, line: t('landing.with.agenda'), href: 'https://agenda.facile.studio' },
		{ name: 'Opus', icon: SUITE_APP_ICONS.opus, line: t('landing.with.opus'), href: 'https://opus.facile.studio' },
		{ name: 'Nuage', icon: SUITE_APP_ICONS.nuage, line: t('landing.with.nuage'), href: 'https://nuage.facile.studio' },
		{ name: 'Prisme', icon: SUITE_APP_ICONS.prisme, line: t('landing.with.prisme'), href: 'https://prisme.facile.studio' }
	]);

	const devRows = $derived<DeveloperRow[]>([
		{ icon: icons.server, title: t('landing.dev.vm.title'), text: t('landing.dev.vm.text') },
		{ kind: 'cli', text: t('landing.dev.cli'), command: 'brew install FacileStudio/tap/bulle' },
		{ icon: icons.layers, title: t('landing.dev.models.title'), text: t('landing.dev.models.text') },
		{ kind: 'source', text: t('landing.dev.source'), href: 'https://github.com/FacileStudio/Bulle' }
	]);

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
	<meta name="description" content={t('landing.description')} />
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
			<section class="relative {WIDTH} py-20 text-center sm:py-24">
				<div class="relative mx-auto mb-8 flex items-center justify-center -space-x-3 sm:-space-x-4">
					<div class="relative z-10 -rotate-6 cursor-pointer transition-transform duration-300 hover:z-30 hover:scale-110">
						<Blob {@attach eyesHeroLeft} name="Spark" size={56} animate="always" class="drop-shadow-md" />
					</div>
					<div class="relative z-20 cursor-pointer transition-transform duration-300 hover:z-30 hover:scale-110">
						<Blob {@attach eyesHeroCenter} name="Bulle" size={96} animate="always" class="drop-shadow-lg" />
					</div>
					<div class="relative z-10 rotate-6 cursor-pointer transition-transform duration-300 hover:z-30 hover:scale-110">
						<Blob {@attach eyesHeroRight} name="Pixel" size={56} animate="always" class="drop-shadow-md" />
					</div>
				</div>

				<h1 class="mx-auto max-w-3xl font-fc-title text-4xl leading-[1.1] font-bold tracking-tight text-fc-fg sm:text-5xl lg:text-6xl">
					<TextElevate text={t('landing.hero.title')} delay={0.1} />
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted sm:text-fc-lg">
					{t('landing.hero.body')}
				</p>
				<LandingCta signInHref={ssoLoginUrl} align="center" class="mt-10" />
				<p class="mx-auto mt-4 max-w-xl text-fc-sm text-fc-fg-muted">
					<a href="#pricing" class="underline-offset-4 hover:text-fc-fg hover:underline">{t('landing.hero.plan')}</a>
				</p>
				<div class="mt-14 sm:mt-16">
					<HeroDemo />
				</div>
			</section>

			<section id="features" class="relative {WIDTH} scroll-mt-20 py-16 sm:py-20">
				<h2 class="mb-10 text-center {LANDING_HEADING}">{t('landing.features.title')}</h2>
				<div class="grid gap-6 md:grid-cols-3">
					<AppsTile />
					<ApproveTile />
					<EverywhereTile />
				</div>
			</section>

			<WorksWith class="relative {WIDTH} py-16 sm:py-20" {apps} description={t('landing.with.description')} />

			<PricingPlans
				id="pricing"
				class="relative {WIDTH} scroll-mt-20 py-16 sm:py-20"
				title={t('landing.pricing')}
				headingClass={LANDING_HEADING}
				highlight={['ai.access']}
				highlightOnly
				notes={{ solo: t('landing.plan.solo'), '*': t('landing.plan.paid') }}
				recommended="pro"
				freeLabel={t('landing.plan.freeCta')}
				freeHref={signUpUrl(locale())}
				suiteHref="https://suite.facile.studio/#tarifs"
				suiteNote={t('landing.plan.note')}
			/>

			<div class="relative {WIDTH} py-16 sm:py-20">
				<ForDevelopers rows={devRows} selfHostHref={t('landing.dev.selfHostHref')} description={t('landing.dev.description')} headingClass={LANDING_HEADING} />
			</div>

			<section id="download" class="relative mx-auto w-full max-w-3xl scroll-mt-20 px-4 py-16 sm:px-6 sm:py-20">
				<div class="mb-8 text-center">
					<h2 class={LANDING_HEADING}>{t('landing.download.title')}</h2>
					<p class="mx-auto mt-3 max-w-xl text-fc-sm text-fc-fg-muted">{t('landing.download.body')}</p>
				</div>
				<Downloads headingLevel={3} mode="device" />
			</section>

			<section class="relative {WIDTH} py-16 text-center sm:py-20">
				<div class="mx-auto mb-5 flex justify-center">
					<div class="cursor-pointer transition-transform duration-300 hover:scale-110">
						<Blob {@attach eyesCta} name="Bulle" size={76} animate="always" class="drop-shadow-md" />
					</div>
				</div>
				<h2 class={LANDING_HEADING}>{t('landing.cta.title')}</h2>
				<p class="mx-auto mt-4 max-w-xl text-fc-sm text-fc-fg-muted sm:text-fc-base">{t('landing.cta.body')}</p>
				<LandingCta signInHref={ssoLoginUrl} align="center" class="mt-8" />
			</section>
		</div>
	</PublicShell>
{/if}
