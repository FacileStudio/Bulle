<script lang="ts">
	import { getCurrentWindow } from '@tauri-apps/api/window';

	type Control = { label: string; color: string; glyph: string; run: () => Promise<void> };

	const win = getCurrentWindow();

	const controls: Control[] = [
		{ label: 'Close', color: 'bg-[#ff5f57]', glyph: 'M3.5 3.5l5 5m0-5l-5 5', run: () => win.close() },
		{ label: 'Minimize', color: 'bg-[#febc2e]', glyph: 'M3 6h6', run: () => win.minimize() },
		{
			label: 'Maximize',
			color: 'bg-[#28c840]',
			glyph: 'M3.5 6h5M6 3.5v5',
			run: () => win.toggleMaximize()
		}
	];
</script>

<div class="group fixed top-[34px] left-[28px] z-50 flex items-center gap-2" role="group" aria-label="Window">
	{#each controls as control (control.label)}
		<button
			type="button"
			aria-label={control.label}
			title={control.label}
			onclick={() => void control.run()}
			class="flex size-3 items-center justify-center rounded-full {control.color} ring-1 ring-black/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
		>
			<svg viewBox="0 0 12 12" class="size-2 opacity-0 transition-opacity group-hover:opacity-100" aria-hidden="true">
				<path d={control.glyph} stroke="rgb(0 0 0 / 0.6)" stroke-width="1.3" stroke-linecap="round" fill="none" />
			</svg>
		</button>
	{/each}
</div>
