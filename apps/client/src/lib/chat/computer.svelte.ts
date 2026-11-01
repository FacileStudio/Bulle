import { api, type Computer } from '$lib/api';

const FAST_MS = 3000;
const SLOW_MS = 30000;

class ComputerStore {
	value = $state.raw<Computer | null>(null);
	error = $state<string | null>(null);
	busy = $state(false);
	private timer: ReturnType<typeof setTimeout> | undefined;
	private active = false;

	async refresh(): Promise<void> {
		try {
			this.value = await api().getComputer();
			this.error = null;
		} catch (err) {
			this.error = err instanceof Error ? err.message : 'Could not read the computer status';
		}
	}

	private schedule(): void {
		if (!this.active) return;
		clearTimeout(this.timer);
		const delay = this.value?.status === 'creating' ? FAST_MS : SLOW_MS;
		this.timer = setTimeout(async () => {
			await this.refresh();
			this.schedule();
		}, delay);
	}

	async watch(): Promise<void> {
		this.active = true;
		await this.refresh();
		this.schedule();
	}

	unwatch(): void {
		this.active = false;
		clearTimeout(this.timer);
	}

	async start(): Promise<void> {
		await this.act(() => api().startComputer());
	}

	async stop(): Promise<void> {
		await this.act(() => api().stopComputer());
	}

	private async act(call: () => Promise<void>): Promise<void> {
		this.busy = true;
		try {
			await call();
			this.error = null;
		} catch (err) {
			this.error = err instanceof Error ? err.message : 'The request failed';
		}
		await this.refresh();
		this.busy = false;
		this.schedule();
	}
}

export const computer = new ComputerStore();
