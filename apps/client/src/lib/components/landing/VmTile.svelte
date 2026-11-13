<script lang="ts">
	import { Icon, Tooltip } from '@facile/muse';
	import { Caret, Stage, Terminal, Tile, type Act } from '@facile/mise-en-scene';
	import '$lib/suite-icons';
	import { t } from '$lib/i18n/index.svelte';

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

	function poster() {
		lines = [...SCRIPT];
	}

	async function typeLine(act: Act, text: string, interval: number) {
		lines.push({ kind: 'cmd', text: '' });
		const at = lines.length - 1;
		await act.type(text, (v) => (lines[at].text = v), interval);
	}

	async function script(act: Act) {
		await act.wait(600);
		for (const line of SCRIPT) {
			if (line.kind === 'cmd') {
				await typeLine(act, line.text, 70);
				await act.wait(450);
			} else {
				lines.push(line);
				await act.wait(900);
			}
		}
		await act.wait(1800);
		await typeLine(act, 'clear', 80);
		await act.wait(350);
		reset();
	}
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

<Tile title={t('landing.feature.vm.title')} body={t('landing.feature.vm.body')} badge={boiteBadge}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Terminal style="top:24px;left:24px;right:-40px;bottom:-24px">
				<div class="mb-3 flex items-center gap-2">
					<span class="size-2 rounded-full bg-fc-success"></span>
					<span>bulle-vm</span>
					<span class="dim">up 14d</span>
				</div>
				{#each lines as line, i (i)}
					<div class="whitespace-nowrap">
						{#if line.kind === 'cmd'}
							<span class="dim">~/app $</span>
							{line.text}{#if i === lines.length - 1}<Caret />{/if}
						{:else}
							<span class:ok={line.kind === 'ok'} class:dim={line.kind === 'out'}>{line.text}</span>
						{/if}
					</div>
				{/each}
				{#if lines.at(-1)?.kind !== 'cmd'}
					<div class="whitespace-nowrap"><span class="dim">~/app $</span> <Caret /></div>
				{/if}
			</Terminal>
		{/snippet}
	</Stage>
</Tile>

<style>
	.dim {
		color: rgb(255 255 255 / 0.5);
	}
	.ok {
		color: color-mix(in oklch, var(--color-fc-success), white 35%);
	}
</style>
