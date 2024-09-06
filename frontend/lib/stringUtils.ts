export const truncate = (str: string, length: number) =>
  str.substring(0, Math.min(str.length, length)) +
  (str.length > length ? '...' : '')
