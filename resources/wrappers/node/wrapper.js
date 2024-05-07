import { join } from 'path';

const origConsoleLog = console.log;

let initializationComplete = false;
let handlerFile = '';
let handlerFunctionName = '';

console.log = (message) => logInfo(message);
console.error = (message) => logError(message);

function logJson(type, properties) {
  origConsoleLog(JSON.stringify({
    timestamp: Date.now(),
    type,
    properties
  }));
}

function logInfo(message) {
  logJson('log', { level: 'info', message });
}

function logError(error) {
  logJson('error', {
    code: error.code || 520, // 520 Unknown Error
    message: error.message,
    cause: error.cause || 'UNKNOWN',
    stack: error.stack || 'UNKNOWN'
  });
}

function logResult(data) {
  logJson('result', { data });
}

function logDone() {
  logJson('done', {});
}

async function processRequest(data) {
  if (!initializationComplete) {
    logError(new Error("Runtime not initialized"));
    logDone();
    return;
  }

  try {
    const inputData = JSON.parse(data);
    const { context, event } = inputData;
    const indexFile = join('/workspace/code', `${handlerFile}.mjs`);

    const indexModule = await import(indexFile);
    const handlerFunction = indexModule[handlerFunctionName];

    if (typeof handlerFunction !== 'function') {
      throw new Error(`Specified handler function does not exist: ${handlerFunctionName}`);
    }

    await handlerFunction(context, event)
      .then(logResult)
      .catch(logError)
      .finally(logDone);
  } catch (error) {
    logError(error);
    logDone();
  }
}

process.stdin.setEncoding('utf8');

// Declare the buffer outside of the event listener because the data can be sent in multiple chunks.
let buffer = '';
process.stdin.on('data', (chunk) => {
  buffer += chunk;
  // Check if the end of the data has been reached by looking for a newline character.
  let boundary = buffer.indexOf('\n');
  while (boundary !== -1) {
    const data = buffer.substring(0, boundary);
    buffer = buffer.substring(boundary + 1);
    if (!initializationComplete) {
      try {
        const initData = JSON.parse(data);
        [handlerFile, handlerFunctionName] = initData.handler.split('.');
        initializationComplete = true;
      } catch (error) {
        logError(error);
        process.exit(1);
      }
    } else {
      processRequest(data).catch(logError);
    }
    boundary = buffer.indexOf('\n');
  }
});