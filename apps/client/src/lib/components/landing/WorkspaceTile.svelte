<script lang="ts">
	import { Blob } from '@facile/blob';
	import { Bar, Caret, Stage, Terminal, Tile, grow, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';

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

	function poster() {
		typed = prompt;
		sent = true;
		answered = true;
	}

	async function script(act: Act) {
		await act.wait(700);
		await act.type(prompt, (v) => (typed = v), 60);
		await act.wait(350);
		sent = true;
		await act.wait(1200);
		answered = true;
		await act.wait(2600);
		reset();
		await act.wait(500);
	}
</script>

<Tile title={t('landing.feature.workspace.title')} body={t('landing.feature.workspace.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Terminal style="top:20px;left:20px;right:48px;height:128px">
				<div class="dim mb-2">bulle</div>
				<div class="whitespace-nowrap">
					<span class="dim">›</span>
					{typed}{#if !sent}<Caret />{/if}
				</div>
				{#if answered}
					<div class="mt-1 text-white/70" in:grow={{ duration: 300 }}>
						<span class="ok">●</span>
						{reply}
					</div>
				{/if}
			</Terminal>

			<div class="web absolute top-[136px] -right-6 left-12 -bottom-6 flex flex-col gap-3 rounded-fc-lg bg-fc-page p-3.5 pr-10">
				{#if sent}
					<div class="flex justify-end" in:grow={{ duration: 350 }} out:grow={{ duration: 250, delay: 80 }}>
						<p class="rounded-fc-lg bg-fc-surface px-3 py-1.5 text-fc-xs text-fc-fg">{prompt}</p>
					</div>
				{:else}
					<div class="flex justify-end" out:grow={{ duration: 250 }}><Bar w="45%" /></div>
				{/if}
				{#if answered}
					<div class="flex items-start gap-2" in:grow={{ duration: 350 }} out:grow={{ duration: 250 }}>
						<Blob name="Bulle" size={20} class="shrink-0" />
						<p class="text-fc-xs leading-5 text-fc-fg">{reply}</p>
					</div>
				{:else}
					<div out:grow={{ duration: 250 }}><Bar w="70%" /></div>
					<Bar w="52%" />
				{/if}
			</div>
		{/snippet}
	</Stage>
</Tile>

<style>
	.dim {
		color: rgb(255 255 255 / 0.45);
	}
	.ok {
		color: color-mix(in oklch, var(--color-fc-success), white 35%);
	}
	.web {
		box-shadow: 0 18px 40px -12px rgb(0 0 0 / 0.28);
	}
</style>
