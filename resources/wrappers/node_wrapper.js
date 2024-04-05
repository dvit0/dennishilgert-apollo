import { fileURLToPath } from 'url'
import { join, dirname } from 'path'

const __dirname = dirname(fileURLToPath(import.meta.url))

let handlerFile = ''
let handlerFunctionName = ''

const origConsoleLog = console.log

function logJson(level, message) {
  const timestamp = Date.now()
  origConsoleLog(JSON.stringify({ timestamp, 'type': 'log', 'properties': { level, message } }))
}

function errorJson(error) {
  const timestamp = Date.now()
  origConsoleLog(JSON.stringify({ timestamp, 'type': 'error', 'properties': { 'code': error.code, 'message': error.message, 'cause': error.cause, 'stack': error.stack } }))
}

function resultJson(data) {
  const timestamp = Date.now()
  origConsoleLog(JSON.stringify({ timestamp, "type": "result", 'properties': { data } }))
}

console.log = (message) => logJson('info', message)
console.error = (message) => logJson('error', message)

process.stdin.setEncoding('utf8')

let initializationComplete = false
let buffer = ''

process.stdin.on('data', (chunk) => {
  buffer += chunk
  let boundary = buffer.indexOf('\n')
  while (boundary !== -1) {
    let data = buffer.substring(0, boundary)
    buffer = buffer.substring(boundary + 1)
    if (!initializationComplete) {
      // Process the initialization data
      let initData
      try {
        initData = JSON.parse(data)
        const handlerParts = initData.handler.split('.')
        handlerFile = handlerParts[0]
        handlerFunctionName = handlerParts[1]
        initializationComplete = true
      } catch (error) {
        errorJson(error)
        process.exit(1)
      }
    } else {
      processRequest(data)
    }
    boundary = buffer.indexOf('\n')
  }
});

async function processRequest(data) {
  if (!initializationComplete) {
    errorJson(new Error("Runtime not initialized"))
    return
  }
  let inputData
  try {
    inputData = JSON.parse(data)
    const { context, event } = inputData
    const indexFile = join(__dirname, `${handlerFile}.mjs`)

    const indexModule = await import(indexFile)
    const handlerFunction = indexModule[handlerFunctionName]

    if (typeof handlerFunction === 'function') {
      await handlerFunction(context, event)
        .then((result) => {
          resultJson(result)
        })
        .catch((error) => {
          errorJson(error)
        });
    } else {
      errorJson(new Error(`Specified handler function does not exist: ${handlerFunctionName}`))
    }
  } catch (error) {
    errorJson(error)
  }
}