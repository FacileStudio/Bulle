<script lang="ts">
	import { tick } from 'svelte';
	import { fade } from 'svelte/transition';
	import { Blob } from '@facile/blob';
	import { Button, Icon, Spinner, StatusDot, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import { kindLabel } from '$lib/chat/tools';
	import Cursor from './Cursor.svelte';
	import { grow, playInView, prefersReducedMotion, sleep, typeOut } from './stage';

	type ToolState = 'hidden' | 'running' | 'ok' | 'failed';
	type ApprovalState = 'hidden' | 'pending' | 'pressed' | 'approved';

	const TOOLS = [
		{ name: 'bash', args: 'bun test', ms: '1.8s', end: 'failed' },
		{ name: 'edit_file', args: 'src/auth/session.ts', ms: '0.2s', end: 'ok' },
		{ name: 'bash', args: 'bun test', ms: '1.6s', end: 'ok' }
	] as const;
	const BRANCH = 'fix/session-refresh';
	const PR_TITLE = 'fix(auth): refresh expired sessions';
	const WIDE = { width: 960, height: 580 };
	const COMPACT = { width: 358, height: 560 };

	const prompt = $derived(t('landing.demo.prompt'));
	const reply = $derived(t('landing.demo.reply').split(' '));

	let draft = $state('');
	let sent = $state(false);
	let assistant = $state(false);
	let thinking = $state(false);
	let tools = $state<ToolState[]>(['hidden', 'hidden', 'hidden']);
	let approval = $state<ApprovalState>('hidden');
	let words = $state(0);
	let cursor = $state({ x: 0, y: 0, visible: false, down: false });

	let hostWidth = $state(0);
	let stageEl = $state<HTMLElement>();
	let approveEl = $state<HTMLElement>();

	const compact = $derived(hostWidth > 0 && hostWidth < 640);
	const design = $derived(compact ? COMPACT : WIDE);
	const scale = $derived(hostWidth ? Math.min(1, hostWidth / design.width) : 1);
	const toolsShown = $derived(tools.some((s) => s !== 'hidden'));
	const anyRunning = $derived(tools.includes('running'));

	function reset() {
		draft = '';
		sent = false;
		assistant = false;
		thinking = false;
		tools = ['hidden', 'hidden', 'hidden'];
		approval = 'hidden';
		words = 0;
		cursor.visible = false;
		cursor.down = false;
	}

	function poster() {
		sent = true;
		assistant = true;
		tools = TOOLS.map((tool) => tool.end);
		approval = 'approved';
		words = reply.length;
	}

	function approvePoint(fx: number, fy: number) {
		if (!stageEl || !approveEl) return { x: 0, y: 0 };
		const s = stageEl.getBoundingClientRect();
		const b = approveEl.getBoundingClientRect();
		return { x: (b.left - s.left + b.width * fx) / scale, y: (b.top - s.top + b.height * fy) / scale };
	}

	async function script(signal: AbortSignal) {
		await sleep(800, signal);
		await typeOut(prompt, (v) => (draft = v), 38, signal);
		await sleep(350, signal);
		sent = true;
		draft = '';
		await sleep(500, signal);
		assistant = true;
		thinking = true;
		await sleep(1100, signal);
		thinking = false;
		for (const [i, tool] of TOOLS.entries()) {
			await sleep(300, signal);
			tools[i] = 'running';
			await sleep(i === 1 ? 700 : 950, signal);
			tools[i] = tool.end;
		}
		await sleep(400, signal);
		approval = 'pending';
		await tick();
		const rest = approvePoint(2.4, 1.6);
		cursor.x = rest.x;
		cursor.y = rest.y;
		await sleep(900, signal);
		cursor.visible = true;
		await sleep(400, signal);
		const target = approvePoint(0.9, 0.78);
		cursor.x = target.x;
		cursor.y = target.y;
		await sleep(1000, signal);
		cursor.down = true;
		approval = 'pressed';
		await sleep(140, signal);
		cursor.down = false;
		await sleep(160, signal);
		approval = 'approved';
		cursor.x = rest.x;
		cursor.y = rest.y + 40;
		await sleep(500, signal);
		for (let i = 1; i <= reply.length; i++) {
			words = i;
			await sleep(45, signal);
		}
		await sleep(2200, signal);
		cursor.visible = false;
		reset();
		await sleep(700, signal);
	}

	if (prefersReducedMotion()) poster();
</script>

<div
	class="relative w-full"
	style:height="{design.height * scale}px"
	bind:clientWidth={hostWidth}
	aria-hidden="true"
	inert
	{@attach playInView(script, reset)}
>
	<div
		bind:this={stageEl}
		class="absolute top-0 left-1/2 origin-top overflow-hidden text-left"
		style:width="{design.width}px"
		style:height="{design.height}px"
		style:transform="translateX(-50%) scale({scale})"
	>
		<div
			class="window absolute flex flex-col overflow-hidden rounded-fc-lg"
			class:compact
		>
			<div class="side flex h-[38px] shrink-0 items-center gap-2 px-3.5">
				<span class="size-3 rounded-full bg-fc-surface"></span>
				<span class="size-3 rounded-full bg-fc-surface"></span>
				<span class="size-3 rounded-full bg-fc-surface"></span>
				{#if compact}
					<span class="ml-auto flex items-center gap-2 text-fc-xs text-fc-fg-muted">
						<StatusDot tone="success" pulse />
						{t('landing.demo.vm')}
					</span>
				{/if}
			</div>

			<div class="flex min-h-0 flex-1">
				{#if !compact}
					<aside class="side flex w-[208px] shrink-0 flex-col gap-1 px-3 pt-4">
						<div class="mb-3 flex items-center gap-2 px-2">
							<Blob name="Bulle" size={22} />
							<span class="text-fc-sm font-medium text-fc-fg">Bulle</span>
						</div>
						<div class="rounded-fc-md bg-fc-surface px-2 py-2.5"><span class="bar w-[72%]"></span></div>
						<div class="px-2 py-2.5"><span class="bar w-[58%]"></span></div>
						<div class="px-2 py-2.5"><span class="bar w-[80%]"></span></div>
						<div class="px-2 py-2.5"><span class="bar w-[46%]"></span></div>
						<div class="px-2 py-2.5"><span class="bar w-[64%]"></span></div>
						<div class="px-2 py-2.5"><span class="bar w-[70%]"></span></div>
						<div class="px-2 py-2.5"><span class="bar w-[40%]"></span></div>
					</aside>
				{/if}

				<div class="flex min-w-0 flex-1 flex-col px-6" class:px-4={compact} style:padding-bottom="20px">
					<div class="thread flex min-h-0 flex-1 flex-col justify-end overflow-hidden pb-4">
						<div class="flex flex-col gap-2.5 pb-2 opacity-70">
							<span class="ml-auto h-9 w-[46%] rounded-fc-lg bg-fc-surface"></span>
							<div class="flex items-center gap-2 pt-3">
								<Blob name="Bulle" size={24} />
								<span class="text-fc-xs font-medium text-fc-fg">Bulle</span>
							</div>
							<div class="flex items-center gap-2 py-1 text-fc-fg-muted">
								<Icon icon={CHAT_ICONS.terminal} size={16} />
								<span class="bar w-[28%]"></span>
							</div>
							<span class="bar w-[88%]"></span>
							<span class="bar w-[74%]"></span>
							<span class="bar w-[52%]"></span>
						</div>

						{#if sent}
							<div class="flex justify-end pt-4" in:grow out:fade={{ duration: 200, delay: 240 }}>
								<p class="max-w-[85%] rounded-fc-lg bg-fc-surface px-4 py-2 text-fc-md text-fc-fg">{prompt}</p>
							</div>
						{/if}

						{#if assistant}
							<div class="flex items-center gap-2 pt-5" in:grow out:fade={{ duration: 200, delay: 180 }}>
								<Blob name="Bulle" size={24} />
								<span class="text-fc-xs font-medium text-fc-fg">Bulle</span>
							</div>
						{/if}

						{#if thinking}
							<div class="flex items-center gap-2 pt-3 text-fc-sm text-fc-fg-muted" in:grow out:grow={{ duration: 300 }}>
								<Icon icon={CHAT_ICONS.chevronRight} size={14} />
								<span class="font-medium">{t('thinking.active')}</span>
							</div>
						{/if}

						{#if toolsShown}
							<div class="pt-3" in:grow out:fade={{ duration: 200, delay: 120 }}>
								<section class="card overflow-hidden rounded-fc-lg">
									<header class="flex items-center gap-2 px-3 py-2 text-fc-sm">
										<Icon icon={CHAT_ICONS.terminal} size={16} class="text-fc-fg-muted" />
										<span class="font-medium text-fc-fg">{kindLabel('commands')}</span>
										<span class="ml-auto flex size-4 items-center justify-center">
											{#if anyRunning}
												<span transition:fade={{ duration: 150 }}><Spinner size="sm" /></span>
											{/if}
										</span>
									</header>
									<ul class="flex flex-col pb-1">
										{#each TOOLS as tool, i (i)}
											{#if tools[i] !== 'hidden'}
												<li class="flex min-h-11 items-center gap-3 px-3 py-2 text-fc-sm" in:grow={{ duration: 350 }}>
													<span class="flex size-4 shrink-0 items-center justify-center">
														{#if tools[i] === 'running'}
															<Spinner size="sm" />
														{:else if tools[i] === 'ok'}
															<span in:fade={{ duration: 150 }}><Icon icon={icons.check} size={16} class="text-fc-success" /></span>
														{:else}
															<span in:fade={{ duration: 150 }}><Icon icon={icons.error} size={16} class="text-fc-danger" /></span>
														{/if}
													</span>
													<span class="shrink-0 font-medium text-fc-fg">{tool.name}</span>
													<span class="min-w-0 flex-1 truncate font-fc-mono text-fc-xs text-fc-fg-muted">{tool.args}</span>
													<span
														class="shrink-0 text-fc-xs text-fc-fg-muted transition-opacity duration-200"
														class:opacity-0={tools[i] === 'running'}
													>
														{tool.ms}
													</span>
												</li>
											{/if}
										{/each}
									</ul>
								</section>
							</div>
						{/if}

						{#if approval === 'pending' || approval === 'pressed'}
							<div class="flex flex-col gap-4 pt-5 pb-2" in:grow out:grow={{ duration: 350 }}>
								<div class="flex flex-col gap-0.5">
									<span class="text-fc-md font-medium text-fc-fg">{t('landing.demo.pr')}</span>
									<span class="text-fc-xs text-fc-fg-muted">{t('approval.needed')}</span>
								</div>
								<dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-4 gap-y-2 text-fc-sm">
									<dt class="text-fc-fg-muted capitalize">title</dt>
									<dd class="truncate font-fc-mono text-fc-xs leading-5 text-fc-fg">{PR_TITLE}</dd>
									<dt class="text-fc-fg-muted capitalize">branch</dt>
									<dd class="truncate font-fc-mono text-fc-xs leading-5 text-fc-fg">{BRANCH}</dd>
								</dl>
								<div class="flex items-center gap-2">
									<span
										bind:this={approveEl}
										class="inline-flex transition-transform duration-100"
										class:scale-95={approval === 'pressed'}
									>
										<Button tabindex={-1}>{t('approval.approve')}</Button>
									</span>
									<Button variant="ghost" tabindex={-1}>{t('approval.deny')}</Button>
								</div>
							</div>
						{:else if approval === 'approved'}
							<p class="flex items-baseline gap-2 pt-4 text-fc-sm text-fc-fg-muted" in:grow out:fade={{ duration: 200, delay: 60 }}>
								<span class="shrink-0 whitespace-nowrap">{t('landing.demo.pr')}</span>
								<span class="min-w-0 truncate font-fc-mono text-fc-xs">{BRANCH}</span>
								<span class="text-fc-xs text-fc-success">{t('approval.approved')}</span>
							</p>
						{/if}

						{#if words > 0}
							<p class="pt-3 text-fc-md leading-relaxed text-fc-fg" in:grow={{ duration: 300 }} out:fade={{ duration: 200 }}>
								{#each reply as word, i (i)}
									<span class="transition-opacity duration-150" class:opacity-0={i >= words}>{word}</span>{' '}
								{/each}
							</p>
						{/if}
					</div>

					<div class="flex items-center gap-3 rounded-fc-lg bg-fc-surface py-2 pr-2 pl-4">
						<span class="min-w-0 flex-1 truncate text-fc-md" class:text-fc-fg-muted={!draft} class:text-fc-fg={draft}>
							{draft || t('composer.placeholder')}
						</span>
						<span
							class="flex size-9 shrink-0 items-center justify-center rounded-fc-md transition-colors duration-150"
							class:bg-fc-accent={!!draft}
							class:text-fc-accent-fg={!!draft}
							class:text-fc-fg-muted={!draft}
						>
							<Icon icon={CHAT_ICONS.send} size={18} />
						</span>
					</div>
				</div>
			</div>
		</div>

		{#if !compact}
			<div class="chip absolute z-20 flex items-center gap-2 rounded-fc-pill bg-fc-page px-3.5 py-2 text-fc-xs text-fc-fg">
				<StatusDot tone="success" pulse />
				<span class="font-medium whitespace-nowrap">{t('landing.demo.vm')}</span>
			</div>
		{/if}

		<Cursor {...cursor} />
	</div>
</div>

<style>
	.window {
		--w-body: var(--color-fc-page);
		--w-side: var(--color-fc-component);
		--w-card: var(--color-fc-component);
		background: var(--w-body);
		top: 28px;
		left: 72px;
		width: 816px;
		height: 516px;
		box-shadow:
			0 30px 70px -20px rgb(0 0 0 / 0.22),
			0 8px 24px -8px rgb(0 0 0 / 0.1);
	}
	@media (prefers-color-scheme: dark) {
		:global(:root:not(.light)) .window {
			--w-body: oklch(0.13 0 0);
			--w-side: oklch(0.16 0 0);
			--w-card: oklch(0.18 0 0);
		}
	}
	:global(:root.dark) .window {
		--w-body: oklch(0.13 0 0);
		--w-side: oklch(0.16 0 0);
		--w-card: oklch(0.18 0 0);
	}
	.side {
		background: var(--w-side);
	}
	.card {
		background: var(--w-card);
	}
	.window.compact {
		top: 16px;
		left: 16px;
		width: 326px;
		height: 504px;
	}
	.chip {
		top: 28px;
		left: 888px;
		translate: -50% -50%;
		box-shadow: 0 8px 24px -6px rgb(0 0 0 / 0.18);
	}
	.bar {
		display: block;
		height: 8px;
		border-radius: 4px;
		background: color-mix(in oklch, var(--color-fc-fg) 9%, transparent);
	}
	.thread {
		mask-image: linear-gradient(to bottom, transparent, #000 56px);
	}
</style>
