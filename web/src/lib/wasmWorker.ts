/**
 * Web Worker for running the Go WASM module off the main thread.
 * This keeps the UI responsive during intensive computations.
 */
/// <reference lib="webworker" />

import type { WasmAPI } from './wasmTypes';

// The globalThis object in the worker will have the Go object and wasm functions
declare const globalThis: Window & typeof global & WasmAPI & {
    Go: any;
};

let wasmReady = false;
const wasmReadyPromise = new Promise<void>(async (resolve) => {
    try {
        // In a module worker, importScripts is not available.
        // We fetch and evaluate the script manually.
        // Use the correct base path for GitHub Pages
        const basePath = import.meta.env.PROD ? '/exam_scheduler/' : '/';
        const wasmExecCode = await fetch(`${basePath}wasm_exec.js`).then(r => r.text());
        (0, eval)(wasmExecCode);

        const go = new globalThis.Go();

        const wasmModule = await WebAssembly.instantiateStreaming(
            fetch(`${basePath}main.wasm`),
            go.importObject
        );

        // Run the wasm instance. This is a blocking call, but it's in a worker.
        // It will resolve when the Go program finishes initializing and is ready to accept calls.
        go.run(wasmModule.instance);

        // Check if the functions are exposed
        if (typeof globalThis.runSchedule !== 'function') {
            throw new Error("WASM module did not expose 'runSchedule' function.");
        }

        wasmReady = true;
        resolve();
    } catch (error) {
        console.error("WASM initialization failed:", error);
        postMessage({ type: 'ERROR', error: `WASM initialization failed: ${error}` });
    }
});

self.onmessage = async (event: MessageEvent) => {
    const { type, data } = event.data;

    if (!wasmReady) {
        await wasmReadyPromise;
    }

    try {
        switch (type) {
            case 'GET_VERSION': {
                const versionJson = globalThis.version();
                postMessage({ type: 'VERSION_RESULT', data: JSON.parse(versionJson) });
                break;
            }
            case 'GENERATE_SCHEDULE': {
                const { regCSV, hallsCSV, paramsJSON } = data;
                // The wasm function returns a JSON string
                const resultJson = globalThis.runSchedule(regCSV, hallsCSV, paramsJSON);
                const result = JSON.parse(resultJson);

                if (result.success) {
                    postMessage({ type: 'RESULT', data: result });
                } else {
                    postMessage({ type: 'ERROR', error: result.error, data: result });
                }
                break;
            }
            case 'VERIFY_SCHEDULE': {
                const { regCSV, scheduleCSV } = data;
                const reportJson = globalThis.verify(regCSV, scheduleCSV);
                const report = JSON.parse(reportJson);
                postMessage({ type: 'VERIFY_RESULT', data: report });
                break;
            }
            default:
                throw new Error(`Unknown message type: ${type}`);
        }
    } catch (error: any) {
        console.error(`Error processing message type ${type}:`, error);
        postMessage({ type: 'ERROR', error: error.message || 'An unknown error occurred in the worker.' });
    }
};

// Signal that the worker is loaded and ready for messages
postMessage({ type: 'WORKER_READY' });
