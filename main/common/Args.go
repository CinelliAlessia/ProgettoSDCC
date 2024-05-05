package common

// Args rappresenta gli argomenti delle chiamate RPC
type Args struct {
	Key         string
	Value       string
	SendingFIFO int    // Assunzione FIFO Ordering
	ClientId    string // Per identificare il client nella operazione RPC
}

/* ARGS STRUCT */

func NewArgs(timestamp int, key string, values ...string) Args {
	args := Args{}
	args.SetSendingFIFO(timestamp)
	args.SetKey(key)
	args.SetIDClient(GenerateUniqueID())

	if len(values) > 0 {
		args.SetValue(values[0])
	}
	return args
}

func (args *Args) SetKey(key string) {
	args.Key = key
}

func (args *Args) GetKey() string {
	return args.Key
}

func (args *Args) SetValue(value string) {
	args.Value = value
}

func (args *Args) GetValue() string {
	return args.Value
}

func (args *Args) SetSendingFIFO(timestamp int) {
	args.SendingFIFO = timestamp
}

func (args *Args) GetSendingFIFO() int {
	return args.SendingFIFO
}

func (args *Args) SetIDClient(id string) {
	args.ClientId = id
}

func (args *Args) GetIDClient() string {
	return args.ClientId
}
