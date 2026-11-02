export type SessionSummary = { id: string; title: string; channel?: string; updated_at: string };

export type StoredToolCall = {
	call_id: string;
	name: string;
	args?: unknown;
	status: 'running' | 'ok' | 'failed';
	output_preview?: string;
	duration_ms?: number;
};

export type StoredMessage = {
	seq: number;
	role: string;
	content: string;
	answer?: string;
	thinking?: string;
	tools?: StoredToolCall[];
};

export type SessionDetail = {
	session: SessionSummary;
	messages: StoredMessage[];
	active_run_id?: string;
};

export type ComputerStatus = 'absent' | 'creating' | 'running' | 'stopped' | 'error';

export type Computer = {
	status: ComputerStatus;
	error?: string;
	cpus: number;
	memory_mb: number;
	disk_gb: number;
	last_active_at: string;
};

export type Usage = Record<string, number>;

export type RunEvent =
	| { type: 'RUN_STARTED'; data: { run_id: string } }
	| { type: 'TEXT_MESSAGE_CONTENT'; data: { delta: string } }
	| { type: 'THINKING'; data: { delta: string } }
	| { type: 'TOOL_CALL_START'; data: { call_id: string; name: string; args: unknown } }
	| {
			type: 'TOOL_CALL_END';
			data: { call_id: string; ok: boolean; output_preview?: string; duration_ms?: number };
	  }
	| { type: 'INTERRUPT'; data: { call_id: string; tool: string; args: unknown } }
	| { type: 'RUN_FINISHED'; data: { stop?: string; usage?: Usage; cost_usd?: number } }
	| { type: 'RUN_ERROR'; data: { message: string } };

export type RunEventType = RunEvent['type'];

export const RUN_EVENT_TYPES: RunEventType[] = [
	'RUN_STARTED',
	'TEXT_MESSAGE_CONTENT',
	'THINKING',
	'TOOL_CALL_START',
	'TOOL_CALL_END',
	'INTERRUPT',
	'RUN_FINISHED',
	'RUN_ERROR'
];

export type SeqEvent = RunEvent & { seq: number };

export type StreamHandlers = {
	onEvent: (event: SeqEvent) => void;
	onOpen: () => void;
	onReconnecting: () => void;
	onFatal: (message: string) => void;
};

export type StreamHandle = { close: () => void };

export interface ChatApi {
	listSessions(): Promise<SessionSummary[]>;
	createSession(title?: string): Promise<{ id: string; title: string }>;
	getSession(id: string): Promise<SessionDetail>;
	deleteSession(id: string): Promise<void>;
	sendMessage(sessionId: string, text: string): Promise<{ run_id: string }>;
	stopRun(runId: string): Promise<void>;
	decide(runId: string, callId: string, approve: boolean): Promise<void>;
	openRunStream(runId: string, lastSeq: number, handlers: StreamHandlers): StreamHandle;
	getComputer(): Promise<Computer>;
	stopComputer(): Promise<void>;
}
