import { createLibrary, defineComponent } from '@openuidev/svelte-lang';
import { z } from 'zod';
import BadgeView from './components/Badge.svelte';
import ButtonView from './components/Button.svelte';
import CalloutView from './components/Callout.svelte';
import CardView from './components/Card.svelte';
import ChartView from './components/Chart.svelte';
import DividerView from './components/Divider.svelte';
import FieldView from './components/Field.svelte';
import FormView from './components/Form.svelte';
import ListView from './components/List.svelte';
import StatView from './components/Stat.svelte';
import TableView from './components/Table.svelte';
import TextView from './components/Text.svelte';

const cell = z.union([z.string(), z.number(), z.boolean()]);

export const Stat = defineComponent({
	name: 'Stat',
	description: 'One headline metric with an optional change note, for example "+12% vs last week".',
	props: z.object({
		label: z.string(),
		value: z.union([z.string(), z.number()]),
		delta: z.string().optional()
	}),
	component: StatView
});

export const Table = defineComponent({
	name: 'Table',
	description: 'A data table. Each row is an array of cells in column order.',
	props: z.object({
		columns: z.array(z.string()),
		rows: z.array(z.array(cell))
	}),
	component: TableView
});

export const List = defineComponent({
	name: 'List',
	description: 'A bulleted or numbered list of short text items.',
	props: z.object({
		items: z.array(z.string()),
		ordered: z.boolean().optional()
	}),
	component: ListView
});

export const Badge = defineComponent({
	name: 'Badge',
	description: 'A small status pill.',
	props: z.object({
		text: z.string(),
		tone: z.enum(['neutral', 'accent', 'info', 'success', 'warning', 'danger']).optional()
	}),
	component: BadgeView
});

export const Callout = defineComponent({
	name: 'Callout',
	description: 'A highlighted message with an optional title. Use for warnings and key takeaways.',
	props: z.object({
		text: z.string(),
		tone: z.enum(['neutral', 'info', 'success', 'warning', 'danger']).optional(),
		title: z.string().optional()
	}),
	component: CalloutView
});

export const Chart = defineComponent({
	name: 'Chart',
	description:
		'A bar or line chart. labels are the x categories; every series has one number per label.',
	props: z.object({
		kind: z.enum(['bar', 'line']),
		labels: z.array(z.string()),
		series: z.array(z.object({ name: z.string(), data: z.array(z.number()) })),
		title: z.string().optional()
	}),
	component: ChartView
});

export const Button = defineComponent({
	name: 'Button',
	description:
		'A button. When clicked the user sends back the action name so you can continue the conversation. name is a short snake_case action id.',
	props: z.object({
		label: z.string(),
		name: z.string(),
		variant: z.enum(['primary', 'outline', 'ghost', 'danger']).optional()
	}),
	component: ButtonView
});

export const Field = defineComponent({
	name: 'Field',
	description:
		'One input of a Form. kind defaults to text. options is required for select. name is a short snake_case key.',
	props: z.object({
		name: z.string(),
		label: z.string(),
		kind: z.enum(['text', 'number', 'textarea', 'select', 'checkbox']).optional(),
		options: z.array(z.string()).optional(),
		placeholder: z.string().optional(),
		required: z.boolean().optional()
	}),
	component: FieldView
});

export const Form = defineComponent({
	name: 'Form',
	description:
		'A form with a submit button. On submit the user sends back the form name and all field values. name is a short snake_case form id.',
	props: z.object({
		name: z.string(),
		submitLabel: z.string(),
		fields: z.array(Field.ref)
	}),
	component: FormView
});

export const Text = defineComponent({
	name: 'Text',
	description: 'A block of prose text with optional tone styling.',
	props: z.object({
		content: z.string(),
		tone: z.enum(['neutral', 'muted', 'accent', 'success', 'warning', 'danger']).optional()
	}),
	component: TextView
});

export const Divider = defineComponent({
	name: 'Divider',
	description: 'Extra space between two sections of a Card.',
	props: z.object({}),
	component: DividerView
});

export const Card = defineComponent({
	name: 'Card',
	description:
		'The root container. Always wrap the whole answer in one Card. It cannot be nested inside another Card.',
	props: z.object({
		children: z.array(
			z.union([
				Stat.ref,
				Table.ref,
				List.ref,
				Badge.ref,
				Callout.ref,
				Chart.ref,
				Button.ref,
				Form.ref,
				Text.ref,
				Divider.ref
			])
		),
		title: z.string().optional(),
		description: z.string().optional()
	}),
	component: CardView
});

export const library = createLibrary({
	components: [Card, Stat, Table, List, Badge, Callout, Chart, Button, Form, Field, Text, Divider],
	root: 'Card'
});

export const COMPONENT_NAMES = Object.keys(library.components);
