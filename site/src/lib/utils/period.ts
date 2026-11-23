const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

function formatMonth(value: string): string {
  const [year, month] = value.slice(0, 7).split('-').map(Number);
  return `${months[month - 1]} ${year}`;
}

function monthIndex(value: string): number {
  const [year, month] = value.split('-').map(Number);
  return year * 12 + (month - 1);
}

export function formatPeriod(date: string, updated: string, published: boolean): string {
  const now = new Date();
  const currentMonth = now.getUTCFullYear() * 12 + now.getUTCMonth();
  const completed = published && currentMonth - monthIndex(updated) > 3;
  const end = completed ? formatMonth(updated) : 'now';

  return `${formatMonth(date)} - ${end}`;
}
