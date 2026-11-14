<script module lang="ts">
	export type StepState = 'hidden' | 'running' | 'ok';
	export type ApprovalState = 'hidden' | 'pending' | 'pressed' | 'approved';

	export interface HeroScene {
		draft: string;
		sent: boolean;
		assistant: boolean;
		steps: StepState[];
		approval: ApprovalState;
		words: number;
	}
</script>

<script lang="ts">
	import { Blob } from '@facile/blob';
	import { Button, Icon, Spinner, icons } from '@facile/muse';
	import { Stream } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import { SUITE_APP_ICONS } from '$lib/suite-icons';
	import { kindLabel } from '$lib/chat/tools';

	let {
		scene,
		compact,
		approveEl = $bindable()
	}: { scene: HeroScene; compact: boolean; approveEl?: HTMLElement } = $props();

	const STEPS = [
		{ app: 'Courrier', icon: SUITE_APP_ICONS.courrier, key: 'landing.demo.mail' },
		{ app: 'Agenda', icon: SUITE_APP_ICONS.agenda, key: 'landing.demo.slot' },
		{ app: 'Opus', icon: SUITE_APP_ICONS.opus, key: 'landing.demo.tasks' }
	] as const;

	const prompt = $derived(t('landing.demo.prompt'));
	const reply = $derived(t('landing.demo.reply'));
	const asking = $derived(scene.approval === 'pending' || scene.approval === 'pressed');
	const stepsShown = $derived(scene.steps.some((s) => s !== 'hidden'));
	const anyRunning = $derived(scene.steps.includes('running'));
	const streaming = $derived(scene.approval === 'approved' && scene.words < reply.split(' ').length);
</script>

<div class="flex min-h-0 flex-1 flex-col pb-5" class:px-6={!compact} class:px-4={compact}>
	<div class="relative flex min-h-0 flex-1 flex-col pt-5">
		<div class="fade late absolute inset-0 flex flex-col items-center justify-center gap-3" class:on={!scene.sent}>
			<Blob name="Bulle" size={64} />
			<p class="text-fc-lg font-bold text-fc-fg">{t('landing.demo.empty')}</p>
		</div>
		<div class="beat flex justify-end" class:on={scene.sent}>
			<p class="max-w-[85%] rounded-fc-lg bg-fc-surface px-4 py-2 text-fc-md text-fc-fg">{prompt}</p>
		</div>

		<div class="beat flex items-center gap-2 pt-4" class:on={scene.assistant}>
			<Blob name="Bulle" size={24} />
			<span class="text-fc-xs font-bold text-fc-fg">Bulle</span>
		</div>

		<div class="beat pt-3" class:on={stepsShown}>
			<section class="card overflow-hidden rounded-fc-lg">
				<header class="flex items-center gap-2 px-3 py-2 text-fc-sm">
					<Icon icon={CHAT_ICONS.tool} size={16} class="text-fc-fg-muted" />
					<span class="text-fc-fg">{kindLabel('apps')}</span>
					<span class="ml-auto flex size-4 items-center justify-center">
						<span class="fade" class:on={anyRunning}><Spinner size="sm" /></span>
					</span>
				</header>
				<ul class="flex flex-col pb-1">
					{#each STEPS as step, i (step.app)}
						<li class="beat flex h-10 items-center gap-3 px-3 text-fc-sm" class:on={scene.steps[i] !== 'hidden'}>
							<Icon icon={step.icon} size={18} class="shrink-0 text-fc-accent" />
							<span class="shrink-0 font-semibold text-fc-fg">{step.app}</span>
							<span class="min-w-0 flex-1 truncate text-fc-xs text-fc-fg-muted">{t(step.key)}</span>
							<span class="relative flex size-4 shrink-0 items-center justify-center">
								<span class="fade absolute" class:on={scene.steps[i] === 'running'}><Spinner size="sm" /></span>
								<span class="fade absolute" class:on={scene.steps[i] === 'ok'}>
									<Icon icon={icons.check} size={16} class="text-fc-success" />
								</span>
							</span>
						</li>
					{/each}
				</ul>
			</section>
		</div>

		<div class="grid pt-4">
			<div class="beat col-start-1 row-start-1 flex flex-col gap-3" class:on={asking}>
				<div class="flex flex-col gap-0.5">
					<span class="text-fc-md font-bold text-fc-fg">{t('approval.tool.courrier_send')}</span>
					<span class="text-fc-xs text-fc-fg-muted">{t('approval.needed')}</span>
				</div>
				<dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-4 gap-y-1.5 text-fc-sm">
					<dt class="text-fc-fg-muted">{t('landing.demo.to')}</dt>
					<dd class="truncate text-fc-fg">claire@atelier-lune.fr</dd>
					<dt class="text-fc-fg-muted">{t('landing.demo.subjectLabel')}</dt>
					<dd class="truncate text-fc-fg">{t('landing.demo.subject')}</dd>
				</dl>
				<div class="flex items-center gap-2">
					<span bind:this={approveEl} class="press inline-flex" class:down={scene.approval === 'pressed'}>
						<Button tabindex={-1}>{t('approval.approve')}</Button>
					</span>
					<Button variant="ghost" tabindex={-1}>{t('approval.deny')}</Button>
				</div>
			</div>
			<div class="beat col-start-1 row-start-1 flex flex-col gap-2" class:on={scene.approval === 'approved'}>
				<p class="flex items-center gap-2 text-fc-sm text-fc-fg-muted">
					<Icon icon={SUITE_APP_ICONS.courrier} size={16} class="text-fc-accent" />
					<span class="truncate">{t('approval.tool.courrier_send')}</span>
					<span class="text-fc-xs text-fc-success">{t('approval.approved')}</span>
				</p>
				<Stream text={reply} count={scene.words} class="text-fc-md leading-relaxed text-fc-fg" />
			</div>
		</div>
	</div>

	<div class="flex items-center gap-2 rounded-[31px] bg-fc-field p-2">
		<span class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill text-fc-fg-muted">
			<Icon icon={icons.plus} size={20} />
		</span>
		<span class="min-w-0 flex-1 truncate py-2.5 text-fc-md" class:text-fc-fg-muted={!scene.draft} class:text-fc-fg={!!scene.draft}>
			{scene.draft || t('composer.placeholder')}
		</span>
		<span
			class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill bg-fc-accent text-fc-accent-fg transition-opacity duration-150"
			class:opacity-40={!scene.draft && !streaming}
		>
			<Icon icon={streaming ? CHAT_ICONS.stop : CHAT_ICONS.send} size={18} />
		</span>
	</div>
</div>

<style>
	.card {
		background: var(--color-fc-component);
	}
	@media (prefers-color-scheme: dark) {
		:global(:root:not(.light)) .card {
			background: var(--color-fc-surface);
		}
	}
	:global(:root.dark) .card {
		background: var(--color-fc-surface);
	}
	.beat {
		transition:
			opacity 0.4s cubic-bezier(0.25, 0.8, 0.25, 1),
			transform 0.4s cubic-bezier(0.25, 0.8, 0.25, 1);
	}
	.beat:not(.on) {
		opacity: 0;
		transform: translateY(8px);
	}
	.fade {
		transition: opacity 0.2s ease;
	}
	.fade:not(.on) {
		opacity: 0;
	}
	.late.on {
		transition-delay: 0.35s;
	}
	.press {
		transition: transform 0.1s ease;
	}
	.press.down {
		transform: scale(0.95);
	}
</style>
