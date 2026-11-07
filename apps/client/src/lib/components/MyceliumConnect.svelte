<script lang="ts">
	import { Button, EmptyState, icons } from '@facile/muse';

	let { onnavigate }: { onnavigate?: () => void } = $props();
</script>

<EmptyState
	icon={icons.plug}
	title="Connect Mycelium to keep artifacts"
	description="Artifacts and memory live in Mycelium. Connect it in Settings, Apps, and the assistant can save documents you can come back to."
>
	<Button href="/account/apps" icon={icons.settings} class="h-11" onclick={() => onnavigate?.()}>
		Open Settings, Apps
	</Button>
</EmptyState>
