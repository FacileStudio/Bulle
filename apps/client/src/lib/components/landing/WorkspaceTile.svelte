<script lang="ts">
	import { Blob } from '@facile/blob';
	import { t } from '$lib/i18n/index.svelte';
	import DemoTile from './DemoTile.svelte';
	import { grow, playInView, prefersReducedMotion, sleep, typeOut } from './stage';

	const prompt = $derived(t('landing.demo.ws.prompt'));
	const reply = $derived(t('landing.demo.ws.reply'));

	let typed = $state('');
	let sent = $state(false);
	let answered = $state(false);

	function reset() {
		typed = '';
		sent = false;
		answered = false;
	}

	async function script(signal: AbortSignal) {
		await sleep(700, signal);
		await typeOut(prompt, (v) => (typed = v), 60, signal);
		await sleep(350, signal);
		sent = true;
		await sleep(1200, signal);
		answered = true;
		await sleep(2600, signal);
		reset();
		await sleep(500, signal);
	}

	if (prefersReducedMotion()) {
		typed = t('landing.demo.ws.prompt');
		sent = true;
		answered = true;
	}
</script>

<DemoTile title={t('landing.feature.workspace.title')} body={t('landing.feature.workspace.body')}>
	<div {@attach playInView(script, reset)}>
		<div class="term absolute top-5 left-5 right-12 h-[128px] rounded-fc-lg px-3.5 py-3 font-fc-mono text-fc-xs leading-5">
			<div class="mb-2 text-white/45">bulle</div>
			<div class="whitespace-nowrap">
				<span class="text-white/45">›</span>
				<span class="text-white">{typed}</span>{#if !sent}<span class="caret"></span>{/if}
			</div>
			{#if answered}
				<div class="mt-1 text-white/70" in:grow={{ duration: 300 }}>
					<span class="text-[oklch(0.8_0.15_150)]">●</span>
					{reply}
				</div>
			{/if}
		</div>

		<div class="web absolute top-[136px] -right-6 left-12 -bottom-6 flex flex-col gap-3 rounded-fc-lg bg-fc-page p-3.5 pr-10">
			{#if sent}
				<div class="flex justify-end" in:grow={{ duration: 350 }} out:grow={{ duration: 250, delay: 80 }}>
					<p class="rounded-fc-lg bg-fc-surface px-3 py-1.5 text-fc-xs text-fc-fg">{prompt}</p>
				</div>
			{:else}
				<span class="bar ml-auto w-[45%]" out:grow={{ duration: 250 }}></span>
			{/if}
			{#if answered}
				<div class="flex items-start gap-2" in:grow={{ duration: 350 }} out:grow={{ duration: 250 }}>
					<Blob name="Bulle" size={20} class="shrink-0" />
					<p class="text-fc-xs leading-5 text-fc-fg">{reply}</p>
				</div>
			{:else}
				<span class="bar w-[70%]" out:grow={{ duration: 250 }}></span>
				<span class="bar w-[52%]"></span>
			{/if}
		</div>
	</div>
</DemoTile>

<style>
	.term {
		background: oklch(0.18 0 0);
		box-shadow: 0 18px 40px -16px rgb(0 0 0 / 0.35);
	}
	.web {
		box-shadow: 0 18px 40px -12px rgb(0 0 0 / 0.28);
	}
	.bar {
		display: block;
		height: 8px;
		border-radius: 4px;
		background: var(--color-fc-surface);
	}
	.caret {
		display: inline-block;
		width: 1ch;
		height: 1.1em;
		margin-left: 1px;
		vertical-align: -0.2em;
		background: rgb(255 255 255 / 0.75);
		animation: blink 1.1s steps(1) infinite;
	}
	@keyframes blink {
		50% {
			opacity: 0;
		}
	}
</style>
