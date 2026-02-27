package pushz

type PushState struct {
	File        string
	Hash        string   `json:"hash"`
	SyncedNodes []string `json:"synced_nodes"`
}

func TrackState() {

}
func NewPushState() {

}
