<script lang="ts">
	import { Badge as MuseBadge } from '@facile/muse';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	type Tone = 'neutral' | 'accent' | 'info' | 'success' | 'warning' | 'danger';

	let { props }: ComponentRenderProps<{ text?: string; tone?: Tone | null }> = $props();
</script>

<MuseBadge tone={props.tone ?? 'neutral'}>{props.text ?? ''}</MuseBadge>
