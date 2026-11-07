<script lang="ts">
	import { fly } from 'svelte/transition';
	import { Icon, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';
	import DemoTile from './DemoTile.svelte';
	import { playInView, prefersReducedMotion, sleep } from './stage';

	const MODELS = ['Claude Opus', 'Gemini Pro', 'Kimi K2'];
	const BARS = [42, 58, 50, 74, 66, 88, 80];
	const DAYS = ['M', 'T', 'W', 'T', 'F', 'S', 'S'];

	let model = $state(0);
	let drawn = $state(true);

	function reset() {
		model = 0;
		drawn = true;
	}

	async function script(signal: AbortSignal) {
		await sleep(2600, signal);
		model = (model + 1) % MODELS.length;
		await sleep(250, signal);
		drawn = false;
		await sleep(350, signal);
		drawn = true;
	}

	if (prefersReducedMotion()) reset();
</script>

<DemoTile title={t('landing.feature.models.title')} body={t('landing.feature.models.body')}>
	<div
		class="panel absolute top-6 left-6 -right-10 -bottom-6 flex flex-col gap-4 rounded-fc-lg bg-fc-page p-4"
		{@attach playInView(script, reset)}
	>
		<div class="flex items-center gap-2 self-start rounded-fc-pill bg-fc-surface py-1.5 pr-2.5 pl-3 text-fc-xs font-medium text-fc-fg">
			<span class="relative grid h-4 w-24 overflow-hidden">
				{#key model}
					<span
						class="col-start-1 row-start-1 leading-4"
						in:fly={{ y: 14, duration: 380, opacity: 0 }}
						out:fly={{ y: -14, duration: 380, opacity: 0 }}
					>
						{MODELS[model]}
					</span>
				{/key}
			</span>
			<Icon icon={icons.chevronDown} size={14} class="text-fc-fg-muted" />
		</div>
		<div class="flex flex-1 flex-col gap-3 rounded-fc-md bg-fc-component p-3 pr-12">
			<span class="text-fc-xs font-medium text-fc-fg">{t('landing.demo.chart')}</span>
			<div class="flex flex-1 items-end gap-2">
				{#each BARS as value, i (i)}
					<div class="flex flex-1 flex-col items-center gap-1.5">
						<div class="flex h-20 w-full items-end">
							<div
								class="bar w-full rounded-fc-xs bg-fc-chart-1"
								style:height="{drawn ? value : 0}%"
								style:transition-delay="{drawn ? i * 60 : 0}ms"
							></div>
						</div>
						<span class="text-[10px] leading-3 text-fc-fg-muted">{DAYS[i]}</span>
					</div>
				{/each}
			</div>
		</div>
	</div>
</DemoTile>

<style>
	.panel {
		box-shadow: 0 18px 40px -16px rgb(0 0 0 / 0.25);
	}
	.bar {
		transition: height 0.55s cubic-bezier(0.25, 0.8, 0.25, 1);
	}
</style>
