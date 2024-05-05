package common

// Args rappresenta gli argomenti delle chiamate RPC
type Args struct {
	Key             string
	Value           string
	TimestampClient int    // Chiedere ulteriori spiegazioni
	ClientId        string // Per identificare il client nella operazione RPC
}

/* ARGS STRUCT */

func NewArgs(timestamp int, key string, values ...string) Args {
	args := Args{}
	args.SetTimestamp(timestamp)
	args.SetKey(key)
	args.setIDClient(GenerateUniqueID())

	if len(values) > 0 {
		args.SetValue(values[0])
	}
	return args
}

func (args *Args) setIDClient(id string) {
	args.ClientId = id
}

func (args *Args) GetIDClient() string {
	return args.ClientId
}

func (args *Args) SetKey(key string) {
	args.Key = key
}

func (args *Args) SetValue(value string) {
	args.Value = value
}

func (args *Args) SetTimestamp(timestamp int) {
	args.TimestampClient = timestamp
}

func (args *Args) GetKey() string {
	return args.Key
}

func (args *Args) GetValue() string {
	return args.Value
}

func (args *Args) GetTimestamp() int {
	return args.TimestampClient
}

func (args *Args) SetIDClient(client string) {
	args.ClientId = client
}
