<script lang="ts">
	import { Icon } from '@facile/muse';
	import { LOGO_ICON } from '$lib/icons';

	let { size = 24, class: className = '' }: { size?: number; class?: string } = $props();
</script>

<Icon icon={LOGO_ICON} {size} class={className} />
