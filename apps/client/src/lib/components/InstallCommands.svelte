<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Icon, icons, Tooltip } from '@facile/muse';
	import { INSTALL_METHODS } from '$lib/downloads';

	let copiedIndex = $state<number | null>(null);

	async function copyCommand(command: string, index: number): Promise<void> {
		try {
			await navigator.clipboard.writeText(command);
			copiedIndex = index;
			setTimeout(() => {
				if (copiedIndex === index) {
					copiedIndex = null;
				}
			}, 2000);
		} catch {}
	}
</script>

<div class="flex flex-col gap-4">
	{#each INSTALL_METHODS as method, index (method.command)}
		<div class="flex flex-col gap-1.5">
			<span class="text-fc-xs font-medium text-fc-fg-muted">{method.label}</span>
			<div
				class="flex items-center justify-between gap-3 rounded-fc-md bg-fc-surface px-3.5 py-2.5 font-fc-mono text-fc-sm text-fc-fg"
			>
				<div class="flex min-w-0 items-center gap-2 overflow-x-auto">
					<span class="select-none text-fc-fg-muted">$</span>
					<span class="whitespace-nowrap">{method.command}</span>
				</div>
				<Tooltip text={copiedIndex === index ? t('common.copied') : t('downloads.copyMethod', { label: method.label })} describe={false} class="contents">
					<button
						type="button"
						class="-my-2 -mr-2 inline-flex size-11 shrink-0 items-center justify-center rounded-fc-sm text-fc-fg-muted transition-colors hover:bg-fc-page hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
						aria-label={copiedIndex === index ? t('common.copied') : t('downloads.copyMethod', { label: method.label })}
						onclick={() => copyCommand(method.command, index)}
					>
						<Icon icon={copiedIndex === index ? icons.check : icons.copy} size={16} />
					</button>
				</Tooltip>
			</div>
		</div>
	{/each}
</div>
