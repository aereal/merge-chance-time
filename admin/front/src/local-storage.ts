export type Presented<T = any> = T extends undefined ? never : T extends null ? never : T

export const setItem = (key: string, value: any): void => {
  const serialized = JSON.stringify(value)
  window.localStorage.setItem(key, serialized)
}

export const getItem = (key: string): Presented | null => {
  const got = window.localStorage.getItem(key)
  if (got === null) {
    return null
  }
  return JSON.parse(got)
}
