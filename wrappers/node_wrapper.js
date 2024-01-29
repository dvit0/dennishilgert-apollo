import { fileURLToPath } from 'url'
import { join, dirname } from 'path'

// Helper function to get __dirname equivalent in ES module
const __dirname = dirname(fileURLToPath(import.meta.url))

const origConsoleLog = console.log

function logJson(level, message) {
  const timestamp = Date.now()
  origConsoleLog(JSON.stringify({ timestamp, 'type': 'log', 'properties': { level, message } }))
}

function errorJson(error) {
  const timestamp = Date.now()
  origConsoleLog('error', JSON.stringify({ timestamp, 'type': 'error', 'properties': { 'code': error.code, 'message': error.message, 'cause': error.cause, 'stack': error.stack } }))
}

function resultJson(data) {
  const timestamp = Date.now()
  origConsoleLog(JSON.stringify({ timestamp, "type": "result", 'properties': { data } }))
}

// Override console.log and console.error
console.log = (message) => logJson('info', message)
console.error = (message) => logJson('error', message)

let rawData = ''
process.stdin.on('data', (chunk) => {
  rawData += chunk
});

process.stdin.on('end', async () => {
  let inputData
  try {
    inputData = JSON.parse(rawData)
  } catch (error) {
    errorJson(error)
    process.exit(1)
  }

  const { context, event } = inputData

  const handlerParts = context.runtimeHandler.split('.')
  const index = handlerParts[0]
  const handler = handlerParts[1]
  const indexFile = join(__dirname, index + '.mjs')
  
  const indexModule = await import(indexFile)
  const handlerFunction = indexModule[handler]

  if (typeof handlerFunction === 'function') {
    await handlerFunction(context, event)
      .then((result) => {
        resultJson(result)
      })
      .catch((error) => {
        errorJson(error)
      })
  } else {
    errorJson(new Error('specified handler function does not exist: ' + context.Handler))
  }
})