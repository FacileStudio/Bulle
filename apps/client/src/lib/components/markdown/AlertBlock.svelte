<script lang="ts">
	import { richBlock } from './MarkdownMuseRich';
	import type { AlertData } from './MarkdownMuseUtils';

	let { alert }: { alert: AlertData } = $props();

	const TONES = {
		pros: { bg: 'bg-fc-success/10', chip: 'bg-fc-success/20', text: 'text-fc-success', mark: '+' },
		tip: { bg: 'bg-fc-success/10', chip: 'bg-fc-success/20', text: 'text-fc-success', mark: '+' },
		cons: { bg: 'bg-fc-danger/10', chip: 'bg-fc-danger/20', text: 'text-fc-danger', mark: '-' },
		caution: { bg: 'bg-fc-danger/10', chip: 'bg-fc-danger/20', text: 'text-fc-danger', mark: '-' },
		warning: { bg: 'bg-fc-warning/10', chip: 'bg-fc-warning/20', text: 'text-fc-warning', mark: '!' },
		important: { bg: 'bg-fc-accent/10', chip: 'bg-fc-accent/20', text: 'text-fc-accent', mark: '*' },
		note: { bg: 'bg-fc-info/10', chip: 'bg-fc-info/20', text: 'text-fc-info', mark: 'i' }
	} as const;

	const tone = $derived(TONES[alert.tone] ?? TONES.note);
</script>

<div class="flex flex-col gap-1.5 rounded-fc-lg p-4 text-fc-sm {tone.bg}" role="note">
	<div class="flex items-center gap-2 font-semibold {tone.text}">
		<span
			class="inline-flex size-5 shrink-0 items-center justify-center rounded-full text-fc-xs font-bold {tone.chip}"
			aria-hidden="true">{tone.mark}</span
		>
		<span>{alert.title}</span>
	</div>
	<div class="md-flow pl-7 text-fc-sm text-fc-fg">
		{@html richBlock(alert.body)}
	</div>
</div>
