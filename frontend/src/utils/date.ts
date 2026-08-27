import dayjs from 'dayjs';

/** Format datetime as YYYY-MM-DD HH:mm:ss */
export function formatDateTime(dateStr?: string | null): string {
  if (!dateStr) return '-';
  const date = dayjs(dateStr);
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm:ss') : '-';
}
