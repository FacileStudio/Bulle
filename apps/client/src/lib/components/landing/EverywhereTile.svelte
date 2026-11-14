<script lang="ts">
	import { Blob } from '@facile/blob';
	import { Icon, icons } from '@facile/muse';
	import { Stage, Tile, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import { SUITE_APP_ICONS } from '$lib/suite-icons';

	const PLATFORMS = [
		{ name: 'Web', icon: icons.globe },
		{ name: 'macOS', icon: icons.apple },
		{ name: 'Windows', icon: icons.windows },
		{ name: 'Linux', icon: icons.linux },
		{ name: 'Android', icon: icons.android }
	];

	let asked = $state(false);
	let answered = $state(false);
	let lit = $state(-1);

	function reset() {
		asked = false;
		answered = false;
		lit = -1;
	}

	function poster() {
		asked = true;
		answered = true;
		lit = -1;
	}

	async function script(act: Act) {
		await act.wait(500);
		asked = true;
		await act.wait(1300);
		answered = true;
		await act.wait(900);
		for (let i = 0; i < PLATFORMS.length; i++) {
			lit = i;
			await act.wait(450);
		}
		lit = -1;
		await act.wait(1800);
		reset();
		await act.wait(500);
	}
</script>

<Tile title={t('landing.card.everywhere.title')} body={t('landing.card.everywhere.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="panel absolute inset-x-4 top-4 flex h-[150px] flex-col gap-2.5 rounded-fc-lg bg-fc-page p-3">
				<div class="flex items-center gap-2 text-fc-xs text-fc-fg-muted">
					<Icon icon={SUITE_APP_ICONS.prisme} size={16} class="text-fc-accent" />
					<span class="font-bold text-fc-fg">Prisme</span>
					<span># projet-lune</span>
				</div>
				<div class="beat flex items-start gap-2" class:on={asked}>
					<span class="flex size-5 shrink-0 items-center justify-center rounded-full bg-fc-chart-2 text-[10px] font-bold text-white">L</span>
					<p class="min-w-0 text-fc-xs leading-5 text-fc-fg"><span class="font-bold">Léa</span> {t('landing.card.everywhere.ask')}</p>
				</div>
				<div class="beat flex items-start gap-2" class:on={answered}>
					<Blob name="Bulle" size={20} class="shrink-0" />
					<p class="min-w-0 text-fc-xs leading-5 text-fc-fg">{t('landing.card.everywhere.reply')}</p>
				</div>
			</div>
			<ul class="absolute inset-x-4 bottom-4 flex h-12 items-center justify-between gap-1">
				{#each PLATFORMS as platform, i (platform.name)}
					<li class="chip flex size-11 items-center justify-center rounded-fc-pill bg-fc-page text-fc-fg" class:lit={lit === i}>
						<Icon icon={platform.icon} size={20} />
						<span class="sr-only">{platform.name}</span>
					</li>
				{/each}
			</ul>
		{/snippet}
	</Stage>
</Tile>

<style>
	.panel {
		box-shadow: 0 18px 40px -16px rgb(0 0 0 / 0.25);
	}
	.beat {
		transition:
			opacity 0.4s ease,
			transform 0.4s ease;
	}
	.beat:not(.on) {
		opacity: 0;
		transform: translateY(6px);
	}
	.chip {
		transition: transform 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
	}
	.chip.lit {
		transform: translateY(-4px) scale(1.08);
	}
</style>
