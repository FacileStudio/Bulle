<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import type { AppKey } from '$lib/i18n/messages';
	import { getCurrentWindow } from '@tauri-apps/api/window';

	type Control = { label: string; color: string; glyph: string; run: () => Promise<void> };

	function controlLabel(id: string): string {
		return t(`native.window.${id}` as AppKey);
	}

	const win = getCurrentWindow();

	const controls: Control[] = [
		{ label: 'close', color: 'bg-fc-danger', glyph: 'M3.5 3.5l5 5m0-5l-5 5', run: () => win.close() },
		{ label: 'minimize', color: 'bg-fc-warning', glyph: 'M3 6h6', run: () => win.minimize() },
		{
			label: 'maximize',
			color: 'bg-fc-success',
			glyph: 'M3.5 6h5M6 3.5v5',
			run: () => win.toggleMaximize()
		}
	];
</script>

<div class="group fixed top-[34px] left-[28px] z-50 flex items-center gap-2" role="group" aria-label={t('native.windowGroup')}>
	{#each controls as control (control.label)}
		<button
			type="button"
			aria-label={controlLabel(control.label)}
			title={controlLabel(control.label)}
			onclick={() => void control.run()}
			class="flex size-3 items-center justify-center rounded-full {control.color} focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
		>
			<svg viewBox="0 0 12 12" class="size-2 opacity-0 transition-opacity group-hover:opacity-100" aria-hidden="true">
				<path d={control.glyph} stroke="currentColor" stroke-width="1.3" stroke-linecap="round" fill="none" />
			</svg>
		</button>
	{/each}
</div>
