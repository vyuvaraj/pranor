/**
 * @pranor/queue-wasm — Embedded Browser Event Broker TypeScript Definitions
 */

export interface LogEntry {
  offset: number;
  topic: string;
  payload: string;
  timestamp: number;
  synced: boolean;
}

export class PranorPulseEmbedded {
  constructor(options?: { opfsPath?: string; serverUrl?: string });

  /**
   * Enqueue a new event payload onto a topic persistent in OPFS
   */
  enqueue(topic: string, payload: string | object): Promise<LogEntry>;

  /**
   * Dequeue an offset range of events for a given topic
   */
  dequeue(topic: string, startOffset?: number, limit?: number): Promise<LogEntry[]>;

  /**
   * Manually trigger offline outbox sync to backend PranorPulse server
   */
  syncOutbox(serverUrl?: string): Promise<{ status: string }>;

  /**
   * Get unacknowledged offline events pending backend sync
   */
  getPendingSync(limit?: number): Promise<LogEntry[]>;
}

export class PranorPulseSharedWorkerCoordinator {
  constructor(workerUrl?: string);
  /**
   * Broadcast event to SharedWorker across all connected browser tabs
   */
  broadcast(topic: string, payload: string | object): void;
  /**
   * Subscribe to multi-tab BroadcastChannel events
   */
  onEvent(callback: (entry: LogEntry) => void): void;
}
