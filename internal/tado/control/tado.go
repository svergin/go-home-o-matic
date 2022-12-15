package control

type TadoController interface {
}

type tadoControllerImpl struct {
}

var _ TadoController = &tadoControllerImpl{}
