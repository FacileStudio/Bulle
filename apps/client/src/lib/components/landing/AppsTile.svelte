<script lang="ts">
	import { Blob } from '@facile/blob';
	import { Icon, Spinner, icons } from '@facile/muse';
	import { Stage, Tile, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import { SUITE_APP_ICONS } from '$lib/suite-icons';

	const ROWS = [
		{ app: 'Courrier', icon: SUITE_APP_ICONS.courrier, key: 'landing.card.apps.mail' },
		{ app: 'Agenda', icon: SUITE_APP_ICONS.agenda, key: 'landing.card.apps.slot' },
		{ app: 'Opus', icon: SUITE_APP_ICONS.opus, key: 'landing.card.apps.task' },
		{ app: 'Nuage', icon: SUITE_APP_ICONS.nuage, key: 'landing.card.apps.file' }
	] as const;

	let active = $state(-1);
	let done = $state(0);

	function reset() {
		active = -1;
		done = 0;
	}

	function poster() {
		active = -1;
		done = ROWS.length;
	}

	async function script(act: Act) {
		await act.wait(500);
		for (let i = 0; i < ROWS.length; i++) {
			active = i;
			await act.wait(900);
			done = i + 1;
			await act.wait(250);
		}
		active = -1;
		await act.wait(2200);
		reset();
		await act.wait(400);
	}
</script>

<Tile title={t('landing.card.apps.title')} body={t('landing.card.apps.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="panel absolute inset-4 flex flex-col rounded-fc-lg bg-fc-page p-3">
				<div class="flex h-7 items-center gap-2 px-1">
					<Blob name="Bulle" size={20} />
					<span class="text-fc-xs font-bold text-fc-fg">Bulle</span>
				</div>
				<ul class="flex flex-col">
					{#each ROWS as row, i (row.app)}
						<li class="relative flex h-[38px] items-center gap-2.5 px-2 text-fc-xs">
							<span class="glow absolute inset-0 rounded-fc-md bg-fc-surface" class:on={active === i}></span>
							<Icon icon={row.icon} size={18} class="relative shrink-0 text-fc-accent" />
							<span class="relative shrink-0 font-semibold text-fc-fg">{row.app}</span>
							<span class="relative min-w-0 flex-1 truncate text-fc-fg-muted">{t(row.key)}</span>
							<span class="relative flex size-4 shrink-0 items-center justify-center">
								<span class="glow absolute" class:on={active === i && done <= i}><Spinner size="sm" /></span>
								<span class="glow absolute" class:on={done > i}>
									<Icon icon={icons.check} size={16} class="text-fc-success" />
								</span>
							</span>
						</li>
					{/each}
				</ul>
			</div>
		{/snippet}
	</Stage>
</Tile>

<style>
	.panel {
		box-shadow: 0 18px 40px -16px rgb(0 0 0 / 0.25);
	}
	.glow {
		transition: opacity 0.3s ease;
	}
	.glow:not(.on) {
		opacity: 0;
	}
</style>
