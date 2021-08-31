package globals

const ID_HEADER = "Persistence-Stream-Id"
const INITIAL_OR_REATTACH_HEADER = "Persistence-Stream-New-Or-Reattach"
const COORD_OR_STREAM_HEADER = "Persistence-Stream-Coord-Or-Stream"

const INITIAL = "initial"
const REATTACH = "reattach"
const COORD = "coord"
const STREAM = "stream"

type JsonResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
