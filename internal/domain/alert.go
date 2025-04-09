package domain


type Alert struct {
	NumeroSerie        string `json:"numeroSerie"`
	Sensor             string `json:"sensor"`
	FechaActivacion    string `json:"fecha_activacion"`
	FechaDesactivacion string `json:"fecha_desactivacion"`
	Estado             int    `json:"estado"`
}