package appenv

type Enum string

func (e Enum) String() string {
	return string(e)
}

const (
	Dev  Enum = "dev"
	Prod Enum = "prod"
)
