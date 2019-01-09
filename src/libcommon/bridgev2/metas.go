package bridgev2

// operation meta for connect/validate
type ConnectMeta struct {
    Secret string `json:"secret"`
    UUID   string `json:"uuid"` // this is client uuid
}

// operation meta for connect/validate
type ConnectResponseMeta struct {
    UUID        string `json:"uuid"` // this is server uuid
    New4Tracker bool `json:"new"`
}
