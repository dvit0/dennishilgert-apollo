export async function handler(event) {
  console.log("Hello World!")

  console.log("The following message represents the json string of the event object passed to this function.")
  console.log(JSON.stringify(event, null, 2))

  return { "currentMillis": Date.now() }
}