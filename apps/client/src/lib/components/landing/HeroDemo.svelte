<script lang="ts">
	import { tick } from 'svelte';
	import { fade } from 'svelte/transition';
	import { Blob } from '@facile/blob';
	import { Button, Icon, Spinner, StatusDot, icons } from '@facile/muse';
	import { Bar, Cursor, Stage, Stream, Window, cursorAt, grow, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import { kindLabel } from '$lib/chat/tools';

	type ToolState = 'hidden' | 'running' | 'ok' | 'failed';
	type ApprovalState = 'hidden' | 'pending' | 'pressed' | 'approved';

	const TOOLS = [
		{ name: 'bash', args: 'bun test', ms: '1.8s', end: 'failed' },
		{ name: 'edit_file', args: 'src/auth/session.ts', ms: '0.2s', end: 'ok' },
		{ name: 'bash', args: 'bun test', ms: '1.6s', end: 'ok' }
	] as const;
	const BRANCH = 'fix/session-refresh';
	const PR_TITLE = 'fix(auth): refresh expired sessions';
	const HISTORY = ['72%', '58%', '80%', '46%', '64%', '70%', '40%'];

	const prompt = $derived(t('landing.demo.prompt'));
	const reply = $derived(t('landing.demo.reply'));
	const replyWords = $derived(reply.split(' ').length);

	let draft = $state('');
	let sent = $state(false);
	let assistant = $state(false);
	let thinking = $state(false);
	let tools = $state<ToolState[]>(['hidden', 'hidden', 'hidden']);
	let approval = $state<ApprovalState>('hidden');
	let words = $state(0);
	let cursor = $state(cursorAt());
	let approveEl = $state<HTMLElement>();

	const toolsShown = $derived(tools.some((s) => s !== 'hidden'));
	const anyRunning = $derived(tools.includes('running'));
	const running = $derived(sent && words < replyWords);

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
		words = replyWords;
	}

	async function runTools(act: Act) {
		for (const [i, tool] of TOOLS.entries()) {
			await act.wait(300);
			tools[i] = 'running';
			await act.wait(i === 1 ? 700 : 950);
			tools[i] = tool.end;
		}
	}

	async function approve(act: Act) {
		approval = 'pending';
		await tick();
		const rest = act.point(approveEl, 2.4, 1.6);
		await act.move(cursor, rest, 900);
		cursor.visible = true;
		await act.wait(400);
		await act.move(cursor, act.point(approveEl, 0.9, 0.78), 1000);
		await act.click(cursor, () => (approval = 'pressed'));
		approval = 'approved';
		cursor.y = rest.y + 40;
	}

	async function script(act: Act) {
		await act.wait(800);
		await act.type(prompt, (v) => (draft = v));
		await act.wait(350);
		sent = true;
		draft = '';
		await act.wait(500);
		assistant = true;
		thinking = true;
		await act.wait(1100);
		thinking = false;
		await runTools(act);
		await act.wait(400);
		await approve(act);
		await act.wait(500);
		await act.count(replyWords, (v) => (words = v));
		await act.wait(2200);
		reset();
		await act.wait(700);
	}
</script>

