<script lang="ts">
	import { Button, EmptyState, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let { onnavigate }: { onnavigate?: () => void } = $props();
</script>

<EmptyState
	icon={icons.plug}
	title={t('connect.title')}
	description={t('connect.description')}
>
	<Button href="/account/apps" icon={icons.settings} class="h-11" onclick={() => onnavigate?.()}>
		{t('connect.open')}
	</Button>
</EmptyState>
