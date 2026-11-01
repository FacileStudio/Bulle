# OpenUI

Lets the assistant answer with UI instead of prose. The model writes OpenUI Lang inside a fenced block tagged `openui`. The client renders it with muse components.

## Pieces

| File | Role |
| --- | --- |
| `fence.ts` | `splitFences(text, { streaming })` splits text into markdown and openui segments. An unterminated fence gives `complete: false`. |
| `OpenUIBlock.svelte` | Renders one source string. Props: `source`, `streaming`, `onAction`. Falls back to a raw code block on parse errors or unknown components. |
| `library.ts` | The whitelist: Card, Stat, Table, List, Badge, Callout, Chart, Button, Form, Field. Nothing else renders. |
| `actions.ts` | `OpenUIAction = { type: 'button' \| 'form'; name; values; label }`. |
| `prompt.ts` | `OPENUI_PROMPT`, the system-prompt section. |

## Use in a chat reply

```svelte
{#each splitFences(text, { streaming }) as segment}
	{#if segment.type === 'markdown'}
		<Markdown text={segment.text} />
	{:else}
		<OpenUIBlock
			source={segment.source}
			streaming={streaming || !segment.complete}
			onAction={(action) => send(action)}
		/>
	{/if}
{/each}
```

Post the action as a structured user message, for example `{ "openui_action": { "type": "form", "name": "signup", "values": { ... } } }`, and show `action.label` as the visible text.

## System prompt

The prompt is generated from the library, so it cannot drift from the renderer.

```sh
bun run gen:openui-prompt
```

This writes `prompt.generated.txt`. Commit it. Rerun after any change to `library.ts` or `prompt-source.ts`.

The Go API cannot embed a file outside its module. Copy the file into `apps/api` (for example `apps/api/prompts/openui.txt`) in a build step and use `//go:embed`. Append the contents to the system prompt.
