<script lang="ts">
	import { Alert, SignInWithFacile } from '@facile/muse';
	import { ssoLoginUrl } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';
	import { startNativeLogin } from '$lib/native';

	let { error = $bindable('') }: { error?: string } = $props();

	let opening = $state(false);

	async function signIn(event: MouseEvent) {
		event.preventDefault();
		if (opening) return;
		error = '';
		opening = true;
		try {
			await startNativeLogin();
		} catch (err) {
			error = err instanceof Error ? err.message : t('login.openFailed');
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
	<SignInWithFacile href={ssoLoginUrl} loading={opening} onclick={signIn} />
	<p class="max-w-xs text-center text-fc-xs text-fc-fg-muted">
		{t('login.browserHint')}
	</p>
</div>
