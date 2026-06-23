// runstatus represents enum process.run_status
package runstatus

type Enum string

func (e Enum) String() string {
	return string(e)
}

const (
	Cancelled   Enum = "Cancelled" // process cancelled before it started
	Completed   Enum = "Completed"
	Error       Enum = "Error"
	Interrupted Enum = "Interrupted" // process cancelled after it started
	Running     Enum = "Running"
	Waiting     Enum = "Waiting" // process waiting on a dependency
)
