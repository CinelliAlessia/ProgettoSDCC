package common

// Response è una struttura creata per memorizzare la risposta delle chiamate RPC
type Response struct {
	Value         string
	Result        bool
	Done          chan bool // lo setto ma non viene mai letto da nessuno
	ServerId      int       // non dovrebbe servire, non lo uso
	ReceptionFIFO int
}

/* RESPONDE STRUCT */

func (response *Response) SetValue(value string) {
	response.Value = value
}

func (response *Response) GetValue() string {
	return response.Value
}

func (response *Response) SetResult(result bool) {
	response.Result = result
}

func (response *Response) GetResult() bool {
	return response.Result
}

func (response *Response) SetDone(done bool) {
	response.Done <- done
}

func (response *Response) GetDone() chan bool {
	return response.Done
}

func (response *Response) SetServerId(serverId int) {
	response.ServerId = serverId
}

func (response *Response) GetServerId() int {
	return response.ServerId
}

func (response *Response) SetReceptionFIFO(timestamp int) {
	response.ReceptionFIFO = timestamp
}

func (response *Response) GetReceptionFIFO() int {
	return response.ReceptionFIFO
}
