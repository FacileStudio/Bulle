<script lang="ts">
	import { tick } from 'svelte';
	import { Button, Icon, icons } from '@facile/muse';
	import { Cursor, Stage, Tile, cursorAt, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import { SUITE_APP_ICONS } from '$lib/suite-icons';

	type Phase = 'pending' | 'pressed' | 'approved';

	let phase = $state<Phase>('pending');
	let cursor = $state(cursorAt(250, 210));
	let approveEl = $state<HTMLElement>();

	function reset() {
		phase = 'pending';
		cursor.visible = false;
		cursor.down = false;
	}

	function poster() {
		phase = 'approved';
	}

	async function script(act: Act) {
		await tick();
		const rest = act.point(approveEl, 2.6, 2.2);
		await act.move(cursor, rest, 0);
		await act.wait(600);
		cursor.visible = true;
		await act.wait(400);
		await act.move(cursor, act.point(approveEl, 0.55, 0.6), 1000);
		await act.click(cursor, () => (phase = 'pressed'));
		phase = 'approved';
		await act.move(cursor, rest, 700);
		cursor.visible = false;
		await act.wait(2200);
		reset();
		await act.wait(500);
	}
</script>

<Tile title={t('landing.card.approve.title')} body={t('landing.card.approve.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="panel absolute inset-x-4 top-12 flex flex-col gap-3 rounded-fc-lg bg-fc-page p-4">
				<div class="flex items-center gap-2.5">
					<span class="flex size-9 shrink-0 items-center justify-center rounded-fc-md bg-fc-surface">
						<Icon icon={SUITE_APP_ICONS.agenda} size={20} class="text-fc-accent" />
					</span>
					<div class="flex min-w-0 flex-col">
						<span class="truncate text-fc-sm font-bold text-fc-fg">{t('landing.card.approve.action')}</span>
						<span class="truncate text-fc-xs text-fc-fg-muted">{t('landing.card.approve.detail')}</span>
					</div>
				</div>
				<div class="grid h-10 items-center">
					<div class="swap col-start-1 row-start-1 flex items-center gap-2" class:on={phase !== 'approved'}>
						<span bind:this={approveEl} class="press inline-flex" class:down={phase === 'pressed'}>
							<Button size="sm" tabindex={-1}>{t('approval.approve')}</Button>
						</span>
						<Button size="sm" variant="ghost" tabindex={-1}>{t('approval.deny')}</Button>
					</div>
					<p class="swap col-start-1 row-start-1 flex items-center gap-2 text-fc-sm text-fc-success" class:on={phase === 'approved'}>
						<Icon icon={icons.check} size={18} />
						{t('approval.approved')}
					</p>
				</div>
			</div>
			<Cursor {cursor} />
		{/snippet}
	</Stage>
</Tile>

<style>
	.panel {
		box-shadow: 0 18px 40px -16px rgb(0 0 0 / 0.25);
	}
	.swap {
		transition:
			opacity 0.3s ease,
			transform 0.3s ease;
	}
	.swap:not(.on) {
		opacity: 0;
		transform: translateY(6px);
		pointer-events: none;
	}
	.press {
		transition: transform 0.1s ease;
	}
	.press.down {
		transform: scale(0.95);
	}
</style>
