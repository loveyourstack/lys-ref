
export function statusColor(status: string): string {
  switch (status) {
    case 'Completed':
      return 'success'
    case 'Failed':
      return 'error'
    case 'In progress':
      return '#F4C430'
    case 'Invalid':
      return 'orange'
    case 'Queued':
      return 'blue'
    default:
      return 'grey'
  }
}