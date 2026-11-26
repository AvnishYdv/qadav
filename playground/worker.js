// playground/worker.js - Qadav WebAssembly Multi-Threaded Worker
// Handles parallel execution of qadav_worker_step over shared linear memory

self.onmessage = async function(e) {
  const { wasmBytes, sharedMemory, workerId, numWorkers } = e.data;
  try {
    const importObject = {
      env: {
        memory: sharedMemory,
        qadav_emit_int: (val) => self.postMessage({ type: 'int', workerId, val }),
        qadav_emit_text: (ptr, len) => self.postMessage({ type: 'text', workerId, ptr, len })
      }
    };
    
    const module = await WebAssembly.compile(wasmBytes);
    const instance = await WebAssembly.instantiate(module, importObject);
    
    // Execute worker step partitioned by worker_id and num_workers
    const res = instance.exports.qadav_worker_step(workerId, numWorkers);
    self.postMessage({ type: 'done', workerId, result: Number(res) });
  } catch (err) {
    self.postMessage({ type: 'error', workerId, error: err.message });
  }
};
