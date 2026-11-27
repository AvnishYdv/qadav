// playground/worker.js - Qadav WebAssembly Multi-Threaded Worker
// Handles parallel execution of qadav_worker_step over shared linear memory

self.onmessage = async function(e) {
  const { wasmBytes, sharedMemory, workerId, numWorkers } = e.data;
  const start = performance.now();
  try {
    let instance = null;
    const importObject = {
      env: {
        memory: sharedMemory,
        qadav_emit_int: (val) => self.postMessage({ type: 'int', workerId, val }),
        qadav_emit_text: (ptr, len) => {
          const mem = (instance && instance.exports.memory) ? instance.exports.memory : sharedMemory;
          if (mem) {
            const arr = new Uint8Array(mem.buffer, ptr, len);
            const str = new TextDecoder().decode(arr);
            self.postMessage({ type: 'text', workerId, text: str });
          }
        }
      }
    };
    
    const module = await WebAssembly.compile(wasmBytes);
    instance = await WebAssembly.instantiate(module, importObject);
    
    // Execute worker step partitioned by worker_id and num_workers
    const res = instance.exports.qadav_worker_step(workerId, numWorkers);
    const duration = (performance.now() - start).toFixed(2);
    self.postMessage({
      type: 'done',
      workerId,
      result: Number(res),
      durationMs: duration,
      vectorsProcessed: 64
    });
  } catch (err) {
    self.postMessage({ type: 'error', workerId, error: err.message });
  }
};
