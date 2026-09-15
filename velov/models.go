package velov

type Response struct {
	Values []Station `json:"values"`
}

type Station struct {
	Number     int        `json:"number"`
	Name       string     `json:"name"`
	MainStands MainStands `json:"main_stands"`
	Status     string     `json:"status"`
}

type MainStands struct {
	Availabilities Availabilities `json:"availabilities"`
}

type Availabilities struct {
	Bikes           int `json:"bikes"`
	ElectricalBikes int `json:"electricalBikes"`
	MechanicalBikes int `json:"mechanicalBikes"`
}
