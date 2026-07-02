// launchstatus represents enum digmark.launcher_status
package launchstatus

type Enum string

func (e Enum) String() string {
	return string(e)
}

const (
	Completed  Enum = "Completed"
	Failed     Enum = "Failed"
	InProgress Enum = "In progress"
	Invalid    Enum = "Invalid"
	Queued     Enum = "Queued"
	Ready      Enum = "Ready"
	Unchecked  Enum = "Unchecked"
)

var (
	Editable = [...]Enum{Invalid, Ready, Unchecked}
)
