<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Badge, Button, Icon, Input, Section, Spinner, toast } from '@facile/muse';
	import { BRAND_ICONS } from '$lib/brand-icons';
	import {
		CREDENTIAL_BACKENDS,
		deleteCredential,
		listCredentials,
		putCredential,
		testCredential,
		type Credential,
		type CredentialBackend,
		type CredentialSource
	} from '$lib/api';
	import { t } from '$lib/i18n/index.svelte';

	const LABELS: Record<CredentialBackend, string> = {
		anthropic: 'Anthropic',
		openai: 'OpenAI',
		google: 'Google',
		openrouter: 'OpenRouter'
	};

	const SOURCE_TONE = { user: 'success', instance: 'info', none: 'neutral' } as const;
	const SOURCE_LABEL: Record<CredentialSource, () => string> = {
		user: () => t('providers.source.user'),
		instance: () => t('providers.source.instance'),
		none: () => t('providers.source.none')
	};

	let rows = $state<Credential[]>([]);
	let loading = $state(true);
	let loadError = $state<string | null>(null);
	let inputs = $state<Record<string, string>>(Object.fromEntries(CREDENTIAL_BACKENDS.map((b) => [b, ''])));
	let busy = $state<Record<string, 'save' | 'remove' | 'test' | undefined>>({});

	function rowFor(backend: CredentialBackend): Credential {
		return (
			rows.find((r) => r.backend === backend) ?? {
				backend,
				last4: '',
				updated_at: '',
				source: 'none'
			}
		);
	}

	async function refresh() {
		try {
			rows = await listCredentials();
			loadError = null;
		} catch (err) {
			loadError = err instanceof Error ? err.message : t('providers.loadFailed');
		}
	}

	onMount(async () => {
		await refresh();
		loading = false;
	});

	async function run(backend: CredentialBackend, kind: 'save' | 'remove' | 'test', fn: () => Promise<void>) {
		busy[backend] = kind;
		try {
			await fn();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('providers.requestFailed'));
		} finally {
			busy[backend] = undefined;
		}
	}

	function save(backend: CredentialBackend) {
		const key = (inputs[backend] ?? '').trim();
		if (!key) return;
		return run(backend, 'save', async () => {
			await putCredential(backend, key);
			inputs[backend] = '';
			await refresh();
			toast.success(t('providers.saved', { provider: LABELS[backend] }));
		});
	}

	function remove(backend: CredentialBackend) {
		return run(backend, 'remove', async () => {
			await deleteCredential(backend);
			await refresh();
			toast.success(t('providers.removed', { provider: LABELS[backend] }));
		});
	}

	function test(backend: CredentialBackend) {
		return run(backend, 'test', async () => {
			const result = await testCredential(backend);
			if (result.ok) toast.success(t('providers.works', { provider: LABELS[backend] }));
			else toast.danger(result.error || t('providers.testFailed', { provider: LABELS[backend] }));
		});
	}
</script>

<Section
	card
	id="providers"
	title={t('providers.title')}
	description={t('providers.description')}
>
	{#if loading}
		<div class="flex justify-center py-6"><Spinner /></div>
	{:else}
		<div class="flex flex-col">
			{#if loadError}
				<div class="pb-4"><Alert tone="danger">{loadError}</Alert></div>
			{/if}
			{#each CREDENTIAL_BACKENDS as backend (backend)}
				{@const row = rowFor(backend)}
				{@const state = busy[backend]}
				<form
					class="flex flex-col gap-3 py-4 first:pt-0 last:pb-0"
					onsubmit={(e) => {
						e.preventDefault();
						void save(backend);
					}}
				>
					<div class="flex flex-wrap items-center gap-3">
						<Icon icon={BRAND_ICONS[backend]} size={18} class="text-fc-fg" />
						<span class="text-fc-sm font-medium text-fc-fg">{LABELS[backend]}</span>
						<Badge tone={SOURCE_TONE[row.source]}>{SOURCE_LABEL[row.source]()}</Badge>
						{#if row.last4}
							<span class="font-mono text-fc-sm text-fc-fg-muted">••••{row.last4}</span>
						{/if}
					</div>
					<div class="flex flex-col gap-2 sm:flex-row">
						<Input
							type="password"
							autocomplete="off"
							spellcheck={false}
							aria-label={t('providers.keyAria', { provider: LABELS[backend] })}
							placeholder={row.source === 'user' ? t('providers.pasteReplace') : t('providers.paste')}
							bind:value={inputs[backend]}
						/>
						<div class="flex flex-wrap gap-2">
							<Button type="submit" variant="primary" disabled={!!state || !(inputs[backend] ?? '').trim()}>
								{#if state === 'save'}<Spinner size="sm" />{/if}
								{t('common.save')}
							</Button>
							<Button
								variant="outline"
								disabled={!!state || row.source === 'none'}
								onclick={() => test(backend)}
							>
								{#if state === 'test'}<Spinner size="sm" />{/if}
								{t('providers.test')}
							</Button>
							<Button
								variant="outline"
								disabled={!!state || row.source !== 'user'}
								onclick={() => remove(backend)}
							>
								{#if state === 'remove'}<Spinner size="sm" />{/if}
								{t('providers.remove')}
							</Button>
						</div>
					</div>
				</form>
			{/each}
		</div>
	{/if}
</Section>
