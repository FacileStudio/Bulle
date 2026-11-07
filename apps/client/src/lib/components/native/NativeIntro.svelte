<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Button, Icon, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import OnboardingAvatar from '$lib/components/onboarding/OnboardingAvatar.svelte';

	type Slide = { title: string; body: string; icon?: string };

	const SLIDES: Slide[] = [
		{
			get title() {
				return t('native.intro.one.title');
			},
			get body() {
				return t('native.intro.one.body');
			}
		},
		{
			get title() {
				return t('native.intro.two.title');
			},
			get body() {
				return t('native.intro.two.body');
			},
			icon: icons.monitor
		},
		{
			get title() {
				return t('native.intro.three.title');
			},
			get body() {
				return t('native.intro.three.body');
			},
			icon: icons.shield
		}
	];

	let { onfinish }: { onfinish: () => void } = $props();

	let track = $state<HTMLElement | null>(null);
	let index = $state(0);

	const last = $derived(index === SLIDES.length - 1);

	function onscroll() {
		if (!track) return;
		index = Math.round(track.scrollLeft / track.clientWidth);
	}

	function show(target: number) {
		track?.scrollTo({ left: target * track.clientWidth, behavior: 'smooth' });
	}

	function next() {
		if (last) onfinish();
		else show(index + 1);
	}
</script>

<div class="fixed inset-0 flex flex-col overflow-hidden bg-fc-page text-fc-fg">

	<header class="relative z-10 flex h-14 items-center justify-between px-5">
		<div class="flex items-center gap-2">
			<Logo size={22} />
			<span class="font-fc-title text-fc-md font-semibold tracking-tight">Bulle</span>
		</div>
		{#if !last}
			<Button variant="ghost" size="sm" class="px-3" onclick={onfinish}>{t('native.intro.skip')}</Button>
		{/if}
	</header>

	<div
		bind:this={track}
		{onscroll}
		class="relative z-10 flex min-h-0 flex-1 snap-x snap-mandatory overflow-x-auto overscroll-x-contain [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
		aria-roledescription="carousel"
	>
		{#each SLIDES as slide, i (slide.title)}
			<section
				class="flex w-full shrink-0 snap-center flex-col items-center justify-center gap-8 px-8 text-center"
				aria-roledescription="slide"
				aria-label={t('native.intro.slideOf', { index: i + 1, total: SLIDES.length })}
				aria-hidden={i !== index}
			>
				{#if slide.icon}
					<div
						class="flex size-36 items-center justify-center rounded-[2rem] bg-fc-surface shadow-sm backdrop-blur"
					>
						<Icon icon={slide.icon} size={56} />
					</div>
				{:else}
					<OnboardingAvatar name="Pixel" size={176} />
				{/if}
				<div class="flex max-w-xs flex-col gap-3">
					<h1 class="font-fc-title text-fc-3xl leading-tight font-semibold tracking-tight">
						{slide.title}
					</h1>
					<p class="text-fc-md leading-relaxed text-fc-fg-muted">{slide.body}</p>
				</div>
			</section>
		{/each}
	</div>

	<footer class="relative z-10 flex flex-col items-center gap-6 px-6 pt-4 pb-6">
		<div class="flex items-center gap-1" role="tablist" aria-label={t('native.intro.slides')}>
			{#each SLIDES as slide, i (slide.title)}
				<button
					type="button"
					role="tab"
					aria-selected={i === index}
					aria-label={t('native.intro.goTo', { index: i + 1 })}
					class="flex size-6 items-center justify-center"
					onclick={() => show(i)}
				>
					<span
						class="h-1.5 rounded-full transition-all duration-300 {i === index
							? 'w-6 bg-fc-fg'
							: 'w-1.5 bg-fc-fg/25'}"
					></span>
				</button>
			{/each}
		</div>
		<Button
			size="lg"
			class="w-full max-w-sm rounded-fc-md"
			iconRight={icons.arrow}
			onclick={next}
		>
			{last ? t('native.intro.start') : t('native.intro.next')}
		</Button>
	</footer>
</div>
