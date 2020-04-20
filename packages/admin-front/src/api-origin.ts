export const apiOrigin = (): string => {
  console.log(`process.env = ${JSON.stringify(process.env)}`)
  const REACT_APP_API_ORIGIN = process.env.REACT_APP_API_ORIGIN
  if (REACT_APP_API_ORIGIN === undefined) {
    throw new Error("API_ORIGIN must be defined")
  }
  return REACT_APP_API_ORIGIN
}
