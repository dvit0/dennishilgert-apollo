export async function handler(context, event) {
  console.log('Hello')
  console.log('World!')
  console.log('Third logged messages')

  return { "measurement": "this could be calculated" }
}