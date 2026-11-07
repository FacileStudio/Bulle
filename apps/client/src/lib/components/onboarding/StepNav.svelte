<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Button, Spinner, icons } from '@facile/muse';

	let {
		primary,
		busy = false,
		final = false,
		canBack,
		canSkip,
		oncontinue,
		onback,
		onskip
	}: {
		primary: string;
		busy?: boolean;
		final?: boolean;
		canBack: boolean;
		canSkip: boolean;
		oncontinue: () => void;
		onback: () => void;
		onskip: () => void;
	} = $props();
</script>

<footer
	class="relative z-10 bg-fc-page/70 px-4 pt-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] backdrop-blur sm:px-8"
>
	<div class="mx-auto flex w-full max-w-fc-sm flex-col gap-1">
		<Button
			size="lg"
			class="w-full rounded-fc-md"
			iconRight={final ? undefined : icons.arrow}
			disabled={busy}
			onclick={oncontinue}
		>
			{#if busy}
				<Spinner size="sm" />
			{/if}
			{primary}
		</Button>
		<div class="flex min-h-11 items-center justify-between">
			{#if canBack}
				<Button variant="ghost" size="lg" class="px-3" disabled={busy} onclick={onback}>{t('onboarding.nav.back')}</Button>
			{:else}
				<span></span>
			{/if}
			{#if canSkip}
				<Button variant="ghost" size="lg" class="px-3" onclick={onskip}>{t('onboarding.nav.skip')}</Button>
			{/if}
		</div>
	</div>
</footer>
