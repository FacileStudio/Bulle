<script lang="ts">
	import { Alert, Button, Spinner, icons } from '@facile/muse';
	import { startNativeLogin } from '$lib/native';

	let { error = $bindable('') }: { error?: string } = $props();

	let opening = $state(false);

	async function signIn() {
		if (opening) return;
		error = '';
		opening = true;
		try {
			await startNativeLogin();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not open the sign-in page.';
		} finally {
			opening = false;
		}
	}
</script>

<div class="flex w-full flex-col items-center gap-4">
	{#if error}
		<div class="w-full max-w-sm" role="alert">
			<Alert tone="danger">{error}</Alert>
		</div>
	{/if}
	<Button
		size="lg"
		class="w-full max-w-sm rounded-fc-md"
		icon={opening ? undefined : icons.key}
		disabled={opening}
		onclick={signIn}
	>
		{#if opening}
			<Spinner size="sm" />
		{/if}
		Continue with Facile
	</Button>
	<p class="max-w-xs text-center text-fc-xs text-fc-fg-muted">
		Sign-in opens in your browser, then brings you back here.
	</p>
</div>
