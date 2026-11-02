<script lang="ts">
	import { Button, Icon, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import Ambient from '$lib/components/onboarding/Ambient.svelte';
	import OnboardingAvatar from '$lib/components/onboarding/OnboardingAvatar.svelte';

	type Slide = { title: string; body: string; icon?: string; tint: readonly [string, string] };

	const SLIDES: Slide[] = [
		{
			title: 'Your assistant, in your pocket.',
			body: 'It reads your mail, calendar and tasks across the Facile suite, and answers in seconds.',
			tint: ['var(--color-fc-chart-3)', 'var(--color-fc-chart-1)']
		},
		{
			title: 'It has its own computer.',
			body: 'Ask for code or a quick script. It runs and builds on a private machine, never on your phone.',
			icon: icons.monitor,
			tint: ['var(--color-fc-chart-1)', 'var(--color-fc-chart-6)']
		},
		{
			title: 'You stay in control.',
			body: 'Risky actions wait for your approval. A notification brings you straight to them.',
			icon: icons.shield,
			tint: ['var(--color-fc-chart-2)', 'var(--color-fc-chart-5)']
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
	<Ambient tint={SLIDES[index].tint} />

	<header class="relative z-10 flex h-14 items-center justify-between px-5">
		<div class="flex items-center gap-2">
			<Logo size={22} />
			<span class="font-fc-title text-fc-md font-semibold tracking-tight">Bulle</span>
		</div>
		{#if !last}
			<Button variant="ghost" size="sm" class="px-3" onclick={onfinish}>Skip</Button>
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
				aria-label="{i + 1} of {SLIDES.length}"
				aria-hidden={i !== index}
			>
				{#if slide.icon}
					<div
						class="flex size-36 items-center justify-center rounded-[2rem] border border-fc-border bg-fc-surface/70 shadow-sm backdrop-blur"
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
		<div class="flex items-center gap-1" role="tablist" aria-label="Slides">
			{#each SLIDES as slide, i (slide.title)}
				<button
					type="button"
					role="tab"
					aria-selected={i === index}
					aria-label="Go to slide {i + 1}"
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
			{last ? 'Get started' : 'Next'}
		</Button>
	</footer>
</div>
