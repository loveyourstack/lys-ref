// launchstatus represents enum digmark.launcher_status
package launchstatus

type Enum string

func (e Enum) String() string {
	return string(e)
}

const (
	Completed  Enum = "Completed"
	Failed     Enum = "Failed"
	Processing Enum = "Processing"
	Invalid    Enum = "Invalid"
	Preparing  Enum = "Preparing"
	Queued     Enum = "Queued"
	Ready      Enum = "Ready"
	Unprepared Enum = "Unprepared"
)

var (
	Editable = [...]Enum{Invalid, Ready, Unprepared}
)