{#snippet sidebar()}
	<div class="flex flex-col gap-1 px-3 pt-4">
		<div class="mb-3 flex items-center gap-2 px-2">
			<Blob name="Bulle" size={22} />
			<span class="text-fc-sm font-bold text-fc-fg">Bulle</span>
		</div>
		{#each HISTORY as w, i (i)}
			<div class="rounded-fc-md px-2 py-2.5" class:bg-fc-surface={i === 0}><Bar {w} /></div>
		{/each}
	</div>
{/snippet}

{#snippet status()}
	<span class="flex items-center gap-2 text-fc-xs text-fc-fg-muted">
		<StatusDot tone="success" pulse />
		{t('landing.demo.vm')}
	</span>
{/snippet}

{#snippet thread(compact: boolean)}
	<div class="flex min-h-0 flex-1 flex-col pb-5" class:px-6={!compact} class:px-4={compact}>
		<div class="thread flex min-h-0 flex-1 flex-col justify-end overflow-hidden pb-4">
			<div class="flex flex-col gap-2.5 pb-2 opacity-70">
				<span class="ml-auto h-9 w-[46%] rounded-fc-lg bg-fc-surface"></span>
				<div class="flex items-center gap-2 pt-3">
					<Blob name="Bulle" size={24} />
					<span class="text-fc-xs font-bold text-fc-fg">Bulle</span>
				</div>
				<div class="flex items-center gap-2 py-1 text-fc-fg-muted">
					<Icon icon={CHAT_ICONS.terminal} size={16} />
					<Bar w="28%" />
				</div>
				<Bar w="88%" />
				<Bar w="74%" />
				<Bar w="52%" />
			</div>

			{#if sent}
				<div class="flex justify-end pt-4" in:grow out:fade={{ duration: 200, delay: 240 }}>
					<p class="max-w-[85%] rounded-fc-lg bg-fc-surface px-4 py-2 text-fc-md text-fc-fg">{prompt}</p>
				</div>
			{/if}

			{#if assistant}
				<div class="flex items-center gap-2 pt-5" in:grow out:fade={{ duration: 200, delay: 180 }}>
					<Blob name="Bulle" size={24} />
					<span class="text-fc-xs font-bold text-fc-fg">Bulle</span>
				</div>
			{/if}

			{#if thinking}
				<div class="flex items-center gap-2 pt-3 text-fc-sm text-fc-fg-muted" in:grow out:grow={{ duration: 300 }}>
					<Icon icon={CHAT_ICONS.chevronRight} size={14} />
					<span>{t('thinking.active')}</span>
				</div>
			{/if}

			{#if toolsShown}
				<div class="pt-3" in:grow out:fade={{ duration: 200, delay: 120 }}>
					<section class="card overflow-hidden rounded-fc-lg">
						<header class="flex items-center gap-2 px-3 py-2 text-fc-sm">
							<Icon icon={CHAT_ICONS.terminal} size={16} class="text-fc-fg-muted" />
							<span class="text-fc-fg">{kindLabel('commands')}</span>
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
										<span class="shrink-0 text-fc-fg">{tool.name}</span>
										<code class="min-w-0 flex-1 truncate font-fc-mono text-fc-xs text-fc-fg-muted">{tool.args}</code>
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
						<span class="text-fc-md font-bold text-fc-fg">{t('landing.demo.pr')}</span>
						<span class="text-fc-xs text-fc-fg-muted">{t('approval.needed')}</span>
					</div>
					<dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-4 gap-y-2 text-fc-sm">
						<dt class="text-fc-fg-muted">title</dt>
						<dd class="truncate font-fc-mono text-fc-xs leading-5 text-fc-fg">{PR_TITLE}</dd>
						<dt class="text-fc-fg-muted">branch</dt>
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
					<code class="min-w-0 truncate font-fc-mono text-fc-xs">{BRANCH}</code>
					<span class="text-fc-xs text-fc-success">{t('approval.approved')}</span>
				</p>
			{/if}

			{#if words > 0}
				<div class="pt-3" in:grow={{ duration: 300 }} out:fade={{ duration: 200 }}>
					<Stream text={reply} count={words} class="text-fc-md leading-relaxed text-fc-fg" />
				</div>
			{/if}
		</div>

		<div class="flex items-center gap-2 rounded-[31px] bg-fc-field p-2">
			<span class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill text-fc-fg-muted">
				<Icon icon={icons.plus} size={20} />
			</span>
			<span class="min-w-0 flex-1 truncate py-2.5 text-fc-md" class:text-fc-fg-muted={!draft} class:text-fc-fg={!!draft}>
				{draft || t('composer.placeholder')}
			</span>
			<span
				class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill bg-fc-accent text-fc-accent-fg transition-opacity duration-150"
				class:opacity-40={!draft && !running}
			>
				<Icon icon={running ? CHAT_ICONS.stop : CHAT_ICONS.send} size={18} />
			</span>
		</div>
	</div>
{/snippet}

<Stage
	width={960}
	height={580}
	compact={{ below: 640, width: 358, height: 560 }}
	{script}
	{reset}
	{poster}
	label={t('landing.demo.prompt')}
>
	{#snippet children({ compact })}
		{#if compact}
			<Window style="top:16px;left:16px;width:326px;height:528px" trailing={status}>
				{@render thread(true)}
			</Window>
		{:else}
			<Window style="top:28px;left:72px;width:816px;height:516px" {sidebar}>
				{@render thread(false)}
			</Window>
			<div class="chip absolute z-20 flex items-center gap-2 rounded-fc-pill bg-fc-page px-3.5 py-2 text-fc-xs text-fc-fg">
				<StatusDot tone="success" pulse />
				<span class="font-bold whitespace-nowrap">{t('landing.demo.vm')}</span>
			</div>
		{/if}
		<Cursor {cursor} />
	{/snippet}
</Stage>

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
	.chip {
		top: 28px;
		left: 888px;
		translate: -50% -50%;
		box-shadow: 0 8px 24px -6px rgb(0 0 0 / 0.18);
	}
	.thread {
		mask-image: linear-gradient(to bottom, transparent, #000 56px);
	}
</style>
