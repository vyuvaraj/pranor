/**
 * @pranor/pulse-wasm — SharedWorker Multi-Tab Coordinator
 * Ensures single OPFS file handle access across multiple browser tabs via SharedWorker & BroadcastChannel.
 */

const channel = typeof BroadcastChannel !== 'undefined' ? new BroadcastChannel('pranor-pulse_events') : null;
const ports = new Set();

self.onconnect = (e) => {
  const port = e.ports[0];
  ports.add(port);

  port.onmessage = (event) => {
    const { type, topic, payload, offset } = event.data || {};

    if (type === 'ENQUEUE') {
      const logEntry = {
        offset: offset || Date.now(),
        topic: topic || 'default',
        payload: payload,
        timestamp: Date.now(),
        synced: false
      };

      // Broadcast event to all open browser tabs
      if (channel) {
        channel.postMessage({ type: 'EVENT_ENQUEUED', entry: logEntry });
      }

      // ACK back to requesting tab
      port.postMessage({ type: 'ENQUEUE_ACK', entry: logEntry });
    }
  };

  port.start();
};
