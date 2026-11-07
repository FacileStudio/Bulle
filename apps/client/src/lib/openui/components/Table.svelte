<script lang="ts">
	import { Table as MuseTable } from '@facile/muse';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	let {
		props
	}: ComponentRenderProps<{
		columns?: string[];
		rows?: (string | number | boolean | null)[][];
	}> = $props();

	const columns = $derived(props.columns ?? []);
	const rows = $derived(props.rows ?? []);
</script>

<MuseTable class="[&_tbody_tr]:border-t-0! [&_tbody_tr:nth-child(even)]:bg-fc-surface">
	<thead>
		<tr>
			{#each columns as column}
				<th scope="col">{column}</th>
			{/each}
		</tr>
	</thead>
	<tbody>
		{#each rows as row}
			<tr>
				{#each row as cell}
					<td>{cell ?? ''}</td>
				{/each}
			</tr>
		{/each}
	</tbody>
</MuseTable>
