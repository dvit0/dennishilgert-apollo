import { join } from 'path'

const indexFile = join(__dirname, 'index.js')

const indexModule = require(indexFile)

let rawData = ''
process.stdin.on('data', (chunk) => {
  rawData += chunk
})

process.stdin.on('end', () => {
  let inputData
  try {
    inputData = JSON.parse(rawData)
  } catch(error) {
    console.log('error parsing JSON input', error)
    process.exit(1)
  }

  const { fnContext, fnData } = inputData

  if (typeof indexModule.handler === 'function') {
    indexModule.handler(fnContext, fnData)
      .then((result) => {
        console.log(result)
      })
      .catch((error) => {
        console.error('error in handler function', error)
      })
  } else {
    console.error('cannot find handler function in index file')
  }
})