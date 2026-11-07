<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Icon, IconButton, Input } from '@facile/muse';
	import { getCaretCoordinates, type CaretCoordinates } from '$lib/caret';
	import { CHAT_ICONS } from '$lib/icons';
	import { NAME_MAX_LENGTH, randomName } from '$lib/onboarding/profile';

	let {
		value = $bindable(''),
		id,
		label = t('onboarding.name.label'),
		hideLabel = false,
		oncaret
	}: {
		value?: string;
		id: string;
		label?: string;
		hideLabel?: boolean;
		oncaret?: (coords: CaretCoordinates | null) => void;
	} = $props();

	function report(event: Event) {
		const field = event.currentTarget as HTMLInputElement;
		oncaret?.(getCaretCoordinates(field, field.selectionStart ?? undefined));
	}
</script>

<div class="flex flex-col gap-1.5 text-left">
	<label for={id} class={hideLabel ? 'sr-only' : 'text-fc-sm font-medium text-fc-fg'}>
		{label}
	</label>
	<div class="flex items-center gap-2">
		<Input
			{id}
			bind:value
			maxlength={NAME_MAX_LENGTH}
			autocomplete="off"
			placeholder={t('onboarding.name.placeholder')}
			oninput={report}
			onkeyup={report}
			onclick={report}
			onfocus={report}
			onblur={() => oncaret?.(null)}
		/>
		<IconButton aria-label={t('onboarding.name.random')} onclick={() => (value = randomName(value))}>
			<Icon icon={CHAT_ICONS.dice} size={18} />
		</IconButton>
	</div>
</div>
