let counter = 0;

function prefersDark(): boolean {
	const root = document.documentElement.classList;
	if (root.contains('dark')) return true;
	if (root.contains('light')) return false;
	return window.matchMedia('(prefers-color-scheme: dark)').matches;
}

/**
 * Loads mermaid on first use and renders one diagram to SVG. `strict` keeps scripts, click
 * handlers and raw HTML labels out of the output, which mermaid then sanitises itself.
 */
export async function renderMermaid(code: string): Promise<string> {
	const { default: mermaid } = await import('mermaid');
	mermaid.initialize({
		startOnLoad: false,
		securityLevel: 'strict',
		theme: prefersDark() ? 'dark' : 'default',
		fontFamily: 'inherit'
	});
	counter += 1;
	const { svg } = await mermaid.render(`bulle-mermaid-${counter}`, code);
	return svg;
}
