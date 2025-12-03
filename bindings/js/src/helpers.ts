export function toInt(value: unknown): number {
  const num = Number(value ?? 0)
  if (!Number.isFinite(num)) return 0
  return Math.trunc(num)
}
