<script lang="ts">
	import { Checkbox, Field as MuseField, Input, Select, Textarea } from '@facile/muse';
	import {
		getFormName,
		getGetFieldValue,
		getIsStreaming,
		getSetFieldValue
	} from '@openuidev/svelte-lang';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	type Kind = 'text' | 'number' | 'textarea' | 'select' | 'checkbox';

	let {
		props
	}: ComponentRenderProps<{
		name?: string;
		label?: string;
		kind?: Kind | null;
		options?: string[] | null;
		placeholder?: string | null;
		required?: boolean | null;
	}> = $props();

	const formName = getFormName();
	const getValue = getGetFieldValue();
	const setValue = getSetFieldValue();
	const streaming = getIsStreaming();

	const kind = $derived(props.kind ?? 'text');
	const stored = $derived(props.name ? getValue(formName, props.name) : undefined);

	function commit(value: unknown) {
		if (props.name) setValue(formName, kind, props.name, value, true);
	}

	function typed(raw: string) {
		return kind === 'number' && raw !== '' ? Number(raw) : raw;
	}
</script>

{#if kind === 'checkbox'}
	<Checkbox
		label={props.label ?? ''}
		checked={stored === true}
		disabled={streaming()}
		onchange={(e) => commit(e.currentTarget.checked)}
	/>
{:else}
	<MuseField label={props.label ?? ''}>
		{#if kind === 'textarea'}
			<Textarea
				value={typeof stored === 'string' ? stored : ''}
				placeholder={props.placeholder ?? undefined}
				required={props.required ?? false}
				disabled={streaming()}
				oninput={(e) => commit(e.currentTarget.value)}
			/>
		{:else if kind === 'select'}
			<Select
				value={typeof stored === 'string' ? stored : ''}
				required={props.required ?? false}
				disabled={streaming()}
				onchange={(e) => commit(e.currentTarget.value)}
			>
				<option value="" disabled>{props.placeholder ?? ''}</option>
				{#each props.options ?? [] as option}
					<option value={option}>{option}</option>
				{/each}
			</Select>
		{:else}
			<Input
				type={kind === 'number' ? 'number' : 'text'}
				value={stored ?? ''}
				placeholder={props.placeholder ?? undefined}
				required={props.required ?? false}
				disabled={streaming()}
				oninput={(e) => commit(typed(e.currentTarget.value))}
			/>
		{/if}
	</MuseField>
{/if}
