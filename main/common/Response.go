package common

// Response è una struttura creata per memorizzare la risposta delle chiamate RPC
type Response struct {
	Value         string
	Result        bool
	ReceptionFIFO int // Timestamp per come il client dovrebbe leggere la risposta processata dal server
}

// SetValue imposta il valore della risposta passata in argomento nella struttura Response
func (response *Response) SetValue(value string) {
	response.Value = value
}

// GetValue restituisce il valore della risposta memorizzato nella struttura Response
func (response *Response) GetValue() string {
	return response.Value
}

// SetResult imposta il risultato della risposta passata in argomento nella struttura Response
// la risposta può essere true o false, per indicare se la chiamata RPC è andata a buon fine o meno
func (response *Response) SetResult(result bool) {
	response.Result = result
}

// GetResult restituisce il risultato della risposta memorizzato nella struttura Response
// la risposta può essere true o false, per indicare se la chiamata RPC è andata a buon fine o meno
func (response *Response) GetResult() bool {
	return response.Result
}

// SetReceptionFIFO imposta il timestamp di ricezione della risposta passata in argomento nella struttura Response
// il timestamp è utilizzato per indicare come il client dovrebbe leggere la risposta processata dal server
func (response *Response) SetReceptionFIFO(timestamp int) {
	response.ReceptionFIFO = timestamp
}

// GetReceptionFIFO restituisce il timestamp di ricezione della risposta memorizzato nella struttura Response
// il timestamp è utilizzato per indicare come il client dovrebbe leggere la risposta processata dal server
func (response *Response) GetReceptionFIFO() int {
	return response.ReceptionFIFO
}
