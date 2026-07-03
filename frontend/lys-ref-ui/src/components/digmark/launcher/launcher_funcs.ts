
export function statusColor(status: string): string {
  switch (status) {
    case 'Completed':
      return 'success'
    case 'Failed':
      return 'error'
    case 'Processing':
      return '#F4C430'
    case 'Invalid':
      return 'orange'
    case 'Preparing':
      return '#F4C430'
    case 'Queued':
      return 'blue'
    default:
      return 'grey'
  }
}