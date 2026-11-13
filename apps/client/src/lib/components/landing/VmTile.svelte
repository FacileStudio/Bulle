<script lang="ts">
	import { Icon, StatusDot, icons, Tooltip } from '@facile/muse';
	import '$lib/suite-icons';
	import { t } from '$lib/i18n/index.svelte';
	import DemoTile from './DemoTile.svelte';
	import { playInView, prefersReducedMotion, sleep, typeOut } from './stage';

	type Line = { kind: 'cmd' | 'ok' | 'out'; text: string };

	const SCRIPT: Line[] = [
		{ kind: 'cmd', text: 'bun test' },
		{ kind: 'ok', text: '✓ 48 pass  1.21s' },
		{ kind: 'cmd', text: 'git push' },
		{ kind: 'out', text: '→ origin fix/session-refresh' }
	];

	let lines = $state<Line[]>([]);

	function reset() {
		lines = [];
	}

	async function script(signal: AbortSignal) {
		await sleep(600, signal);
		for (const line of SCRIPT) {
			if (line.kind === 'cmd') {
				lines.push({ kind: 'cmd', text: '' });
				const at = lines.length - 1;
				await typeOut(line.text, (v) => (lines[at].text = v), 70, signal);
				await sleep(450, signal);
			} else {
				lines.push(line);
				await sleep(900, signal);
			}
		}
		await sleep(1800, signal);
		lines.push({ kind: 'cmd', text: '' });
		const at = lines.length - 1;
		await typeOut('clear', (v) => (lines[at].text = v), 80, signal);
		await sleep(350, signal);
		reset();
	}

	if (prefersReducedMotion()) lines = [...SCRIPT];
</script>

{#snippet boiteBadge()}
	<Tooltip text="Boite" describe={false} class="contents">
		<a
			href="https://boite.facile.studio"
			target="_blank"
			rel="noreferrer"
			class="inline-flex size-6 shrink-0 items-center justify-center rounded-full bg-fc-surface text-fc-accent transition-colors hover:bg-fc-field hover:text-fc-fg"
			aria-label="Boite"
		>
			<Icon icon="solar:box-bold-duotone" size={14} />
		</a>
	</Tooltip>
{/snippet}

<DemoTile title={t('landing.feature.vm.title')} body={t('landing.feature.vm.body')} badge={boiteBadge}>
	<div
		class="term absolute top-6 left-6 -right-10 -bottom-6 rounded-fc-lg px-4 py-3 font-fc-mono text-fc-xs"
		{@attach playInView(script, reset)}
	>
		<div class="mb-3 flex items-center gap-2 text-white/60">
			<StatusDot tone="success" pulse />
			<span class="text-white/90">bulle-vm</span>
			<span>up 14d</span>
		</div>
		<div class="flex flex-col gap-1 leading-5">
			{#each lines as line, i (i)}
				<div class="whitespace-nowrap">
					{#if line.kind === 'cmd'}
						<span class="text-white/45">~/app $</span>
						<span class="text-white">{line.text}</span>{#if i === lines.length - 1}<span class="caret"></span>{/if}
					{:else if line.kind === 'ok'}
						<span class="text-[oklch(0.8_0.15_150)]">{line.text}</span>
					{:else}
						<span class="text-white/60">{line.text}</span>
					{/if}
				</div>
			{/each}
			{#if lines.length === 0 || lines.at(-1)?.kind !== 'cmd'}
				<div><span class="text-white/45">~/app $</span> <span class="caret"></span></div>
			{/if}
		</div>
	</div>
</DemoTile>

<style>
	.term {
		background: oklch(0.18 0 0);
		box-shadow: 0 18px 40px -16px rgb(0 0 0 / 0.35);
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
